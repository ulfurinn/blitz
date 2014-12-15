package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"bitbucket.org/ulfurinn/blitz"
	"bitbucket.org/ulfurinn/cli"
)

type Commander struct {
}

type Cli struct {
	conn net.Conn
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func (c *Cli) Run() {
	app := cli.NewApp()
	app.Name = "blitz"
	app.Usage = "blitz/blizzard control utility"
	app.EnableShellCompletion = true
	app.Main.Commands = []cli.Command{{
		Name:      "deploy",
		Usage:     "Registers a worker with blizzard",
		ShortName: "d",
		Before:    func(*cli.Context) error { return c.Connect() },
		Action:    c.Deploy,
		Options: []cli.Option{
			cli.StringOption{
				Name:       "worker",
				Usage:      "native worker binary",
				Completion: cli.StdCompletion,
			},
			cli.StringOption{
				Name:       "adapter",
				Usage:      "foreign worker adapter type",
				ValueList:  []string{"thin"},
				Completion: cli.ValueListCompletion,
			},
			cli.StringOption{
				Name:       "config",
				Usage:      "adapter-specific config file",
				Completion: cli.StdCompletion,
			},
		},
	}, {
		Name: "list",
		Commands: []cli.Command{{
			Name:   "apps",
			Action: c.ListApps,
		}},
	}, {
		Name: "restart",
		Commands: []cli.Command{{
			Name:   "takeover",
			Action: c.Takeover,
			Options: []cli.Option{
				cli.StringOption{Name: "app"},
				//cli.BoolOption{Name: "kill"},
			},
		}},
	}, {
		Name:   "proc-stats",
		Action: c.ProcStats,
	}, {
		Name: "config",
		Commands: []cli.Command{{
			Name:   "log",
			Usage:  "Sets logging options.\nExactly one of 'stderr' or 'syslog' must be provided.",
			Action: c.ConfigLog,
			Options: []cli.Option{
				cli.BoolOption{Name: "stderr"},
				cli.BoolOption{Name: "syslog"},
				cli.StringOption{
					Name:  "severity",
					Usage: "syslog severity",
					Value: "info",
					ValueList: []string{
						"emerg", "alert", "crit", "err", "warning",
						"notice", "info", "debug"},
					Completion: cli.ValueListCompletion,
				},
				cli.StringOption{
					Name:  "facility",
					Usage: "syslog facility",
					Value: "local7",
					ValueList: []string{
						"kern", "user", "mail", "daemon", "auth",
						"syslog", "lpr", "news", "uucp",
						"cron", "authpriv", "ftp",
						"local0", "local1", "local2", "local3",
						"local4", "local5", "local6", "local7"},
					Completion: cli.ValueListCompletion,
				},
			},
		}},
	},
	}

	err := app.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func (c *Cli) Connect() (err error) {
	c.conn, err = net.Dial("unix", "blitz/ctl")
	if err != nil {
		return
	}
	err = c.send(blitz.ConnectionTypeCommand{blitz.Command{"connection-type"}, "cli"})
	return
}

func (c *Cli) GetResponse(out interface{}) error {
	decoder := json.NewDecoder(c.conn)
	return decoder.Decode(out)
}

func (c *Cli) Deploy(ctx *cli.Context) error {
	worker := ctx.String("worker")
	adapter := ctx.String("adapter")
	config := ctx.String("config")
	if worker == "" && adapter == "" {
		return fmt.Errorf("neither 'worker' not 'adapter' were specified, exactly one is required")
	}
	if worker != "" && adapter != "" {
		return fmt.Errorf("both 'worker' and 'adapter' were specified, exactly one is required")
	}
	if adapter != "" && config == "" {
		return fmt.Errorf("'adapter' requires 'config'")
	}
	cmd := blitz.DeployCommand{
		Command:     blitz.Command{Type: blitz.CmdDeploy},
		Application: worker,
		Adapter:     adapter,
		Config:      config,
	}
	var resp blitz.Response
	err := c.callBlizzard(cmd, &resp)
	if err != nil {
		return err
	}
	return resp.Err()
}

func (c *Cli) ListApps(ctx *cli.Context) error {
	cmd := blitz.ListExecutablesCommand{blitz.Command{Type: blitz.CmdListApps}}
	var resp blitz.ListExecutablesResponse
	err := c.callBlizzard(cmd, &resp)
	if err != nil {
		return err
	}
	if resp.Err() != nil {
		return resp.Err()
	}
	for _, app := range resp.Executables {
		fmt.Println(app)
	}
	return nil
}

func (c *Cli) ProcStats(ctx *cli.Context) error {
	cmd := blitz.ProcStatCommand{blitz.Command{Type: blitz.CmdProcStats}}
	var resp blitz.ProcStatResponse
	err := c.callBlizzard(cmd, &resp)
	if err != nil {
		return err
	}
	if resp.Err() != nil {
		return resp.Err()
	}
	fmt.Printf("apps %d\nproc-groups %d\nprocs %d\n", resp.Apps, resp.ProcGroups, resp.Procs)
	return nil
}

func (c *Cli) Takeover(ctx *cli.Context) error {
	app := ctx.String("app")
	if app == "" {
		return fmt.Errorf("restart requires --app")
	}
	cmd := blitz.RestartTakeoverCommand{blitz.Command{Type: blitz.CmdTakeover}, app, ctx.Bool("kill")}
	resp := blitz.RestartTakeoverResponse{}
	err := c.callBlizzard(cmd, &resp)
	if err != nil {
		return err
	}
	return resp.Err()
}

func (c *Cli) ConfigLog(ctx *cli.Context) error {
	isStderr := ctx.Bool("stderr")
	isSyslog := ctx.Bool("syslog")
	if (isSyslog && isStderr) || (!isSyslog && !isStderr) {
		return fmt.Errorf("Exactly one of --stderr or --syslog must be given.")
	}
	if isSyslog {
		if ctx.String("facility") == "" {
			return fmt.Errorf("--syslog requires --facility")
		}
		if ctx.String("severity") == "" {
			return fmt.Errorf("--syslog requires --severity")
		}
	}
	command := blitz.ConfigLoggerCommand{}
	command.Command.Type = blitz.CmdConfigLogger
	if isStderr {
		command.LogType = "stderr"
	} else {
		command.LogType = "syslog"
	}
	command.Syslog.Facility = ctx.String("facility")
	command.Syslog.Severity = ctx.String("severity")
	resp := blitz.ConfigLoggerResponse{}
	err := c.callBlizzard(command, &resp)
	if err != nil {
		return err
	}
	return resp.Err()
}

func (c *Cli) callBlizzard(cmd interface{}, resp interface{}) error {
	err := c.send(cmd)
	if err != nil {
		return err
	}
	return c.GetResponse(resp)
}

func (c *Cli) send(data interface{}) error {
	encoder := json.NewEncoder(c.conn)
	return encoder.Encode(data)
}
