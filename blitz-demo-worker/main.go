package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

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
	flag.Parse()
	fmt.Println(patch)
	i, err := strconv.ParseInt(patch, 10, 0)
	if err != nil {
		fatal(err)
	}
	w := blitz.Worker{Patch: i}
	w.Init()
	err = w.Listen()
	if err != nil {
		fatal(err)
	}
	err = w.Connect()
	if err != nil {
		fatal(err)
	}
	err = w.Announce([]blitz.PathSpec{{Path: "/sleep", Version: 1}})
	if err != nil {
		fatal(err)
	}
	w.Serve(api())
}

func api() http.Handler {
	m := martini.Classic()
	m.Get("/sleep", demo)
	return m
}

func demo() (int, string) {
	return 200, "blitz!"
}
