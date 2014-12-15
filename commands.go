package blitz

import "fmt"

const (
	CmdDeploy       = "deploy"
	CmdListApps     = "list-apps"
	CmdProcStats    = "proc-stats"
	CmdTakeover     = "restart-takeover"
	CmdConfigLogger = "config-logger"
)

type Command struct {
	Type string `json:"type"`
}

type Response struct {
	Error *string `json:"error"`
}

func (r Response) Err() error {
	if r.Error == nil {
		return nil
	}
	return fmt.Errorf(*r.Error)
}

type ConnectionTypeCommand struct {
	Command
	ConnectionType string `json:"connType"`
}

type DeployCommand struct {
	Command
	Application string `json:"executable,omitempty"`
	Adapter     string `json:"adapter,omitempty"`
	Config      string `json:"config,omitempty"`
}

type BootstrapCommand struct {
	Command
	Tag       string `json:"tag"`
	Instances int    `json:"instances"`
	AppName   string `json:"appName"`
}

type AnnounceCommand struct {
	Command
	Tag     string     `json:"tag"`
	Patch   uint64     `json:"patch"`
	Paths   []PathSpec `json:"paths"`
	Network string     `json:"network"`
	Address string     `json:"address"`
}

type ListExecutablesCommand struct {
	Command
}

type ListExecutablesResponse struct {
	Response
	Executables []string `json:"execs"`
}

type ProcStatCommand struct {
	Command
}

type ProcStatResponse struct {
	Response
	Apps       int32 `json:"apps"`
	ProcGroups int32 `json:"procGroups"`
	Procs      int32 `json:"procs"`
}

type RestartTakeoverCommand struct {
	Command
	App  string `json:"app"`
	Kill bool   `json:"kill"`
}

type RestartTakeoverResponse struct {
	Response
}

type ConfigLoggerCommand struct {
	Command
	LogType string `json:"logType"`
	Syslog  struct {
		Facility string `json:"facility"`
		Severity string `json:"severity"`
	} `json:"syslog"`
}

type ConfigLoggerResponse struct {
	Response
}
