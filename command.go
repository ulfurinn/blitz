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
	app.Commands = []cli.Command{cli.Command{
		Name:      "deploy",
		Usage:     "Registers a worker with blizzard",
		ShortName: "d",
		Action:    c.Deploy,
		Flags: []cli.Flag{
			cli.StringFlag{Name: "worker", Usage: "native worker binary"},
			cli.StringFlag{Name: "adapter", Usage: "foreign worker adapter type"},
			cli.StringFlag{Name: "config", Usage: "adapter-specific config file"},
		},
	}}
	app.Run(os.Args)
}

func (c *Cli) Connect() (err error) {
	c.conn, err = net.Dial("unix", "blitz/ctl")
	return
}

func (c *Cli) GetResponse() {
	decoder := json.NewDecoder(c.conn)
	v := Response{}
	err := decoder.Decode(&v)
	if err != nil {
		fatal(err)
	}
	if v.Error != nil {
		fatal(fmt.Errorf(*v.Error))
	}
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
	c.GetResponse()
}

func (c *Cli) send(data interface{}) error {
	encoder := json.NewEncoder(c.conn)
	return encoder.Encode(data)
}
