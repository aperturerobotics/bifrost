//go:build tinygo

package space_world_ops

import (
	"github.com/aperturerobotics/common/typeid"
	git_world "github.com/s4wave/spacewave/db/git/world"
)

// buildTypes lists the types of the TinyGo build.
//
// TinyGo Spaces apply no Git operations but still read Git worktrees.
var buildTypes = []typeid.Type{
	{ID: git_world.WorktreeTypeID, New: func() any { return new(git_world.Worktree) }},
}
