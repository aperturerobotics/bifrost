//go:build !tinygo && !goscript

package space_world_ops

import (
	"slices"

	s4wave_git "github.com/s4wave/spacewave/core/git"
	git_world "github.com/s4wave/spacewave/db/git/world"
	identity_world "github.com/s4wave/spacewave/identity/world"
	s4wave_org "github.com/s4wave/spacewave/sdk/org"
)

// buildTypes lists the operations of the native build.
var buildTypes = slices.Concat(
	git_world.GitTypes,
	s4wave_git.GitTypes,
	identity_world.WorldTypes,
	s4wave_org.OrgTypes,
)
