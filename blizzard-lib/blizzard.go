package blizzard

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v1"

	"bitbucket.org/ulfurinn/blitz"
)

type ExeConfig struct {
	Type   string
	Binary string
	Config string
}

type BlizzardConfig struct {
	Executables []ExeConfig
}

type tagCallbackSet map[string]func(interface{})

type Blizzard struct {
	*BlizzardCh `gen_proc:"gen_server"`
	config      BlizzardConfig
	static      *assetServer
	routers     *RouteSet
	apps        []*Application
	procGroups  []*ProcGroup
	server      *http.Server
	tpl         *template.Template
	tplErr      error
	cleanup     *time.Timer
	inspect     func(interface{})
	callbacks   tagCallbackSet
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
		callbacks:  make(tagCallbackSet),
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
	b.readConfig()
	b.cleanup = time.NewTimer(time.Second)
	b.cleanup.Stop()
	go func() {
		for {
			<-b.cleanup.C
			b.Cleanup()
		}
	}()
	listener, err := net.Listen("unix", blitz.ControlAddress())
	if err != nil {
		fatal(err)
	}
	closeSocketOnShutdown(listener)
	go b.HTTP()
	go b.Run()
	go b.static.HTTP()
	go b.static.Run()
	go b.processControl(listener)
	err = b.bootAllDeployed()
	if err != nil {
		fatal(err)
	}
	select {}
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

func (b *Blizzard) Command(cmd workerCommand) interface{} {
	log("[blizzard] command: %v\n", reflect.TypeOf(cmd.command))
	switch command := cmd.command.(type) {
	case *blitz.AnnounceCommand:
		b.Announce(command, cmd.WorkerConnection)
		return blitz.Response{}
	case *blitz.DeployCommand:
		resp := blitz.Response{}
		app, err := b.Deploy(command)
		if err != nil {
			resp.Error = new(string)
			*resp.Error = err.Error()
			return resp
		}
		log("[blizzard] deployed, bootstrapping\n")
		err = app.Bootstrap()
		log("[blizzard] bootstrap return")
		if err != nil {
			resp.Error = new(string)
			*resp.Error = err.Error()
		}
		return resp
	case *blitz.BootstrapCommand:
		b.RunTagCallback(command.BinaryTag, command)
		return blitz.Response{}
	case *blitz.ListExecutablesCommand:
		resp := blitz.ListExecutablesResponse{Executables: []string{}}
		b.GenCall(func() interface{} {
			for _, app := range b.apps {
				resp.Executables = append(resp.Executables, app.AppName)
			}
			return nil
		})
		return resp
	case *blitz.ProcStatCommand:
		resp := blitz.ProcStatResponse{}
		resp.Apps = atomic.LoadInt32(&ApplicationProcCounter)
		resp.ProcGroups = atomic.LoadInt32(&ProcGroupProcCounter)
		resp.Procs = atomic.LoadInt32(&ProcessProcCounter)
		return resp
	case *blitz.RestartTakeoverCommand:
		resp := blitz.RestartTakeoverResponse{}
		err := b.Takeover(command.App, command.Kill)
		if err != nil {
			e := err.Error()
			resp.Error = &e
		}
		return resp
	default:
		resp := blitz.Response{}
		resp.Error = new(string)
		*resp.Error = fmt.Sprintf("unsupported command type %v", reflect.TypeOf(cmd.command))
		return resp
	}
}

func (b *Blizzard) handleAddTagCallback(tag string, cb func(interface{})) {
	b.callbacks[tag] = cb
}

func (b *Blizzard) handleRemoveTagCallback(tag string) {
	delete(b.callbacks, tag)
}

func (b *Blizzard) handleRunTagCallback(tag string, data interface{}) {
	cb, ok := b.callbacks[tag]
	delete(b.callbacks, tag)
	if ok {
		cb(data)
	}
}

func (b *Blizzard) handleAnnounce(cmd *blitz.AnnounceCommand, worker *WorkerConnection) {
	procGroup, proc := b.findProcByTag(cmd.Tag)
	if proc == nil {
		log("[blizzard] no matching proc found for tag %s\n", cmd.Tag)
		return
	}
	log("[blizzard] announce from proc group %p proc %p\n", procGroup, proc)
	worker.monitor = b
	log("[blizzard] announced paths: %v\n", cmd.Paths)
	procGroup.Announced(proc, cmd, worker)
	//	TODO; what to do in case of patch mismatch?
}

func (b *Blizzard) handleDeploy(cmd *blitz.DeployCommand) (*Application, error) {

	app := &Application{
		ApplicationGen: NewApplicationGen(),
		server:         b,
	}

	if cmd.Application != "" {
		components := strings.Split(cmd.Application, string(os.PathSeparator))
		basename := components[len(components)-1]
		deployedName := fmt.Sprintf("%s.blitz%d", basename, time.Now().Unix())
		command := fmt.Sprintf("blitz/deploy/%s", deployedName)
		origin, err := os.Open(cmd.Application)
		if err != nil {
			return nil, err
		}
		newfile, err := os.Create(command)
		if err != nil {
			return nil, err
		}
		err = os.Chmod(command, 0775)
		if err != nil {
			return nil, err
		}
		_, err = io.Copy(newfile, origin)
		if err != nil {
			return nil, err
		}
		err = newfile.Close()
		if err != nil {
			return nil, err
		}
		app.Exe = command
		app.Basename = filepath.Base(command)
	} else {
		app.Adapter = cmd.Adapter
		app.Config = cmd.Config
	}

	b.apps = append(b.apps, app)
	b.addExeConfig(app)
	go app.Run()
	return app, nil
}

func (b *Blizzard) addExeConfig(app *Application) {
	c := ExeConfig{}
	if app.Exe != "" {
		c.Type = "native"
		c.Binary = app.Basename
	} else {
		c.Type = app.Adapter
		c.Config = app.Config
	}
	b.config.Executables = append(b.config.Executables, c)
	b.writeConfig()
}

var configPath = "blitz/blizzard.yml"

func (b *Blizzard) writeConfig() {
	f, err := os.Create(configPath)
	if err != nil {
		log("[blizzard] %v\n", err)
		return
	}
	defer f.Close()
	yml, err := yaml.Marshal(b.config)
	if err != nil {
		log("[blizzard] %v\n", err)
		return
	}
	_, err = f.Write(yml)
	if err != nil {
		log("[blizzard] %v\n", err)
		return
	}
}

func (b *Blizzard) readConfig() {
	f, err := os.Open(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log("[blizzard] %v\n", err)
		}
		return
	}
	defer f.Close()
	yml, err := ioutil.ReadAll(f)
	if err != nil {
		log("[blizzard] %v\n", err)
		return
	}
	err = yaml.Unmarshal(yml, &b.config)
	if err != nil {
		log("[blizzard] %v\n", err)
		return
	}
}

func (b *Blizzard) handleBootstrapped(cmd *blitz.BootstrapCommand) error {
	log("[blizzard] bootstrapped, spawning\n")
	app := b.findExeByTag(cmd.BinaryTag)
	if app == nil {
		log("[blizzard] no matching binary for tag %s\n", cmd.BinaryTag)
		return fmt.Errorf("unknown tag %s", cmd.BinaryTag)
	}
	log("[blizzard] bootstrapped %p: %d instances of %s\n", app, cmd.Instances, cmd.AppName)
	app.Instances = cmd.Instances
	app.AppName = cmd.AppName
	pg, err := app.spawn(b.mount)
	if err == nil {
		app.inspect()
		b.procGroups = append(b.procGroups, pg)
	} else {
		log("[blizzard] while spawning: %v\n, err")
	}
	log("[blizzard] spawned\n")
	return err
}

func (b *Blizzard) handleTakeover(app string, kill bool) (err error) {
	e := b.findAppByName(app)
	if e == nil {
		return fmt.Errorf("unknown app: %s", app)
	}
	pg := b.findGroupByApp(e)
	if e == nil {
		return fmt.Errorf("no proc group for app: %s", app)
	}
	newPg := e.takeover(pg, b.mount, kill)
	b.procGroups = append(b.procGroups, newPg)
	return
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

func (b *Blizzard) bootAllDeployed() error {
	for _, c := range b.config.Executables {
		app := &Application{ApplicationGen: NewApplicationGen(), server: b}
		go app.Run()
		if c.Type == "native" {
			app.Basename = c.Binary
			app.Exe = fmt.Sprintf("blitz/deploy/%s", c.Binary)
		} else {
			app.Adapter = c.Type
			app.Config = c.Config
		}
		b.apps = append(b.apps, app)
		if err := app.Bootstrap(); err != nil {
			return err
		}
	}
	return nil
}

func (b *Blizzard) findAppByName(name string) *Application {
	for _, e := range b.apps {
		if e.AppName == name {
			return e
		}
	}
	return nil
}

func (b *Blizzard) findExeByTag(tag string) *Application {
	for _, e := range b.apps {
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

func (b *Blizzard) findGroupByApp(e *Application) (pg *ProcGroup) {
	for _, pg := range b.procGroups {
		if pg.exe == e {
			return pg
		}
	}
	return nil
}

func (b *Blizzard) findProcByConnection(w *WorkerConnection) (group *ProcGroup, proc *Process) {
	for _, group = range b.procGroups {
		if proc = group.FindProcByConnection(w); proc != nil {
			return
		}
	}
	return nil, nil
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
			pg.busyWg.Wait()
			pg.Stop()
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
		if !pg.IsReady() {
			continue
		}
		if _, isUsed := used[pg]; !isUsed {
			result[pg] = struct{}{}
		}
	}
	return
}

func (b *Blizzard) handleWorkerClosed(w *WorkerConnection) {
	log("[blizzard] lost connection: %s\n", w.connType)
	pg, i := b.findProcByConnection(w)
	if i == nil {
		return
	}
	pg.RemoveProc(i)
}

func (b *Blizzard) handleSnapshot(f func(interface{})) {
	// for _, e := range b.apps {
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
				f(map[string]interface{}{"type": "add-route", "pathSpec": r})
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
