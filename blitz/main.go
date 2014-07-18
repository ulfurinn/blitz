package main

import "bitbucket.org/ulfurinn/blitz"

func main() {
	blitz.InitTool()
	cli := &blitz.Cli{}
	cli.Run()
}
