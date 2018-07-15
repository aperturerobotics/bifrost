package main

import (
	"github.com/urfave/cli"
)

func init() {
	commands = append(
		commands,
		cli.Command{
			Name:  "node",
			Usage: "run a bifrost node",
		},
	)
}
