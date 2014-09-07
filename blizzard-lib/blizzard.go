package blizzard

import (
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bitbucket.org/ulfurinn/blitz"
)

type Blizzard struct {
	*BlizzardCh `gen_proc:"gen_server"`
	static      *assetServer
	routers     *RouteSet
	execs       []*Executable
	procGroups  []*ProcGroup
	server      *http.Server
	tpl         *template.Template
	tplErr      error
	cleanup     *time.Timer
	inspect     func(interface{})
}

type workerCommand struct {
	command interface{}
	*WorkerConnection
}

type SpawnedCallback func(*ProcGroup)

func NewBlizzard() *Blizzard {
	b := &Blizzard{
		BlizzardCh: NewBlizzardCh(),
		routers:    NewRouteSet(),
	}
	static, err := NewAssetServer(b)
	if err != nil {
		fatal(err)
	}
	b.static = static
	b.inspect = static.Broadcast
	return b
}

func (b *Blizzard) Start() {
	blitz.CreateDirectoryStructure()
	listener, err := net.Listen("unix", blitz.ControlAddress())
	if err != nil {
		fatal(err)
	}
	closeSocketOnShutdown(listener)
	go b.HTTP()
	go b.Run()
	go b.static.HTTP()
	go b.static.Run()
	// err = b.bootAllDeployed()
	// if err != nil {
	// 	fatal(err)
	// }
	b.cleanup = time.NewTimer(time.Second)
	b.cleanup.Stop()
	go func() {
		for {
			<-b.cleanup.C
			b.Cleanup()
		}
	}()
	b.processControl(listener)
}

func (b *Blizzard) HTTP() {
	b.server = &http.Server{
		Addr:    ":8080",
		Handler: b,
	}
	err := b.server.ListenAndServe()
	if err != nil {
		fatal(err)
	}
}

var globalCounter uint64 = 0

func (b *Blizzard) ServeHTTP(resp http.ResponseWriter, req *http.Request) {

	var serve http.HandlerFunc

	b.routers.reading(func() {
		serve = b.routers.ServeHTTP(resp, req)
	})

	if serve != nil {
		serve(resp, req)
	}

}

func (b *Blizzard) processControl(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fatal(err)
		}
		worker := &WorkerConnection{conn: conn, server: b}
		go worker.Run()
	}
}

func (b *Blizzard) handleCommand(cmd workerCommand) (resp blitz.Response) {
	switch command := cmd.command.(type) {
	case blitz.AnnounceCommand:
		b.announce(command, cmd.WorkerConnection)
		return
	case blitz.DeployCommand:
		err := b.deploy(command)
		if err != nil {
			resp.Error = new(string)
			*resp.Error = err.Error()
		}
		log("[blizzard] deploy: %v\n", err)
		return
	case blitz.BootstrapCommand:
		b.bootstrapped(command)
		return
	default:
		return
	}
}

func (b *Blizzard) announce(cmd blitz.AnnounceCommand, worker *WorkerConnection) {
	procGroup, proc := b.findProcByTag(cmd.ProcTag)
	if proc == nil {
		log("[blizzard] no matching proc found for tag %s\n", cmd.ProcTag)
		return
	}
	log("[blizzard] announce from proc group %p proc %p pid %d\n", procGroup, proc, proc.cmd.Process.Pid)
	worker.monitor = b
	procGroup.Announced(proc, cmd, worker)
	//	TODO; what to do in case of patch mismatch?
}

func (b *Blizzard) deploy(cmd blitz.DeployCommand) error {
	components := strings.Split(cmd.Executable, string(os.PathSeparator))
	basename := components[len(components)-1]
	deployedName := fmt.Sprintf("%s.blitz%d", basename, time.Now().Unix())
	newname := fmt.Sprintf("blitz/deploy/%s", deployedName)
	origin, err := os.Open(cmd.Executable)
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
	err = newfile.Close()
	if err != nil {
		return err
	}

	return b.bootstrap(newname)
}

func (b *Blizzard) bootstrap(command string) error {
	e := &Executable{Exe: command, Basename: filepath.Base(command), server: b}
	b.execs = append(b.execs, e)
	return e.bootstrap()
}

func (b *Blizzard) bootstrapped(cmd blitz.BootstrapCommand) {
	e := b.findExeByTag(cmd.BinaryTag)
	if e == nil {
		log("[blizzard] no matching binary for tag %s\n", cmd.BinaryTag)
		return
	}
	log("[blizzard] bootstrapped %p: %d instances of %s\n", e, cmd.Instances, cmd.AppName)
	e.Instances = cmd.Instances
	e.AppName = cmd.AppName
	pg, err := e.spawn(b.mount)
	if err == nil {
		b.procGroups = append(b.procGroups, pg)
	} else {
		log("[blizzard] while spawning: %v\n, err")
	}
}

func (b *Blizzard) mount(proc *ProcGroup) {
	log("[blizzard] mounting proc %p\n", proc)
	b.routers.writing(func() {
		for _, path := range proc.paths {
			if len(path.Path) > 0 {
				if path.Path[0] == '/' {
					path.Path = path.Path[1:]
				}
			}
			split := strings.Split(path.Path, "/")
			router := b.routers.forVersion(path.Version, true)
			router.Mount(split, proc, "")
		}
		b.scheduleCleanup()
	})
}

// func (b *Blizzard) bootAllDeployed() error {
// 	return filepath.Walk("blitz/deploy", func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !info.IsDir() && info.Mode().Perm()&0111 > 0 {
// 			err := b.bootDeployed(path, 1)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		return nil
// 	})
// }

func (b *Blizzard) findExeByTag(tag string) *Executable {
	for _, e := range b.execs {
		if e.Tag == tag {
			return e
		}
	}
	return nil
}

func (b *Blizzard) findProcByTag(tag string) (group *ProcGroup, proc *Process) {
	for _, pg := range b.procGroups {
		for _, p := range pg.Procs {
			if p.tag == tag {
				group = pg
				proc = p
				return
			}
		}
		for _, p := range pg.PendingProcs {
			if p.tag == tag {
				group = pg
				proc = p
				return
			}
		}
	}
	return
}

func (b *Blizzard) findProcByConnection(w *WorkerConnection) (group *ProcGroup, proc *Process) {
out:
	for _, pg := range b.procGroups {
		for _, p := range pg.Procs {
			if p.connection == w {
				group = pg
				proc = p
				break out
			}
		}
	}
	return
}

func findProcGroup(p *ProcGroup, list []*ProcGroup) (index int, found bool) {
	for i, pr := range list {
		if pr == p {
			index = i
			found = true
			return
		}
	}
	return
}

func removeProcGroup(index int, list []*ProcGroup) (result []*ProcGroup) {
	list[index] = nil
	result = append(result, list[:index]...)
	result = append(result, list[index+1:]...)
	return
}

func (b *Blizzard) removeGroup(pg *ProcGroup) {
	b.routers.writing(func() {
		log("[blizzard] removing proc group %p\n", pg)
		b.unmount(pg)
		index, found := findProcGroup(pg, b.procGroups)
		if found {
			b.procGroups = removeProcGroup(index, b.procGroups)
		}
		b.scheduleCleanup()
	})
}

func (b *Blizzard) unmount(proc *ProcGroup) {
	for _, r := range b.routers.routers {
		r.Unmount(proc)
	}
}

func (b *Blizzard) scheduleCleanup() {
	b.cleanup.Reset(10 * time.Millisecond)
}

func (b *Blizzard) handleCleanup() {
	b.routers.writing(func() {
		used := b.routers.UsedInstances()
		unused := b.unusedHandlers(used)
		for pg := range unused {
			log("[blizzard] shutting down unused proc group %p\n", pg)
			pg.Shutdown()
		}
	})
	return
}

func (b *Blizzard) unusedHandlers(used ProgGroupSet) (result ProgGroupSet) {
	result = make(ProgGroupSet)
	for _, pg := range b.procGroups {
		if _, isUsed := used[pg]; !isUsed {
			result[pg] = struct{}{}
		}
	}
	return
}

func (b *Blizzard) handleWorkerClosed(w *WorkerConnection) {
	pg, i := b.findProcByConnection(w)
	if i == nil {
		return
	}
	pg.Remove(i)
}

func (b *Blizzard) handleSnapshot(f func(interface{})) {
	// for _, e := range b.execs {
	// 	s.Execs = append(s.Execs, e)
	// }
	// for _, i := range b.procGroups {
	// 	s.Procs = append(s.Procs, i)
	// }
	// for v, router := range b.routers.routers {
	// 	flat := router.snapshot()
	// 	for _, r := range flat {
	// 		r.Version = v
	// 	}
	// 	s.Routes = append(s.Routes, flat...)
	// }
	b.routers.reading(func() {
		for v, router := range b.routers.routers {
			flat := router.snapshot()
			for _, r := range flat {
				r.Version = v
				f(map[string]interface{}{"type": "add-route", "data": r})
			}
		}
	})
	for _, pg := range b.procGroups {
		f(ProcGroupInspect(pg))
		for _, i := range pg.Procs {
			f(ProcInspect(i))
		}
	}
}
