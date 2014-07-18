package blizzard

import "os/exec"

type Executable struct {
	Exe          string
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
	e.Tag = randstr(32)
	e.BootstrapCmd = exec.Command(e.Exe, "-bootstrap", "-tag", e.Tag)
	err := e.BootstrapCmd.Start()
	if err == nil {
		go e.BootstrapCmd.Wait()
	}
	return err
}

func (e *Executable) spawn(cb SpawnedCallback) (pg *ProcGroup, err error) {
	pg = NewProcGroup(e.server, e)
	go pg.Run()
	err = pg.Spawn(e.Instances, cb)
	return
}
