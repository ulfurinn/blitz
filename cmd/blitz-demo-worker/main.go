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
	i, err := strconv.ParseUint(patch, 10, 0)
	if err != nil {
		fatal(err)
	}
	w := blitz.Worker{
		AppName: "demo-worker",
		Patch:   i,
		Paths: []blitz.PathSpec{
			{Path: "/sleep", Version: 1},
			{Path: "/ok", Version: 1},
		},
		Run:     blitz.StandardRunner,
		Handler: api(),
	}
	err = w.Start()
	if err != nil {
		fatal(err)
	}
	w.Wait()
}

func api() http.Handler {
	m := martini.Classic()
	m.Get("/ok", ok)
	m.Get("/sleep", sleep)
	return m
}

func ok() (int, string) {
	return 200, "blitz!\n"
}

func sleep() (int, string) {
	time.Sleep(10 * time.Second)
	return 200, "blitz!\n"
}
