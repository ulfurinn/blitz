package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"bitbucket.org/ulfurinn/blitz"
)
import "gopkg.in/yaml.v1"

type BlitzExt struct {
	Patch   uint64
	AppName string
	Mount   []struct {
		Version int
		Paths   []string
	}
}

func (c BlitzExt) toPathSpec() (result []blitz.PathSpec) {
	for _, m := range c.Mount {
		for _, p := range m.Paths {
			result = append(result, blitz.PathSpec{Version: m.Version, Path: p})
		}
	}
	return
}

type ThinConfig struct {
	Servers            int
	Chdir              string
	Environment        string
	MaxConns           int
	MaxPersistentConns int
	Timeout            int
	Wait               int
	Log                string
	Require            []string
	User               string
	Group              string
	Threaded           bool
	Tag                string
	Blitz              BlitzExt
}

func main() {
	blitz.InitWorker()
	var thin ThinConfig
	file, err := ioutil.ReadFile(blitz.Config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(file, &thin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	w := blitz.Worker{
		AppName: thin.Blitz.AppName,
		Patch:   thin.Blitz.Patch,
		Paths:   thin.Blitz.toPathSpec(),
		Bootstrap: func(w *blitz.Worker, cmd *blitz.BootstrapCommand) error {
			cmd.Instances = thin.Servers
			return nil
		},
		Run: func(w *blitz.Worker, cmd *blitz.AnnounceCommand) error {
			thinProc := exec.Command("thin", "-S", cmd.Address, "start")
			stdout, err := thinProc.StdoutPipe()
			if err != nil {
				return err
			}
			stderr, err := thinProc.StderrPipe()
			if err != nil {
				return err
			}
			go io.Copy(os.Stdout, stdout)
			go io.Copy(os.Stderr, stderr)
			err = thinProc.Start()
			if err != nil {
				return err
			}
			err = checkConnection(cmd.Address)
			t := time.Now()
			for err != nil && time.Since(t) < 30*time.Second {
				time.Sleep(50 * time.Millisecond)
				err = checkConnection(cmd.Address)
			}
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			w.Mainproc.Add(1)
			go func() {
				thinProc.Wait()
				w.Mainproc.Done()
			}()
			w.OnShutdown(func() {
				thinProc.Process.Signal(os.Interrupt)
			})
			return nil
		},
	}
	err = w.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
	}
	w.Wait()
}

func checkConnection(socket string) error {
	c := http.Client{}
	c.Transport = blitz.SubdirUnixTransport
	parts := strings.Split(socket, "/")
	_, err := c.Get(fmt.Sprintf("http://%s/", parts[1]))
	return err
}
