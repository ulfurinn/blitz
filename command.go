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

func (c *Cli) Connect() error {
	conn, err := net.Dial("unix", "blitz/ctl")
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *Cli) GetResponse() {
	decoder := json.NewDecoder(c.conn)
	v := Response{}
	err := decoder.Decode(&v)
	if err != nil {
		fatal(err)
	}
	if v.Error != "" {
		fatal(fmt.Errorf(v.Error))
	}
}

func (c *Cli) Deploy(ctx *cli.Context) {
	args := ctx.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "file name required")
		return
	}
	binary := args[0]
	cmd := Command{
		Type:   "deploy",
		Binary: binary,
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
