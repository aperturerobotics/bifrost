//go:build !js && !windows

package spacewave_cli

import "github.com/aperturerobotics/cli"

// appendInstallCommands adds nothing; only the Windows app installs itself.
func appendInstallCommands(commands []*cli.Command) []*cli.Command {
	return commands
}
