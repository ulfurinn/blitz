package blizzard

import (
	"net/http"
	"net/http/httputil"
	"os/exec"
	"syscall"

	"bitbucket.org/ulfurinn/blitz"
)

type Process struct {
	Exe              *Executable
	connection       *WorkerConnection
	id               string
	Pid              int
	Patch            int64
	network, Address string
	proxy            http.Handler
	Requests         int64
	TotalRequests    uint64
	Written          uint64
	Obsolete         bool
	dead             bool
	cmd              *exec.Cmd
}

type ProcessSet map[*Process]struct{}

func (i *Process) makeRevProxy() {
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

func (i *Process) Shutdown() {
	if i.dead {
		return
	}
	//fmt.Printf("releasing instance %v\n", *i)
	//i.connection.conn.Close()
	syscall.Kill(i.Pid, syscall.SIGINT)
	i.cmd.Wait()
}
