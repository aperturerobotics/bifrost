package optypes

import (
	"testing"

	manifest_world "github.com/s4wave/spacewave/bldr/manifest/world"
)

// TestBuildSpaceLookupOpResolvesPluginBuildPublication covers the operations
// decoded by a remote build worker's mounted Space capability.
func TestBuildSpaceLookupOpResolvesPluginBuildPublication(t *testing.T) {
	requireLookupWorldAndBuildSpaceOp[*manifest_world.StoreManifestOp](t, manifest_world.StoreManifestOpTypeID)
	requireLookupWorldAndBuildSpaceOp[*manifest_world.ExtractManifestBundleOp](t, manifest_world.ExtractManifestBundleOpTypeID)
}
