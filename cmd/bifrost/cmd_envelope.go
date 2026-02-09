package main

import (
	bcli "github.com/aperturerobotics/bifrost/cli"
	"github.com/aperturerobotics/cli"
)

// envelopeArgs are the envelope command arguments.
var envelopeArgs bcli.EnvelopeArgs

func init() {
	commands = append(commands, &cli.Command{
		Name:        "envelope",
		Usage:       "seal and unseal secret-sharing envelopes",
		Subcommands: envelopeArgs.BuildCommands(),
	})
}
