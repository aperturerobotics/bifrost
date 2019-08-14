package main

import (
	bcli "github.com/aperturerobotics/bifrost/cli"
	"github.com/urfave/cli"
)

// cliArgs are the client arguments
var cliArgs bcli.ClientArgs

func init() {
	clientCommands := cliArgs.BuildCommands()
	clientFlags := cliArgs.BuildFlags()
	commands = append(
		commands,
		cli.Command{
			Name:        "client",
			Usage:       "client sub-commands",
			Subcommands: clientCommands,
			Flags:       clientFlags,
		},
	)
}
