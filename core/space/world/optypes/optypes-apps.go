//go:build !tinygo && !goscript

package optypes

import (
	"slices"

	manifest_world "github.com/s4wave/spacewave/bldr/manifest/world"
	s4wave_apt "github.com/s4wave/spacewave/sdk/apt"
)

// buildTypes lists the operations of the native build.
var buildTypes = slices.Concat(manifest_world.WorldTypes, s4wave_apt.RepositoryTypes)
