package blizzard

import (
	"net/http"
	"net/http/httputil"
	"os/exec"
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
	return i.cmd.Start()
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

func (i *Process) handleAnnounced(cmd blitz.AnnounceCommand, w *WorkerConnection) {
	i.makeRevProxy()
	i.tag = "" // not needed anymore
	i.connection = w
	i.network = cmd.Network
	i.Address = cmd.Address
	i.state = "ready"
	i.inspect()
}

func (i *Process) handleShutdown() {
	if i.cmd != nil {
		log("[proc %p] shutting down process %d\n", i, i.cmd.Process.Pid)
		i.state = "shutting-down"
		i.inspect()
		i.cmd.Process.Signal(syscall.SIGTERM)
	}
}

func (i *Process) handleCleanupProcess() {
	if i.cmd != nil {
		log("[proc %p] collecting process %d\n", i, i.cmd.Process.Pid)
		i.state = "collecting"
		i.inspect()
		i.cmd.Wait()
		i.inspectDispose()
		i.cmd = nil
	}
}
