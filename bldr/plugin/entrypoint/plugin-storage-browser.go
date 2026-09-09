//go:build js && !bldr_cloudflare

package plugin_entrypoint

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	plugin_host_storage "github.com/s4wave/spacewave/bldr/plugin/host/storage"
	"github.com/s4wave/spacewave/bldr/storage"
	browser_storage "github.com/s4wave/spacewave/bldr/storage/browser"
)

// buildPluginStorages builds the storage backends for the plugin.
//
// When the host selected a Storage ID for this plugin instance, every volume
// request proxies through the plugin host to that Storage, matching the native
// path. Otherwise js builds use direct OPFS access instead of proxying through
// the host.
func buildPluginStorages(
	b bus.Bus,
	sr *static.Resolver,
	hostStorageID string,
) []storage.Storage {
	if hostStorageID != "" {
		hostStorage := plugin_host_storage.NewPluginHostStorage(hostStorageID)
		hostStorage.AddFactories(b, sr)
		return []storage.Storage{hostStorage}
	}

	storages := browser_storage.BuildStorage(b, "")
	for _, st := range storages {
		st.AddFactories(b, sr)
	}
	return storages
}
