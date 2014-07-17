package blizzard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"bitbucket.org/ulfurinn/blitz"

	"github.com/GeertJohan/go.rice"
)

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

type SnapshotRoute struct {
	Path          string
	Version       int
	Instance      *Instance
	Requests      int64
	TotalRequests uint64
	Written       uint64
}

type Snapshot struct {
	Execs  []*Executable
	Procs  []*Instance
	Routes []*SnapshotRoute
}

type TemplateResponse struct {
	tpl *template.Template
	err error
}

type Master struct {
	execs              []*Executable
	procs              []*Instance
	scheduledRemoval   InstanceSet
	routers            map[int]*Router
	routeLock          *sync.RWMutex
	cmdCh              chan masterRequest
	snapshotCh         chan chan *Snapshot
	connectionClosedCh chan *WorkerConnection
	templateCh         chan chan TemplateResponse
	server             *http.Server
}

type Executable struct {
	Exe      string
	Basename string
	Obsolete bool
}

type Instance struct {
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
	cmd              *exec.Cmd
}

type InstanceSet map[*Instance]struct{}

type masterRequest struct {
	cmd  blitz.Command
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
		routers:            make(map[int]*Router),
		routeLock:          &sync.RWMutex{},
		cmdCh:              make(chan masterRequest),
		snapshotCh:         make(chan chan *Snapshot),
		connectionClosedCh: make(chan *WorkerConnection),
		templateCh:         make(chan chan TemplateResponse),
		scheduledRemoval:   make(InstanceSet),
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
	err = m.BootAllDeployed()
	if err != nil {
		fatal(err)
	}
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

type VersionStrategy interface {
	Version(*http.Request) (int, []string, error)
}

type PathVersionStrategy struct{}

func (PathVersionStrategy) Version(req *http.Request) (version int, path []string, err error) {
	split := strings.Split(req.URL.Path[1:], "/")
	if len(split) == 0 {
		err = fmt.Errorf("No version provided")
		return
	}
	_, err = fmt.Sscanf(split[0], "v%d", &version)
	if err != nil {
		err = fmt.Errorf("No version provided")
		return
	}
	path = split[1:]
	return
}

func concatPath(path []string) string {
	var buf bytes.Buffer
	for _, c := range path {
		fmt.Fprint(&buf, "/", c)
	}
	return buf.String()
}

func (m *Master) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/_blitz":
		m.serveSnapshot(resp, req)
	case "/_blitz_ws":

	default:
		version, path, err := (PathVersionStrategy{}).Version(req)
		if err != nil {
			resp.WriteHeader(400)
			fmt.Fprint(resp, err)
			return
		}
		m.routeLock.RLock()
		versionRouter, ok := m.routers[version]
		if !ok {
			resp.WriteHeader(400)
			fmt.Fprintf(resp, "Version %d is not recognised", version)
			return
		}
		r, h := versionRouter.Route(path)
		// do this before unlocking so that the collector in announce will see it as busy
		if h != nil {
			atomic.AddInt64(&r.requests, 1)
			atomic.AddUint64(&r.totalRequests, 1)
			atomic.AddInt64(&h.Requests, 1)
			atomic.AddUint64(&h.TotalRequests, 1)
		}
		m.routeLock.RUnlock()
		if h == nil {
			resp.WriteHeader(404)
			return
		}
		counter := &countingResponseWriter{ResponseWriter: resp}
		defer func() {
			atomic.AddInt64(&r.requests, -1)
			atomic.AddInt64(&h.Requests, -1)
			atomic.AddUint64(&r.written, counter.written)
			atomic.AddUint64(&h.Written, counter.written)
		}()
		req.Header.Set("X-Blitz-Path", concatPath(path))
		h.ServeHTTP(counter, req)
	}
}

func (m *Master) serveSnapshot(resp http.ResponseWriter, req *http.Request) {
	tplret := make(chan TemplateResponse)
	m.templateCh <- tplret
	tplresp := <-tplret
	if tplresp.err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte(tplresp.err.Error()))
		return
	}
	ret := make(chan *Snapshot)
	m.snapshotCh <- ret
	snapshot := <-ret
	var generated bytes.Buffer
	err := tplresp.tpl.Execute(&generated, snapshot)
	if err != nil {
		resp.WriteHeader(500)
		resp.Write([]byte(err.Error()))
		return
	}
	io.Copy(resp, &generated)
}

func (i *Instance) makeRevProxy() {
	i.proxy = &httputil.ReverseProxy{
		Transport: blitz.UnixTransport,
		Director: func(newreq *http.Request) {
			newreq.URL.Scheme = "http"
			newreq.URL.Host = i.Address
			newreq.URL.Path = newreq.Header.Get("X-Blitz-Path")
		},
	}
}

func (i *Instance) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	i.proxy.ServeHTTP(resp, req)
}

func (i *Instance) Shutdown() {
	//fmt.Printf("releasing instance %v\n", *i)
	//i.connection.conn.Close()
	syscall.Kill(i.Pid, syscall.SIGINT)
	i.cmd.Wait()
}

func (e *Executable) release() {
	//fmt.Printf("releasing executable %v\n", *e)
	//os.Rename(e.Exe, fmt.Sprintf("blitz/deploy-old/%s", e.Basename))
	e.Obsolete = true
}

func (m *Master) Loop() {
	var tpl *template.Template
	t := time.NewTicker(time.Second)
	for {
		select {
		case cmd := <-m.cmdCh:
			//fmt.Fprintln(os.Stderr, cmd.cmd)
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
		case <-t.C:
			m.cleanupInstances(m.scheduledRemoval)
		case ret := <-m.snapshotCh:
			ret <- m.snapshot()
		case w := <-m.connectionClosedCh:
			m.connectionClosed(w)
		case ret := <-m.templateCh:
			resp := TemplateResponse{}
			if tpl == nil {
				box, err := rice.FindBox("blitz-templates")
				if err != nil {
					resp.err = err
					ret <- resp
					break
				}
				src, err := box.String("status.html")
				if err != nil {
					resp.err = err
					ret <- resp
					break
				}
				tpl, err = template.New("status").Parse(src)
				if err != nil {
					resp.err = err
					ret <- resp
					break
				}
			}
			resp.tpl = tpl
			ret <- resp
		}
	}
}

func (m *Master) snapshot() *Snapshot {
	s := &Snapshot{}
	for _, e := range m.execs {
		s.Execs = append(s.Execs, e)
	}
	for _, i := range m.procs {
		s.Procs = append(s.Procs, i)
	}
	for v, router := range m.routers {
		flat := router.snapshot()
		for _, r := range flat {
			r.Version = v
		}
		s.Routes = append(s.Routes, flat...)
	}
	return s
}

func (m *Master) ShutdownAndRemoveProcs(procs map[*Instance]struct{}) {
	if len(procs) == 0 {
		return
	}
	remaining := []*Instance{}
	for _, i := range m.procs {
		_, removable := procs[i]
		if removable {
			i.Shutdown()
		} else {
			remaining = append(remaining, i)
		}
	}
	m.procs = remaining
}

func (m *Master) Announce(cmd blitz.Command, c *WorkerConnection) {
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
	proc.Pid = cmd.PID
	proc.Patch = cmd.Patch
	proc.network = cmd.Network
	proc.Address = cmd.Address
	m.routeLock.Lock()
	defer m.routeLock.Unlock()
	m.Mount(cmd.Paths, proc)
	m.CollectUnusedInstances()
}

func (m *Master) allMountedInstances() (used InstanceSet) {
	used = make(InstanceSet)
	for _, router := range m.routers {
		for _, i := range router.UsedInstances() {
			used[i] = struct{}{}
		}
	}
	return
}

func (m *Master) allUnusedInstances(used InstanceSet) (unused InstanceSet) {
	unused = make(InstanceSet)
	for _, i := range m.procs {
		_, isUsed := used[i]
		if !isUsed {
			i.Obsolete = true
			unused[i] = struct{}{}
		}
	}
	return
}

func (m *Master) partitionUnusedInstances(unused InstanceSet) (immediate, scheduled InstanceSet) {
	immediate = make(InstanceSet)
	scheduled = make(InstanceSet)
	for i, _ := range unused {
		if atomic.LoadInt64(&i.Requests) == 0 {
			immediate[i] = struct{}{}
		} else {
			scheduled[i] = struct{}{}
		}
	}
	return
}

func (m *Master) CollectUnusedInstances() {
	used := m.allMountedInstances()
	unused := m.allUnusedInstances(used)
	m.cleanupInstances(unused)
}

func (m *Master) cleanupInstances(unused InstanceSet) {
	immediate, scheduled := m.partitionUnusedInstances(unused)
	m.ShutdownAndRemoveProcs(immediate)
	m.scheduledRemoval = scheduled
	if len(immediate) > 0 {
		m.CollectUnusedBinaries()
	}
}

func (m *Master) CollectUnusedBinaries() {
	usedBinaries := make(map[*Executable]struct{})
	for _, i := range m.procs {
		usedBinaries[i.Exe] = struct{}{}
	}
	for _, e := range m.execs {
		if _, isUsed := usedBinaries[e]; !isUsed {
			e.release()
		}
	}
}

func (m *Master) Mount(paths []blitz.PathSpec, proc *Instance) {
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
		router.Mount(split, proc, "")
	}
}

func (m *Master) Unmount(proc *Instance) {

}

func (m *Master) Deploy(exe string) error {
	components := strings.Split(exe, string(os.PathSeparator))
	basename := components[len(components)-1]
	deployedName := fmt.Sprintf("%s.blitz%d", basename, time.Now().Unix())
	newname := fmt.Sprintf("blitz/deploy/%s", deployedName)
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

func (m *Master) BootAllDeployed() error {
	return filepath.Walk("blitz/deploy", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Mode().Perm()&0111 > 0 {
			err := m.BootDeployed(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (m *Master) BootDeployed(exe string) error {
	id := randstr(32)
	e := &Executable{Exe: exe, Basename: filepath.Base(exe)}
	i := &Instance{Exe: e, id: id}
	m.execs = append(m.execs, e)
	m.procs = append(m.procs, i)
	i.cmd = exec.Command(exe, "--blitz-proc-id", id)
	err := i.cmd.Start()
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
	//fmt.Printf("instance left: %v\n", *proc)
	m.Unmount(proc)
	m.procs[index] = nil
	m.procs = append(m.procs[:index], m.procs[index+1:]...)
}

func (m *Master) PrintProcList() {
	pids := []int{}
	for _, i := range m.procs {
		pids = append(pids, i.Pid)
	}
	fmt.Println(pids)
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
	w.master.connectionClosedCh <- w
}

func (w *WorkerConnection) Run() {
	defer w.closed()
	decoder := json.NewDecoder(w.conn)
	for {
		v := blitz.Command{}
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
		clientResp := blitz.Response{}
		if resp.err != nil {
			clientResp.Error = resp.err.Error()
		}
		w.send(clientResp)
	}
}
