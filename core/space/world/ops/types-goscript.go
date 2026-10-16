//go:build goscript

package space_world_ops

import (
	"slices"

	"github.com/aperturerobotics/common/typeid"
	s4wave_git "github.com/s4wave/spacewave/core/git"
	"github.com/s4wave/spacewave/core/git/opids"
	git_world "github.com/s4wave/spacewave/db/git/world"
	identity_world "github.com/s4wave/spacewave/identity/world"
)

// buildTypes lists the operations of the GoScript build.
//
// GoScript Spaces apply only the Git operations that create a repository.
// The wizard operation is constructed through the deferred core/git
// boundary, so the Git package loads only when that operation is applied.
var buildTypes = slices.Concat(
	[]typeid.Type{
		{ID: git_world.GitInitOpTypeID, New: func() any { return new(git_world.GitInitOp) }},
		{ID: git_world.GitCreateWorktreeOpTypeID, New: func() any { return new(git_world.GitCreateWorktreeOp) }},
		{ID: opids.CreateRepoWizard, New: s4wave_git.NewCreateGitRepoWizardOp},
	},
	identity_world.WorldTypes,
)
