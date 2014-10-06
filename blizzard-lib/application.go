package blizzard

import (
	"fmt"
	"io"
	"os"
	"strings"

	"bitbucket.org/ulfurinn/blitz"
	"bitbucket.org/ulfurinn/gen_proc"

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
	log("[exe %s] releasing\n", app.Basename)
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
	log("[app %p] bootstrapping binary=%s adapter=%s config=%s\n", app, app.Exe, app.Adapter, app.Config)
	app.Tag = randstr(32)
	args := []string{"-bootstrap", "-tag", app.Tag}
	args = append(args, app.args()...)
	app.BootstrapCmd = exec.Command(app.executable(), args...)
	log("[app %p] bootstrap command: %s\n", app, strings.Join(app.BootstrapCmd.Args, " "))

	ok := make(chan *blitz.BootstrapCommand, 1)
	died := make(chan struct{}, 1)

	procout, _ := app.BootstrapCmd.StdoutPipe()
	procerr, _ := app.BootstrapCmd.StderrPipe()

	app.server.AddTagCallback(app.Tag, func(cmd interface{}) {
		log("[app %p] received bootstrap\n", app)
		ok <- cmd.(*blitz.BootstrapCommand)
	})

	err := app.BootstrapCmd.Start()
	if err != nil {
		log("[app %p] bootstrap failed: %v\n", app, err)
		app.server.RemoveTagCallback(app.Tag)
		return false, err
	}

	return app.deferBootstrap(func(ret func(error)) {
		go func() {
			go io.Copy(os.Stderr, procout)
			go io.Copy(os.Stderr, procerr)
			err := app.BootstrapCmd.Wait()
			if err != nil {
				log("[app %p] %v\n", app, err)
			}
			died <- struct{}{}
		}()

		select {
		case bootstrapped := <-ok:
			log("[app %p] bootstrap ok\n", app)
			err := app.server.Bootstrapped(bootstrapped)
			log("[app %p] spawned: %v; returning from deferred\n", app, err)
			ret(err)
		case <-died:
			err := fmt.Errorf("process died unexpectedly during bootstrap phase")
			log("[app %p] bootstrap failed: %v\n", app, err)
			ret(err)
		}
	})
}

func (app *Application) spawn(cb SpawnedCallback) (pg *ProcGroup, err error) {
	pg = NewProcGroup(app.server, app)
	go pg.Run()
	err = pg.Spawn(app.Instances, cb)
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
