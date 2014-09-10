package blizzard

import "os/exec"

type Executable struct {
	Exe          string
	Adapter      string
	Config       string
	Basename     string
	Tag          string
	AppName      string
	Instances    int
	Obsolete     bool
	BootstrapCmd *exec.Cmd
	server       *Blizzard
}

func (e *Executable) release() {
	log("[exe %s] releasing\n", e.Basename)
	//os.Rename(e.Exe, fmt.Sprintf("blitz/deploy-old/%s", e.Basename))
	e.Obsolete = true
}

func (e *Executable) bootstrap() error {
	log("[app %p] bootstrapping binary=%s adapter=%s config=%s\n", e, e.Exe, e.Adapter, e.Config)
	e.Tag = randstr(32)
	args := []string{"-bootstrap", "-tag", e.Tag}
	args = append(args, e.args()...)
	e.BootstrapCmd = exec.Command(e.executable(), args...)
	err := e.BootstrapCmd.Start()
	if err == nil {
		go e.BootstrapCmd.Wait()
	} else {
		log("[app %p] bootstrap failed: %v\n", e, err)
	}
	return err
}

func (e *Executable) spawn(cb SpawnedCallback) (pg *ProcGroup, err error) {
	pg = NewProcGroup(e.server, e)
	go pg.Run()
	err = pg.Spawn(e.Instances, cb)
	return
}

func (e *Executable) takeover(old *ProcGroup, cb SpawnedCallback) (pg *ProcGroup) {
	pg = NewProcGroup(e.server, e)
	go pg.Run()
	go pg.Takeover(old, cb)
	return
}

func (e *Executable) executable() string {
	if e.Exe != "" {
		return e.Exe
	}
	return "blitz-adapter-" + e.Adapter
}

func (e *Executable) args() []string {
	if e.Exe != "" {
		return []string{}
	}
	return []string{"-config", e.Config}
}
