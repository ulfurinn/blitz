package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/martini"

	"bitbucket.org/ulfurinn/blitz"
)

// use ldflags to set
var patch string

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	fmt.Println("worker boot")
	blitz.InitWorker()
	fmt.Println(patch)
	i, err := strconv.ParseInt(patch, 10, 0)
	if err != nil {
		fatal(err)
	}
	w := blitz.Worker{
		Patch:   i,
		Handler: api(),
		Paths: []blitz.PathSpec{
			{Path: "/sleep", Version: 1},
			{Path: "/ok", Version: 1},
		},
		Bootstrap: bootstrapper,
	}
	err = w.Run()
	if err != nil {
		fatal(err)
	}
}

func bootstrapper(cmd *blitz.BootstrapCommand) error {
	cmd.Instances = 1
	cmd.AppName = "demo-worker"
	return nil
}

func api() http.Handler {
	m := martini.Classic()
	m.Get("/ok", ok)
	m.Get("/sleep", sleep)
	return m
}

func ok() (int, string) {
	return 200, "blitz!"
}

func sleep() (int, string) {
	time.Sleep(10 * time.Second)
	return 200, "blitz!"
}
