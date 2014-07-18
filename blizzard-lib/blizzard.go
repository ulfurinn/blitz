package blizzard

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"bitbucket.org/ulfurinn/blitz"

	"github.com/GeertJohan/go.rice"
)

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

type Master struct {
	execs              []*Executable
	procs              []*Process
	scheduledRemoval   ProcessSet
	routers            map[int]*Router
	routeLock          *sync.RWMutex
	cmdCh              chan masterRequest
	snapshotCh         chan chan *Snapshot
	connectionClosedCh chan *WorkerConnection
	templateCh         chan chan TemplateResponse
	server             *http.Server
}

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
		scheduledRemoval:   make(ProcessSet),
	}
}

func (m *Master) Run() {
	blitz.CreateDirectoryStructure()
	listener, err := net.Listen("unix", blitz.ControlAddress())
	if err != nil {
		fatal(err)
	}
	closeSocketOnShutdown(listener)
	go m.HTTP()
	go m.Loop()
	err = m.BootAllDeployed()
	if err != nil {
		fatal(err)
	}
	m.ProcessControl(listener)
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
	switch req.URL.Path {
	case "/_blitz":
		m.serveSnapshot(resp, req)
	case "/_blitz_ws":

	default:
		m.BlitzDispatch(resp, req)
	}
}

func (m *Master) Loop() {
	var tpl *template.Template
	t := time.NewTicker(time.Second)
	for {
		select {
		case cmd := <-m.cmdCh:
			m.handleCommand(cmd)
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

func (m *Master) ShutdownAndRemoveProcs(procs map[*Process]struct{}) {
	if len(procs) == 0 {
		return
	}
	remaining := []*Process{}
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
	var proc *Process
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

func (m *Master) allMountedInstances() (used ProcessSet) {
	used = make(ProcessSet)
	for _, router := range m.routers {
		for _, i := range router.UsedInstances() {
			used[i] = struct{}{}
		}
	}
	return
}

func (m *Master) allUnusedInstances(used ProcessSet) (unused ProcessSet) {
	unused = make(ProcessSet)
	for _, i := range m.procs {
		_, isUsed := used[i]
		if !isUsed && i.Pid != 0 { // if pid is 0, we haven't received an announce yet
			i.Obsolete = true
			unused[i] = struct{}{}
		}
	}
	return
}

func (m *Master) partitionUnusedInstances(unused ProcessSet) (immediate, scheduled ProcessSet) {
	immediate = make(ProcessSet)
	scheduled = make(ProcessSet)
	for i, _ := range unused {
		if i.dead || atomic.LoadInt64(&i.Requests) == 0 {
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

func (m *Master) cleanupInstances(unused ProcessSet) {
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

func (m *Master) Mount(paths []blitz.PathSpec, proc *Process) {
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

func (m *Master) Unmount(proc *Process) {
	for _, r := range m.routers {
		r.Unmount(proc)
	}
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
	i := &Process{Exe: e, id: id}
	m.execs = append(m.execs, e)
	m.procs = append(m.procs, i)
	i.cmd = exec.Command(exe, "--blitz-proc-id", id)
	err := i.cmd.Start()
	return err
}

func (m *Master) PrintProcList() {
	pids := []int{}
	for _, i := range m.procs {
		pids = append(pids, i.Pid)
	}
	fmt.Println(pids)
}
