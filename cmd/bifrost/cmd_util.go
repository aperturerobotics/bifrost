package main

import (
	util "github.com/aperturerobotics/bifrost/cli/util"
	"github.com/urfave/cli"
)

var utilArgs util.UtilArgs

func init() {
	utilCommands := (&utilArgs).BuildCommands()
	commands = append(
		commands,
		cli.Command{
			Name:        "util",
			Usage:       "utility sub-commands",
			Subcommands: utilCommands,
			Flags:       (&utilArgs).BuildFlags(),
		},
	)
}
