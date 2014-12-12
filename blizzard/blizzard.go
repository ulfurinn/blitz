package blizzard

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"log/syslog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"time"

	"bitbucket.org/ulfurinn/blitz"
	"bitbucket.org/ulfurinn/gen_proc"

	"gopkg.in/yaml.v1"
)

type ExeConfig struct {
	Type   string
	Binary string
	Config string
}

type BlizzardConfig struct {
	Executables []ExeConfig
	Logger      struct {
		Type           string
		SyslogSeverity syslog.Priority
		SyslogFacility syslog.Priority
	}
}

type tagCallbackSet map[string]TagCallback

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
	b.readConfig()
	b.createLogger()
	Logger().Printf("[blizzard] starting\n")
	blitz.CreateDirectoryStructure()
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
		Logger().Printf("[blizzard] error starting configured applications: %v\n", err)
	}
	b.writeConfig()
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
	Logger().Printf("[blizzard] command: %v\n", reflect.TypeOf(cmd.command))
	switch command := cmd.command.(type) {
	case *blitz.AnnounceCommand:
		b.RunTagCallback(command.Tag, command, cmd.WorkerConnection)
		return blitz.Response{}
	case *blitz.DeployCommand:
		resp := blitz.Response{}
		app, err := b.Deploy(command)
		if err != nil {
			resp.Error = new(string)
			*resp.Error = err.Error()
			return resp
		}
		Logger().Printf("[blizzard] bootstrapping\n")
		err = app.Bootstrap()
		Logger().Printf("[blizzard] bootstrap complete: %v\n", err)
		if err != nil {
			resp.Error = new(string)
			*resp.Error = err.Error()
		}
		return resp
	case *blitz.BootstrapCommand:
		b.RunTagCallback(command.Tag, command, cmd.WorkerConnection)
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

type TagCallback func(cmd interface{}, connection *WorkerConnection)

func (b *Blizzard) handleAddTagCallback(tag string, cb TagCallback) {
	b.callbacks[tag] = cb
}

func (b *Blizzard) handleRemoveTagCallback(tag string) {
	delete(b.callbacks, tag)
}

func (b *Blizzard) handleRunTagCallback(tag string, data interface{}, w *WorkerConnection) {
	cb, ok := b.callbacks[tag]
	delete(b.callbacks, tag)
	if ok {
		cb(data, w)
	}
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
		Logger().Printf("[blizzard] %v\n", err)
		return
	}
	defer f.Close()
	yml, err := yaml.Marshal(b.config)
	if err != nil {
		Logger().Printf("[blizzard] %v\n", err)
		return
	}
	_, err = f.Write(yml)
	if err != nil {
		Logger().Printf("[blizzard] %v\n", err)
		return
	}
}

func (b *Blizzard) readConfig() {
	f, err := os.Open(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			Logger().Printf("[blizzard] %v\n", err)
		}
		return
	}
	defer f.Close()
	yml, err := ioutil.ReadAll(f)
	if err != nil {
		Logger().Printf("[blizzard] %v\n", err)
		return
	}
	err = yaml.Unmarshal(yml, &b.config)
	if err != nil {
		Logger().Printf("[blizzard] %v\n", err)
		return
	}
}

func (b *Blizzard) createLogger() {
	if b.config.Logger.Type == "" {
		b.config.Logger.Type = "stderr"
	}
	logFlag := log.Ldate | log.Ltime | log.Lshortfile
retry:
	fmt.Fprintf(os.Stderr, "[blizzard] logging to %s\n", b.config.Logger.Type)
	switch b.config.Logger.Type {
	case "stderr":
		SetLogger(log.New(os.Stderr, "", logFlag))
	case "syslog":
		l, err := syslog.NewLogger(b.config.Logger.SyslogFacility|b.config.Logger.SyslogSeverity, logFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[blizzard] error initializing syslog: %v\n", err)
			goto retry
		}
		SetLogger(l)
	}
}

func (b *Blizzard) handleBootstrapped(app *Application) (gen_proc.Deferred, error) {
	pg := app.createProcGroup()
	b.procGroups = append(b.procGroups, pg)
	return b.deferBootstrapped(func(ret func(error)) {
		Logger().Printf("[blizzard] spawning procgroup %p\n", pg)
		err := pg.Spawn()
		if err == nil {
			Logger().Printf("[blizzard] spawned\n")
			b.mount(pg)
			b.apps = append(b.apps, app)
			b.addExeConfig(app)
		} else {
			app.Stop()
			Logger().Printf("[blizzard] while spawning: %v\n", err)
			pg.Shutdown() // TODO: make sure this does not conflict with pg apoptosis
		}
		ret(err)
	})
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

func (b *Blizzard) mount(pg *ProcGroup) {
	Logger().Printf("[blizzard] mounting procgroup %p\n", pg)
	b.routers.writing(func() {
		for _, path := range pg.paths {
			if len(path.Path) > 0 {
				if path.Path[0] == '/' {
					path.Path = path.Path[1:]
				}
			}
			split := strings.Split(path.Path, "/")
			router := b.routers.forVersion(path.Version, true)
			router.Mount(split, pg, "")
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
		Logger().Printf("[blizzard] removing proc group %p\n", pg)
		b.unmount(pg)
		index, found := findProcGroup(pg, b.procGroups)
		if found {
			b.procGroups = removeProcGroup(index, b.procGroups)
			pg.busyWg.Wait()
			pg.inspectDispose()
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
			Logger().Printf("[blizzard] shutting down unused proc group %p\n", pg)
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
