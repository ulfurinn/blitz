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
	app.Commands = []cli.Command{cli.Command{
		Name:      "deploy",
		ShortName: "d",
		Action:    c.Deploy,
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
		fatal(v.Error)
	}
}

func (c *Cli) Deploy(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "file name required")
		return
	}
	executable := args[0]
	cmd := DeployCommand{
		Command:    Command{Type: "deploy"},
		Executable: executable,
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
