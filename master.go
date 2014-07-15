package blitz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

type PathSpec struct {
	Path    string `json:"path"`
	Version int    `json:"version"`
}

type Command struct {
	Type    string     `json:"type"`
	Tag     string     `json:"tag"`
	PID     int        `json:"pid"`
	ProcID  string     `json:"procid"`
	Patch   int64      `json:"patch"`
	Paths   []PathSpec `json:"paths"`
	Binary  string     `json:"binary"`
	Network string     `json:"network"`
	Address string     `json:"address"`
}

type Response struct {
	Error string `json:"error"`
}

type Master struct {
	execs     []*Executable
	procs     []*Instance
	routers   map[int]*Router
	routeLock *sync.RWMutex
	cmdCh     chan masterRequest
	server    *http.Server
}

type Executable struct {
	exe string
}

type Instance struct {
	exe              *Executable
	connection       *WorkerConnection
	id               string
	pid              int
	patch            int64
	network, address string
	proxy            http.Handler
	requests         int64
}

type masterRequest struct {
	cmd  Command
	conn *WorkerConnection
	ret  chan masterResponse
}

type masterResponse struct {
	err error
}

func randstr(n int64) string {
	b := &bytes.Buffer{}
	io.CopyN(b, NewRand(), n)
	return string(b.Bytes())
}

func NewMaster() *Master {
	return &Master{
		routers:   make(map[int]*Router),
		routeLock: &sync.RWMutex{},
		cmdCh:     make(chan masterRequest),
	}
}

func (m *Master) Run() {
	os.MkdirAll("blitz", os.ModeDir|0775)
	os.MkdirAll("blitz/deploy", os.ModeDir|0775)
	os.MkdirAll("blitz/deploy-old", os.ModeDir|0775)
	listener, err := net.Listen("unix", "blitz/ctl")
	if err != nil {
		fatal(err)
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go func() {
		<-ch
		err := listener.Close()
		if err != nil {
			fatal(err)
		}
		os.Exit(0)
	}()
	go m.HTTP()
	go m.Loop()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fatal(err)
		}
		worker := &WorkerConnection{conn: conn, master: m}
		go worker.Run()
	}

}

func (m *Master) HTTP() {
	m.server = &http.Server{
		Addr:    ":8080",
		Handler: m,
	}
	err := m.server.ListenAndServe()
	if err != nil {
		fatal(err)
	}
}

func (m *Master) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	path := req.URL.Path[1:]
	switch path {
	case "_blitz":

	case "_blitzws":
	default:
		m.routeLock.RLock()
		versionRouter, ok := m.routers[1]
		if !ok {
			resp.WriteHeader(404)
			return
		}
		h := versionRouter.Route(strings.Split(path, "/"))
		// do this here so that the collector in announce will see it as busy
		if h != nil {
			atomic.AddInt64(&h.requests, 1)
		}
		m.routeLock.RUnlock()
		if h == nil {
			resp.WriteHeader(404)
			return
		}
		defer atomic.AddInt64(&h.requests, -1)
		h.ServeHTTP(resp, req)
	}
}

func (i *Instance) makeRevProxy() {
	i.proxy = &httputil.ReverseProxy{
		Transport: unixTransport,
		Director: func(newreq *http.Request) {
			newreq.URL.Scheme = "http"
			newreq.URL.Host = i.address
		},
	}
}

func (i *Instance) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	i.proxy.ServeHTTP(resp, req)
}

func (m *Master) Loop() {
	for {
		select {
		case cmd := <-m.cmdCh:
			fmt.Fprintln(os.Stderr, cmd.cmd)
			switch cmd.cmd.Type {
			case "announce":
				m.Announce(cmd.cmd, cmd.conn)
				cmd.ret <- masterResponse{}
			case "deploy":
				err := m.Deploy(cmd.cmd.Binary)
				cmd.ret <- masterResponse{err: err}
			default:
				cmd.ret <- masterResponse{err: fmt.Errorf("unknown command %s", cmd.cmd.Type)}
			}
		}
	}
}

func (m *Master) Announce(cmd Command, c *WorkerConnection) {
	var proc *Instance
	for _, p := range m.procs {
		if p.id == cmd.ProcID {
			proc = p
			break
		}
	}
	if proc == nil {
		return
	}
	proc.makeRevProxy()
	proc.id = "" // not needed anymore
	proc.connection = c
	proc.pid = cmd.PID
	proc.patch = cmd.Patch
	proc.network = cmd.Network
	proc.address = cmd.Address
	m.routeLock.Lock()
	defer m.routeLock.Unlock()
	m.Mount(cmd.Paths, proc)
	// collect unused ones here
}

func (m *Master) Mount(paths []PathSpec, proc *Instance) {
	for _, path := range paths {
		if len(path.Path) > 0 {
			if path.Path[0] == '/' {
				path.Path = path.Path[1:]
			}
		}
		split := strings.Split(path.Path, "/")
		router, ok := m.routers[path.Version]
		if !ok {
			router = NewRouter()
			m.routers[path.Version] = router
		}
		router.Mount(split, proc)
	}
}

func (m *Master) Unmount(proc *Instance) {

}

func (m *Master) Deploy(exe string) error {
	components := strings.Split(exe, string(os.PathSeparator))
	newname := fmt.Sprintf("blitz/deploy/%s.blitz%d", components[len(components)-1], time.Now().Unix())
	origin, err := os.Open(exe)
	if err != nil {
		return err
	}
	newfile, err := os.Create(newname)
	if err != nil {
		return err
	}
	err = os.Chmod(newname, 0775)
	if err != nil {
		return err
	}
	_, err = io.Copy(newfile, origin)
	if err != nil {
		return err
	}

	return m.BootDeployed(newname)
}

func (m *Master) BootDeployed(exe string) error {
	id := randstr(32)
	e := &Executable{exe: exe}
	i := &Instance{exe: e, id: id}
	m.execs = append(m.execs, e)
	m.procs = append(m.procs, i)
	cmd := exec.Command(exe, "--blitz-proc-id", id)
	err := cmd.Start()
	return err
}

func (m *Master) connectionClosed(w *WorkerConnection) {
	var proc *Instance
	var index int
	for i, p := range m.procs {
		if p.connection == w {
			proc = p
			index = i
			break
		}
	}
	if proc == nil {
		return
	}
	fmt.Printf("instance left: %v\n", *proc)
	m.Unmount(proc)
	m.procs[index] = nil
	m.procs = append(m.procs[:index], m.procs[index+1:]...)
}

type WorkerConnection struct {
	conn   net.Conn
	master *Master
}

func (w *WorkerConnection) send(data interface{}) error {
	encoder := json.NewEncoder(w.conn)
	return encoder.Encode(data)
}

func (w *WorkerConnection) closed() {
	w.conn.Close()
	w.master.connectionClosed(w)
}

func (w *WorkerConnection) Run() {
	defer w.closed()
	decoder := json.NewDecoder(w.conn)
	for {
		v := Command{}
		err := decoder.Decode(&v)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintln(os.Stderr, err)
			}
			return
		}
		ret := make(chan masterResponse, 1)
		w.master.cmdCh <- masterRequest{cmd: v, conn: w, ret: ret}
		resp := <-ret
		clientResp := Response{}
		if resp.err != nil {
			clientResp.Error = err.Error()
		}
		w.send(clientResp)
	}
}
