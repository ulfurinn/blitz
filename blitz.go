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

type Command struct {
	Type string `json:"type"`
}

type DeployCommand struct {
	Command
	Executable string `json:"executable"`
}

type BootstrapCommand struct {
	Command
	BinaryTag string `json:"tag"`
	Instances int    `json:"instances"`
	AppName   string `json:"appName"`
}

type Bootstrapper func(*BootstrapCommand) error

type AnnounceCommand struct {
	Command
	ProcTag  string     `json:"procTag"`
	GroupTag string     `json:"groupTag"`
	CmdTag   string     `json:"cmdTag"`
	Patch    int64      `json:"patch"`
	Paths    []PathSpec `json:"paths"`
	Network  string     `json:"network"`
	Address  string     `json:"address"`
}

type Response struct {
	Error error `json:"error"`
}

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
	flag.BoolVar(&bootstrap, "bootstrap", false, "bootstrap mode")
	flag.Parse()
}

func InitTool() {

}
