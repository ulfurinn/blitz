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

func (i *Process) Exec() error {
	i.cmd = exec.Command(i.group.exe.Exe, "--tag", i.tag)
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

func (i *Process) handleShutdown() {
	if i.cmd != nil {
		log("[proc %p] shutting down process %d\n", i, i.cmd.Process.Pid)
		i.cmd.Process.Signal(syscall.SIGTERM)
	}
}

func (i *Process) handleCleanupProcess() {
	if i.cmd != nil {
		log("[proc %p] collecting process %d\n", i, i.cmd.Process.Pid)
		i.cmd.Wait()
		i.cmd = nil
	}
}
