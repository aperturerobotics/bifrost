//go:build js

package browser_storage

import (
	"os"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/s4wave/spacewave/bldr/storage"
)

// StorageModeEnv selects the browser storage backend after the browser probes
// its actual capabilities.
const StorageModeEnv = "BLDR_BROWSER_STORAGE"

// BuildStorage builds all available storage methods.
//
// prefix is used as the IndexedDB prefix in the browser.
func BuildStorage(_ bus.Bus, prefix string) []storage.Storage {
	if os.Getenv(StorageModeEnv) == "indexeddb" {
		return []storage.Storage{NewIndexedDB(prefix, false)}
	}
	return []storage.Storage{NewOpfsStorage(prefix)}
}
