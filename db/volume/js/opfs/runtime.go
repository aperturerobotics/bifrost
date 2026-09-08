//go:build js

package volume_opfs

import (
	"bytes"
	"context"
	"syscall/js"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/opfs"
	"github.com/s4wave/spacewave/db/unixfs"
	"github.com/s4wave/spacewave/db/volume"
	"github.com/sirupsen/logrus"
)

const (
	// currentStorageFormatVersion identifies the clean immutable volume format.
	currentStorageFormatVersion uint32 = 3
	// formatMarkerName distinguishes initialized roots from preexisting data.
	formatMarkerName = ".spacewave-opfs-format"
	// formatMarker is fixed framing, not a second mutable storage descriptor.
	formatMarker = "spacewave-opfs-volume/3\n"
	// driverModeAuto selects the current runtime's supported browser driver.
	driverModeAuto = "auto"
	// driverModeStandardWasm selects standard Go's browser driver ABI.
	driverModeStandardWasm = "standard-wasm"
	// driverModeTinyGo selects TinyGo's browser driver ABI.
	driverModeTinyGo = "tinygo"
)

// runtimeDriverMode resolves the driver selection for the current runtime.
func runtimeDriverMode(conf *Config) string {
	if mode := conf.GetDriverMode(); mode != "" {
		return mode
	}
	return driverModeAuto
}

// openRuntimeRoot initializes only an absent or empty directory.
// Existing data with an unknown marker requires an explicit external transition.
func openRuntimeRoot(ctx context.Context, le *logrus.Entry, root js.Value, conf *Config) (js.Value, error) {
	parts, _ := unixfs.SplitPath(conf.GetRootPath())
	if len(parts) == 0 {
		return js.Undefined(), errors.New("root_path must name an OPFS directory")
	}
	prefix := conf.GetLockPrefix()
	if prefix == "" {
		prefix = conf.GetRootPath()
	}
	lock, err := opfs.DefaultDriver.AcquireWebLock(ctx, prefix+"/engine/publish", true)
	if err != nil {
		return js.Undefined(), err
	}
	if lock.Outcome != opfs.WebLockOutcomeAcquired {
		return js.Undefined(), errors.New("OPFS format lock was not acquired")
	}
	defer lock.Release()
	dir, err := opfs.GetDirectoryPath(root, parts, true)
	if err != nil {
		return js.Undefined(), err
	}
	marker, err := opfs.ReadFile(dir, formatMarkerName)
	emptyMarker := err == nil && len(marker) == 0
	if err == nil {
		if bytes.Equal(marker, []byte(formatMarker)) {
			return dir, nil
		}
		if len(marker) != 0 {
			return js.Undefined(), incompatibleFormat()
		}
	}
	if err != nil && !opfs.IsNotFound(err) {
		return js.Undefined(), err
	}

	// Listing is confined to first initialization, never ordinary compatible open.
	names, err := opfs.ListDirectory(dir)
	if err != nil {
		return js.Undefined(), err
	}
	// An interrupted first marker write may leave its newly created empty entry.
	if len(names) != 0 && !(len(names) == 1 && names[0] == formatMarkerName && emptyMarker) {
		return js.Undefined(), incompatibleFormat()
	}
	if err := opfs.WriteFile(dir, formatMarkerName, []byte(formatMarker)); err != nil {
		return js.Undefined(), err
	}
	if le != nil {
		le.WithField("root_path", conf.GetRootPath()).Info("initialized immutable OPFS volume")
	}
	return dir, nil
}

// incompatibleFormat preserves saved bytes and stops a futile restart loop.
func incompatibleFormat() error {
	return volume.Permanent(errors.New("incompatible OPFS volume format: use a new empty volume or explicitly export and import the existing data"))
}

// deleteRuntimeRoot performs the Volume interface's explicit destructive operation.
func deleteRuntimeRoot(root js.Value, rootPath string) error {
	parts, _ := unixfs.SplitPath(rootPath)
	if len(parts) == 0 {
		return errors.New("root_path must name an OPFS directory")
	}
	parent := root
	for _, part := range parts[:len(parts)-1] {
		next, err := opfs.GetDirectory(parent, part, false)
		if err != nil {
			if opfs.IsNotFound(err) {
				return nil
			}
			return err
		}
		parent = next
	}
	err := opfs.DeleteEntry(parent, parts[len(parts)-1], true)
	if err != nil && !opfs.IsNotFound(err) {
		return err
	}
	return nil
}
