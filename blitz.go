package blitz

import (
	"fmt"
	"os"
)

type PathSpec struct {
	Path    string `json:"path"`
	Version int    `json:"version"`
}

type Command struct {
	Type    string     `json:"type"`
	Tag     string     `json:"tag"`
	PID     int        `json:"pid"`
	ProcID  string     `json:"procid"`
	Patch   int64      `json:"patch"`
	Paths   []PathSpec `json:"paths"`
	Binary  string     `json:"binary"`
	Network string     `json:"network"`
	Address string     `json:"address"`
}

type Response struct {
	Error string `json:"error"`
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
