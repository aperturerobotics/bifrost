//go:build js && !bldr_indexeddb

package browser_storage

import (
	"testing"

	volume_opfs "github.com/s4wave/spacewave/db/volume/js/opfs"
)

// TestOpfsStorageBuildsV3RuntimeConfig checks the product format and lock namespace.
func TestOpfsStorageBuildsV3RuntimeConfig(t *testing.T) {
	conf, err := NewOpfsStorage("prefix/").BuildVolumeConfig("state", nil)
	if err != nil {
		t.Fatal(err)
	}

	opfsConf, ok := conf.(*volume_opfs.Config)
	if !ok {
		t.Fatalf("BuildVolumeConfig returned %T, want *volume_opfs.Config", conf)
	}
	if got, want := opfsConf.GetRootPath(), "prefix/state"; got != want {
		t.Fatalf("RootPath = %q, want %q", got, want)
	}
	if got, want := opfsConf.GetLockPrefix(), opfsConf.GetRootPath(); got != want {
		t.Fatalf("LockPrefix = %q, want %q", got, want)
	}
	if got, want := opfsConf.GetStorageFormatVersion(), uint32(3); got != want {
		t.Fatalf("StorageFormatVersion = %d, want %d", got, want)
	}
	if got, want := opfsConf.GetDriverMode(), "auto"; got != want {
		t.Fatalf("DriverMode = %q, want %q", got, want)
	}
}
