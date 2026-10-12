package plugin_space

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/starpc/srpc"
	bldr_manifest "github.com/s4wave/spacewave/bldr/manifest"
	bldr_plugin "github.com/s4wave/spacewave/bldr/plugin"
)

// manifestPluginHost loads one plugin once its manifest resolves. Like the
// plugin host awaiting initial capability registration, the load stays busy
// until the plugin is ready, even when FetchManifest is idle without a manifest.
type manifestPluginHost struct {
	// b is the bus FetchManifest resolves on.
	b bus.Bus
	// client is the running plugin's RPC client.
	client srpc.Client
}

// GetControllerInfo returns information about the controller.
func (h *manifestPluginHost) GetControllerInfo() *controller.Info {
	return controller.NewInfo(
		"test/manifest-plugin-host",
		controller.MustParseVersion("0.0.1"),
		"loads a test plugin once its manifest resolves",
	)
}

// Execute waits for the controller context to end.
func (h *manifestPluginHost) Execute(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

// Close releases any resources used by the controller.
func (h *manifestPluginHost) Close() error { return nil }

// HandleDirective resolves LoadPlugin for the test plugin from its manifest.
func (h *manifestPluginHost) HandleDirective(
	_ context.Context,
	inst directive.Instance,
) ([]directive.Resolver, error) {
	// Accept only the plugin whose manifest this host can load.
	load, ok := inst.GetDirective().(bldr_plugin.LoadPlugin)
	if !ok || load.LoadPluginID() != spacePluginID {
		return nil, nil
	}

	// Publish the running plugin once its manifest becomes available.
	return directive.R(directive.NewFuncResolver(func(ctx context.Context, handler directive.ResolverHandler) error {
		// Follow the plugin manifest until the load ends.
		fetch, fetchRef, err := h.b.AddDirective(
			bldr_manifest.NewFetchManifest(spacePluginID, nil, []string{spacePlatformID}, 0),
			nil,
		)
		if err != nil {
			return err
		}
		defer fetchRef.Release()

		// Keep the load busy until a manifest can run the plugin.
		loaded := false
		releaseState := fetch.AddStateCallback(func(_ bool, _ []error, vals []directive.AttachedValue) {
			for _, val := range vals {
				fetched, ok := val.GetValue().(*bldr_manifest.FetchManifestValue)
				if ok && len(fetched.GetManifestRefs()) != 0 && !loaded {
					loaded = true
					_, _ = handler.AddValue(bldr_plugin.NewRunningPlugin(h.client))
				}
			}
			handler.MarkIdle(loaded)
		})
		defer releaseState()

		// Retain the manifest watch through the host's lifetime.
		<-ctx.Done()
		return ctx.Err()
	}), nil)
}

// _ verifies the controller contract.
var _ controller.Controller = (*manifestPluginHost)(nil)
