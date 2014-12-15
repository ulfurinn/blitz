package blitz

import (
	"flag"
	"fmt"
	"os"
)

type PathSpec struct {
	Path    string `json:"path"`
	Version int    `json:"version"`
}

type Bootstrapper func(*Worker, *BootstrapCommand) error
type Runner func(*Worker, *AnnounceCommand) error

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func CreateDirectoryStructure() {
	os.MkdirAll("blitz", os.ModeDir|0775)
	os.MkdirAll("blitz/deploy", os.ModeDir|0775)
	os.MkdirAll("blitz/deploy-old", os.ModeDir|0775)
}

func ControlAddress() string {
	return "blitz/ctl"
}

func InitWorker() {
	flag.StringVar(&tag, "tag", "", "internal tag")
	flag.StringVar(&Config, "config", "", "config")
	flag.BoolVar(&bootstrap, "bootstrap", false, "bootstrap mode")
	flag.Parse()
}
