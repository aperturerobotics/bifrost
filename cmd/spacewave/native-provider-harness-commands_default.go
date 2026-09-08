//go:build !native_provider_harness && !js

package main

import (
	aperture_cli "github.com/aperturerobotics/cli"
	cli_entrypoint "github.com/s4wave/spacewave/bldr/cli/entrypoint"
)

func addNativeProviderHarnessCommands(
	commands []*aperture_cli.Command,
	_ func() cli_entrypoint.CliBus,
) []*aperture_cli.Command {
	return commands
}
