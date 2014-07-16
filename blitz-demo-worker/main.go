package main

import (
	"flag"
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
	flag.Parse()
	fmt.Println(patch)
	i, err := strconv.ParseInt(patch, 10, 0)
	if err != nil {
		fatal(err)
	}
	w := blitz.Worker{
		Patch:   i,
		Handler: api(),
		Paths:   []blitz.PathSpec{{Path: "/sleep", Version: 1}},
	}
	err = w.Run()
}

func api() http.Handler {
	m := martini.Classic()
	m.Get("/sleep", demo)
	return m
}

func demo() (int, string) {
	time.Sleep(10 * time.Second)
	return 200, "blitz!"
}
