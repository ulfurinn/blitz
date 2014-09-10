package blitz

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/codegangsta/cli"
)

type Commander struct {
}

type Cli struct {
	conn net.Conn
}

func (c *Cli) Run() {
	err := c.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	app := cli.NewApp()
	app.Name = "blitz"
	app.Usage = "blitz/blizzard control utility"
	app.EnableBashCompletion = true
	app.Commands = []cli.Command{{
		Name:      "deploy",
		Usage:     "Registers a worker with blizzard",
		ShortName: "d",
		Action:    c.Deploy,
		Flags: []cli.Flag{
			cli.StringFlag{Name: "worker", Usage: "native worker binary"},
			cli.StringFlag{Name: "adapter", Usage: "foreign worker adapter type"},
			cli.StringFlag{Name: "config", Usage: "adapter-specific config file"},
		},
	}, {
		Name: "list",
		Subcommands: []cli.Command{{
			Name:   "apps",
			Action: c.ListApps,
		}},
	}, {
		Name: "restart",
		Subcommands: []cli.Command{{
			Name:   "takeover",
			Action: c.Takeover,
			Flags:  []cli.Flag{cli.StringFlag{Name: "app"}},
		}},
	}}

	app.Run(os.Args)
}

func (c *Cli) Connect() (err error) {
	c.conn, err = net.Dial("unix", "blitz/ctl")
	if err != nil {
		return
	}
	err = c.send(ConnectionTypeCommand{Command{"connection-type"}, "cli"})
	return
}

func (c *Cli) GetResponse(out interface{}) error {
	decoder := json.NewDecoder(c.conn)
	return decoder.Decode(out)
}

func (c *Cli) Deploy(ctx *cli.Context) {
	worker := ctx.String("worker")
	adapter := ctx.String("adapter")
	config := ctx.String("config")
	if worker == "" && adapter == "" {
		fatal(fmt.Errorf("neither 'worker' not 'adapter' were specified, exactly one is required"))
	}
	if worker != "" && adapter != "" {
		fatal(fmt.Errorf("both 'worker' and 'adapter' were specified, exactly one is required"))
	}
	if adapter != "" && config == "" {
		fatal(fmt.Errorf("'adapter' requires 'config'"))
	}
	cmd := DeployCommand{
		Command:    Command{Type: "deploy"},
		Executable: worker,
		Adapter:    adapter,
		Config:     config,
	}
	err := c.send(cmd)
	if err != nil {
		fatal(err)
	}
	var resp Response
	err = c.GetResponse(&resp)
	if err != nil {
		fatal(err)
	}
	if resp.Error != nil {
		fatal(fmt.Errorf(*resp.Error))
	}
}

func (c *Cli) ListApps(ctx *cli.Context) {
	cmd := ListExecutablesCommand{Command{Type: "list-apps"}}
	err := c.send(cmd)
	if err != nil {
		fatal(err)
	}
	var resp ListExecutablesResponse
	err = c.GetResponse(&resp)
	if err != nil {
		fatal(err)
	}
	if resp.Error != nil {
		fatal(fmt.Errorf(*resp.Error))
	}
	for _, app := range resp.Executables {
		fmt.Println(app)
	}
}

func (c *Cli) Takeover(ctx *cli.Context) {
	app := ctx.String("app")
	if app == "" {
		fatal(fmt.Errorf("restart requires an app name"))
	}
	cmd := RestartTakeoverCommand{Command{Type: "restart-takeover"}, app}
	resp := RestartTakeoverResponse{}
	err := c.callBlizzard(cmd, &resp)
	if err != nil {
		fatal(err)
	}
	if err = resp.Err(); err != nil {
		fatal(err)
	}
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
