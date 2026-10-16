//go:build !tinygo

package optypes

import (
	"slices"

	bldr_manifest_builder "github.com/s4wave/spacewave/bldr/manifest/builder"
	manifest_world "github.com/s4wave/spacewave/bldr/manifest/world"
	s4wave_apt "github.com/s4wave/spacewave/sdk/apt"
)

// buildTypes lists the types of the native and GoScript builds.
var buildTypes = slices.Concat(
	manifest_world.WorldTypes,
	bldr_manifest_builder.BuilderTypes,
	s4wave_apt.RepositoryTypes,
)
