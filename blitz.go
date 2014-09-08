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
	Executable string `json:"executable,omitempty"`
	Adapter    string `json:"adapter,omitempty"`
	Config     string `json:"config,omitempty"`
}

type BootstrapCommand struct {
	Command
	BinaryTag string `json:"tag"`
	Instances int    `json:"instances"`
	AppName   string `json:"appName"`
}

type Bootstrapper func(*Worker, *BootstrapCommand) error
type Runner func(*Worker, *AnnounceCommand) error

type AnnounceCommand struct {
	Command
	ProcTag  string     `json:"procTag"`
	GroupTag string     `json:"groupTag"`
	CmdTag   string     `json:"cmdTag"`
	Patch    uint64     `json:"patch"`
	Paths    []PathSpec `json:"paths"`
	Network  string     `json:"network"`
	Address  string     `json:"address"`
}

type ListExecutablesCommand struct {
	Command
}

type Response struct {
	Error *string `json:"error"`
}

type ListExecutablesResponse struct {
	Response
	Executables []string `json:"execs"`
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
	flag.StringVar(&Config, "config", "", "config")
	flag.BoolVar(&bootstrap, "bootstrap", false, "bootstrap mode")
	flag.Parse()
}
