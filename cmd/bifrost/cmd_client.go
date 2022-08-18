package main

import (
	bcli "github.com/aperturerobotics/bifrost/cli"
	"github.com/urfave/cli/v2"
)

// cliArgs are the client arguments
var cliArgs bcli.ClientArgs

func init() {
	clientCommands := (&cliArgs).BuildCommands()
	clientFlags := (&cliArgs).BuildFlags()
	cbusCmd := (&cliArgs.CbusConf).BuildControllerBusCommand()
	cbusCmd.Before = func(_ *cli.Context) error {
		client, err := (&cliArgs).BuildClient()
		if err != nil {
			return err
		}
		(&cliArgs.CbusConf).SetClient(client)
		return nil
	}
	clientCommands = append(clientCommands, cbusCmd)
	commands = append(
		commands,
		&cli.Command{
			Name:        "client",
			Usage:       "client sub-commands",
			Subcommands: clientCommands,
			Flags:       clientFlags,
		},
	)
}
