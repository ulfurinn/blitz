package blizzard

import (
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"bitbucket.org/ulfurinn/blitz"
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
	announceCb       func(*Process, bool)
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

func (i *Process) Exec() error {
	i.state = "booting"
	i.inspect()
	args := []string{"-tag", i.tag}
	args = append(args, i.group.exe.args()...)
	i.cmd = exec.Command(i.group.exe.executable(), args...)
	err := i.cmd.Start()
	if err != nil {
		if i.announceCb != nil {
			i.announceCb(i, false)
		}
	}
	return err
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

func (i *Process) handleAnnounced(cmd *blitz.AnnounceCommand, w *WorkerConnection) {
	i.makeRevProxy()
	i.tag = "" // not needed anymore
	i.connection = w
	i.network = cmd.Network
	i.Address = cmd.Address
	i.state = "ready"
	i.inspect()
	if i.announceCb != nil {
		i.announceCb(i, true)
	}
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
