//go:build windows

package launcher_install

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sys/windows"
)

// helperName is the file name of the helper beside the installed launcher.
const helperName = "spacewave-helper.exe"

// Installer places the desktop launcher in the current user's Programs folder.
type Installer struct {
	// dir is the install directory, Programs\Spacewave.
	dir string
	// launcher is the installed launcher path.
	launcher string
	// shortcut is the Start menu shortcut path.
	shortcut string
}

// NewInstaller resolves the per-user install locations.
func NewInstaller() (*Installer, error) {
	// Resolve the per-user Programs and Start menu folders.
	programs, err := windows.KnownFolderPath(windows.FOLDERID_UserProgramFiles, windows.KF_FLAG_CREATE)
	if err != nil {
		return nil, errors.Wrap(err, "resolve user Programs folder")
	}
	startMenu, err := windows.KnownFolderPath(windows.FOLDERID_Programs, windows.KF_FLAG_CREATE)
	if err != nil {
		return nil, errors.Wrap(err, "resolve Start menu folder")
	}

	// Place the launcher and its helper together under Programs\Spacewave.
	dir := filepath.Join(programs, "Spacewave")
	return &Installer{
		dir:      dir,
		launcher: filepath.Join(dir, "Spacewave.exe"),
		shortcut: filepath.Join(startMenu, "Spacewave.lnk"),
	}, nil
}

// Dir returns the install directory.
func (i *Installer) Dir() string {
	return i.dir
}

// Launcher returns the installed launcher path.
func (i *Installer) Launcher() string {
	return i.launcher
}

// Installed reports whether a complete install is present.
func (i *Installer) Installed() bool {
	info, err := os.Stat(i.launcher)
	return err == nil && info.Mode().IsRegular()
}

// Runs reports whether executable is the installed launcher. Windows paths
// ignore case.
func (i *Installer) Runs(executable string) bool {
	return strings.EqualFold(filepath.Clean(executable), i.launcher)
}

// InstallHelper publishes helper beside the launcher, keeping a helper that is
// already installed.
func (i *Installer) InstallHelper(helper []byte) error {
	if err := os.MkdirAll(i.dir, 0o700); err != nil {
		return errors.Wrap(err, "create install directory")
	}
	return publish(filepath.Join(i.dir, helperName), bytes.NewReader(helper))
}

// InstallLauncher writes the Start menu and Settings > Apps entries, then
// publishes source as the launcher. Call InstallHelper first.
func (i *Installer) InstallLauncher(source string) error {
	// Register the install before the launcher marks it complete.
	if err := writeShortcut(i.shortcut, i.launcher, "Spacewave"); err != nil {
		return err
	}
	if err := writeUninstallEntry(i.dir, i.launcher); err != nil {
		return err
	}

	// Publish the launcher from the running executable.
	input, err := os.Open(source)
	if err != nil {
		return errors.Wrap(err, "open launcher source")
	}
	defer input.Close()
	return publish(i.launcher, input)
}

// Start starts the installed launcher detached from this process.
func (i *Installer) Start() error {
	cmd := exec.Command(i.launcher)
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "start installed Spacewave")
	}
	return cmd.Process.Release()
}

// publish stages src beside path and links it to path only when path is free.
// A path published by a concurrent launch is kept.
func publish(path string, src io.Reader) (retErr error) {
	// Stage the contents, removing the staged name once linked.
	staged, err := stage(path, src)
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(staged); err != nil && retErr == nil {
			retErr = errors.Wrap(err, "remove temporary install file")
		}
	}()

	// A hard link publishes the complete file only if the name is free.
	if err := os.Link(staged, path); err != nil && !errors.Is(err, os.ErrExist) {
		return errors.Wrapf(err, "publish %s", filepath.Base(path))
	}
	return nil
}

// replace stages src beside path and renames it over path.
func replace(path string, src io.Reader) error {
	staged, err := stage(path, src)
	if err != nil {
		return err
	}
	if err := os.Rename(staged, path); err != nil {
		_ = os.Remove(staged)
		return errors.Wrapf(err, "replace %s", filepath.Base(path))
	}
	return nil
}

// stage writes src to a durable temporary file beside path and returns its
// name.
func stage(path string, src io.Reader) (string, error) {
	// Write the contents to a private temporary file.
	output, err := os.CreateTemp(filepath.Dir(path), ".install-*")
	if err != nil {
		return "", errors.Wrap(err, "create temporary install file")
	}
	if _, err := io.Copy(output, src); err != nil {
		_ = output.Close()
		_ = os.Remove(output.Name())
		return "", errors.Wrapf(err, "write %s", filepath.Base(path))
	}

	// Make the contents durable before the final name appears.
	if err := output.Sync(); err != nil {
		_ = output.Close()
		_ = os.Remove(output.Name())
		return "", errors.Wrapf(err, "sync %s", filepath.Base(path))
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(output.Name())
		return "", errors.Wrapf(err, "close %s", filepath.Base(path))
	}
	return output.Name(), nil
}
