//go:build windows

package launcher_install

import (
	"github.com/pkg/errors"
	"golang.org/x/sys/windows/registry"
)

// uninstallKey is the per-user registry key that lists Spacewave in
// Settings > Apps.
const uninstallKey = `Software\Microsoft\Windows\CurrentVersion\Uninstall\Spacewave`

// writeUninstallEntry lists the install in Settings > Apps, with an uninstall
// command that runs the launcher's uninstall subcommand.
func writeUninstallEntry(dir, launcher string) error {
	// Create or open the per-user uninstall key.
	key, _, err := registry.CreateKey(registry.CURRENT_USER, uninstallKey, registry.SET_VALUE)
	if err != nil {
		return errors.Wrap(err, "create uninstall registry key")
	}
	defer key.Close()

	// Describe the app, its icon and its uninstall command.
	values := []struct {
		name, value string
	}{
		{"DisplayName", "Spacewave"},
		{"DisplayIcon", launcher + ",0"},
		{"Publisher", "Aperture Robotics"},
		{"InstallLocation", dir},
		{"UninstallString", `"` + launcher + `" uninstall`},
	}
	for _, value := range values {
		if err := key.SetStringValue(value.name, value.value); err != nil {
			return errors.Wrapf(err, "write uninstall %s", value.name)
		}
	}

	// Offer only removal, with no modify or repair action.
	for _, name := range []string{"NoModify", "NoRepair"} {
		if err := key.SetDWordValue(name, 1); err != nil {
			return errors.Wrapf(err, "write uninstall %s", name)
		}
	}
	return nil
}
