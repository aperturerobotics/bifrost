//go:build windows

package spacewave_cli

import (
	"os"
	"path/filepath"

	"github.com/aperturerobotics/cli"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/bldr/entrypoint/storagepath"
	launcher_install "github.com/s4wave/spacewave/core/launcher/install"
)

// appendInstallCommands adds the commands the installed app runs on itself:
// uninstall from Settings > Apps and applying a staged update.
func appendInstallCommands(commands []*cli.Command) []*cli.Command {
	var pid uint
	return append(
		commands,
		&cli.Command{
			Name:   "uninstall",
			Usage:  "remove the installed Spacewave app",
			Hidden: true,
			Action: func(c *cli.Context) error {
				err := runUninstall(c)
				if err != nil {
					launcher_install.ShowError("Spacewave was not uninstalled", err)
				}
				return err
			},
		},
		&cli.Command{
			Name:   "apply-update",
			Usage:  "replace the installed Spacewave app with this executable",
			Hidden: true,
			Flags: []cli.Flag{
				&cli.UintFlag{
					Name:        "pid",
					Usage:       "process ID of the desktop to wait for",
					Destination: &pid,
				},
			},
			Action: func(*cli.Context) error {
				err := runApplyUpdate(uint32(pid))
				if err != nil {
					launcher_install.ShowError("Spacewave was not updated", err)
				}
				return err
			},
		},
	)
}

// runUninstall confirms the uninstall, stops the desktop's daemon, and removes
// the installed app, and the local Spaces when the person chose to.
func runUninstall(c *cli.Context) error {
	// Ask before removing anything.
	confirmed, deleteData, err := launcher_install.ConfirmUninstall()
	if err != nil || !confirmed {
		return err
	}

	// Stop the daemon serving the desktop's state root.
	stateRoot, err := storagepath.DetermineStorageRoot("spacewave")
	if err != nil {
		return err
	}
	if err := runStop(c.Context, filepath.Join(stateRoot, socketName)); err != nil {
		return err
	}

	// Remove the install, and the state root when asked.
	installer, err := launcher_install.NewInstaller()
	if err != nil {
		return err
	}
	return installer.Uninstall(stateRoot, deleteData)
}

// runApplyUpdate replaces the installed app with this executable and its
// embedded helper once the desktop process pid has exited.
func runApplyUpdate(pid uint32) error {
	// Read this executable and the helper it carries.
	helper := launcher_install.EmbeddedHelper()
	if helper == nil {
		return errors.New("this Spacewave build carries no helper")
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}

	// Replace the install and start it.
	installer, err := launcher_install.NewInstaller()
	if err != nil {
		return err
	}
	return installer.Update(executable, helper, pid)
}
