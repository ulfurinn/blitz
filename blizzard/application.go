package blizzard

import (
	"fmt"
	"io"
	"strings"

	"bytes"

	"bitbucket.org/ulfurinn/blitz"
	"bitbucket.org/ulfurinn/gen_proc"

	"os"
	"os/exec"
)

type Application struct {
	*ApplicationGen `gen_proc:"gen_server"`
	Exe             string
	Adapter         string
	Config          string
	Basename        string
	Tag             string
	AppName         string
	Instances       int
	Obsolete        bool
	BootstrapCmd    *exec.Cmd
	server          *Blizzard
}

func (app *Application) release() {
	Logger().Printf("[app %s] releasing\n", app.Basename)
	//os.Rename(app.Exe, fmt.Sprintf("blitz/deploy-old/%s", app.Basename))
	app.Obsolete = true
}

func (app *Application) inspect() {
	app.server.inspect(AppInspect(app))
}
func (app *Application) inspectDispose() {
	app.server.inspect(AppInspectDispose(app))
}

func (app *Application) handleBootstrap() (gen_proc.Deferred, error) {
	Logger().Printf("[app %p] bootstrapping binary=%s adapter=%s config=%s\n", app, app.Exe, app.Adapter, app.Config)
	app.Tag = randstr(32)
	args := []string{"-bootstrap", "-tag", app.Tag}
	args = append(args, app.args()...)
	app.BootstrapCmd = exec.Command(app.executable(), args...)
	Logger().Printf("[app %p] bootstrap command: %s\n", app, strings.Join(app.BootstrapCmd.Args, " "))

	ok := make(chan *blitz.BootstrapCommand, 1)

	procout, _ := app.BootstrapCmd.StdoutPipe()
	procerr, _ := app.BootstrapCmd.StderrPipe()

	app.server.AddTagCallback(app.Tag, func(cmd interface{}, w *WorkerConnection) {
		Logger().Printf("[app %p] received bootstrap\n", app)
		ok <- cmd.(*blitz.BootstrapCommand)
	})

	err := app.BootstrapCmd.Start()
	if err != nil {
		Logger().Printf("[app %p] bootstrap failed: %v\n", app, err)
		app.server.RemoveTagCallback(app.Tag)
		return false, err
	}

	return app.deferBootstrap(func(ret func(error)) {
		died := make(chan struct{}, 1)
		var outlog bytes.Buffer
		var errlog bytes.Buffer
		go func() {
			go io.Copy(&outlog, procout)
			go io.Copy(&errlog, procerr)
			err := app.BootstrapCmd.Wait()
			if err != nil {
				Logger().Printf("[app %p] %v\n", app, err)
			}
			died <- struct{}{}
		}()

		select {
		case cmd := <-ok:
			if err := procout.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			if err := procerr.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			Logger().Printf("[app %p] bootstrapped: %d instances of %s\n", app, cmd.Instances, cmd.AppName)
			app.Instances = cmd.Instances
			app.AppName = cmd.AppName
			err := app.server.Bootstrapped(app)
			app.inspect()
			Logger().Printf("[app %p] spawned: %v\n", app, err)
			ret(err)
		case <-died:
			err := fmt.Errorf("process died unexpectedly during bootstrap phase\nstdout:\n%s\nstderr:\n%s", outlog.Bytes(), errlog.Bytes())
			Logger().Printf("[app %p] bootstrap failed: %v\n", app, err)
			ret(err)
		}
	})
}

func (app *Application) createProcGroup() (pg *ProcGroup) {
	pg = NewProcGroup(app.server, app)
	go pg.Run()
	return
}

func (app *Application) takeover(old *ProcGroup, cb SpawnedCallback, kill bool) (pg *ProcGroup) {
	pg = NewProcGroup(app.server, app)
	go pg.Run()
	go pg.Takeover(old, cb, kill)
	return
}

func (app *Application) executable() string {
	if app.Exe != "" {
		return app.Exe
	}
	return "blitz-adapter-" + app.Adapter
}

func (app *Application) args() []string {
	if app.Exe != "" {
		return []string{}
	}
	return []string{"-config", app.Config}
}
