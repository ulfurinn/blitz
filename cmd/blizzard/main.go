package main

import (
	"runtime"

	"bitbucket.org/ulfurinn/blitz/blizzard"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	m := blizzard.NewBlizzard()
	m.Start()
}
