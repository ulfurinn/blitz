package blizzard

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"bytes"

	"bitbucket.org/ulfurinn/blitz"
	"bitbucket.org/ulfurinn/gen_proc"
)

type ProcState int

const (
	ProcInit ProcState = iota
	ProcBooting
	ProcReady
	ProcShuttingDown
	ProcCollecting
)

type Process struct {
	*ProcessGen      `gen_proc:"gen_server"`
	server           *Blizzard
	state            string
	group            *ProcGroup
	connection       *WorkerConnection
	tag              string
	network, Address string
	proxy            http.Handler
	Requests         int64
	TotalRequests    uint64
	Written          uint64
	Obsolete         bool
	cmd              *exec.Cmd
	busyWg           sync.WaitGroup
}

type ProcessSet map[*ProcGroup]struct{}

type serveRequest struct {
	resp   http.ResponseWriter
	req    *http.Request
	served chan struct{}
}

func (i *Process) inspect() {
	i.server.inspect(ProcInspect(i))
}
func (i *Process) inspectDispose() {
	i.server.inspect(ProcInspectDispose(i))
}

func (i *Process) handleExec() (gen_proc.Deferred, *blitz.AnnounceCommand, error) {
	i.state = "booting"
	i.inspect()
	args := []string{"-tag", i.tag}
	args = append(args, i.group.exe.args()...)
	i.cmd = exec.Command(i.group.exe.executable(), args...)

	ok := make(chan struct {
		announce *blitz.AnnounceCommand
		worker   *WorkerConnection
	}, 1)

	procout, _ := i.cmd.StdoutPipe()
	procerr, _ := i.cmd.StderrPipe()

	i.server.AddTagCallback(i.tag, func(cmd interface{}, w *WorkerConnection) {
		//log("[proc %p] received announce\n", i)
		ok <- struct {
			announce *blitz.AnnounceCommand
			worker   *WorkerConnection
		}{cmd.(*blitz.AnnounceCommand), w}
	})

	err := i.cmd.Start()
	if err != nil {
		log("[proc %p] boot failed: %v\n", i, err)
		i.server.RemoveTagCallback(i.tag)
		return false, nil, err
	}

	return i.deferExec(func(ret func(*blitz.AnnounceCommand, error)) {
		died := make(chan struct{}, 1)
		var outlog bytes.Buffer
		var errlog bytes.Buffer
		go func() {
			go io.Copy(&outlog, procout)
			go io.Copy(&errlog, procerr)
			err := i.cmd.Wait()
			if err != nil {
				log("[proc %p] %v\n", i, err)
			}
			died <- struct{}{}
		}()

		select {
		case announced := <-ok:
			if err := procout.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			if err := procerr.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			log("[proc %p] announced on connection %p\n", i, announced.worker)
			announced.worker.monitor = i.group
			i.makeRevProxy()
			i.tag = "" // not needed anymore
			i.connection = announced.worker
			i.network = announced.announce.Network
			i.Address = announced.announce.Address
			i.state = "ready"
			i.inspect()
			ret(announced.announce, nil)
		case <-died:
			log("[proc %p] died during boot\n", i)
			ret(nil, fmt.Errorf("process died unexpectedly\nstdout:\n%s\nstderr:\n%s", outlog.Bytes(), errlog.Bytes()))
		}
	})
}

func (i *Process) makeRevProxy() {
	if i.proxy != nil {
		return
	}
	i.proxy = &httputil.ReverseProxy{
		Transport: blitz.UnixTransport,
		Director: func(newreq *http.Request) {
			newreq.URL.Scheme = "http"
			newreq.URL.Host = i.Address
			newreq.URL.Path = newreq.Header.Get("X-Blitz-Path")
		},
	}
}

func (i *Process) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	i.proxy.ServeHTTP(resp, req)
}

func (i *Process) handleShutdown(kill bool) {
	if i.cmd != nil {
		var sig os.Signal
		if kill {
			sig = syscall.SIGKILL
		} else {
			sig = syscall.SIGTERM
		}
		log("[proc %p] shutting down process %d using %v\n", i, i.cmd.Process.Pid, sig)
		i.state = "shutting-down"
		i.inspect()
		i.cmd.Process.Signal(sig)
	}
}

// func (i *Process) handleShutdownSync() {
// 	if i.cmd != nil {
// 		i.handleShutdown()
// 		i.handleCleanupProcess()
// 	}
// }

func (i *Process) handleCleanupProcess() {
	if i.cmd != nil {
		log("[proc %p] collecting process %d\n", i, i.cmd.Process.Pid)
		i.state = "collecting"
		i.inspect()
		i.cmd.Wait()
		err := os.Remove(i.Address) // SIGKILL won't leave a change to clean this up
		if err != nil && !os.IsNotExist(err) {
			log("[proc %p] error cleaning up domain socket %v\n", i, i.Address)
		}
		i.inspectDispose()
		i.cmd = nil
	}
}
