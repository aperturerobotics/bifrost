//go:build !tinygo

package space_unixfs

import (
	"context"

	git_world "github.com/s4wave/spacewave/db/git/world"
	"github.com/s4wave/spacewave/db/unixfs"
	unixfs_errors "github.com/s4wave/spacewave/db/unixfs/errors"
	unixfs_world "github.com/s4wave/spacewave/db/unixfs/world"
	"github.com/s4wave/spacewave/db/world"
	world_types "github.com/s4wave/spacewave/db/world/types"
	git_repofs "github.com/s4wave/spacewave/sdk/git/repofs"
	s4wave_git_world "github.com/s4wave/spacewave/sdk/git/world"
	"github.com/sirupsen/logrus"
)

// supportsProjectedType reports whether objects of typeID can be projected.
func supportsProjectedType(typeID string) bool {
	switch typeID {
	case unixfs_world.FSNodeTypeID, unixfs_world.FSObjectTypeID, unixfs_world.FSHostVolumeTypeID:
		return true
	case s4wave_git_world.GitRepoTypeID, s4wave_git_world.GitWorktreeTypeID:
		return true
	default:
		return false
	}
}

// openObjectCursor opens the filesystem cursor of a projected object.
func openObjectCursor(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	objectKey string,
) (unixfs.FSCursor, error) {
	// Require an object type supported by filesystem projection.
	typeID, err := world_types.GetObjectType(ctx, ws, objectKey)
	if err != nil {
		return nil, err
	}
	if !supportsProjectedType(typeID) {
		return nil, unixfs_errors.ErrNotExist
	}

	// Open Git repositories and worktrees through their dedicated cursors.
	switch typeID {
	case s4wave_git_world.GitRepoTypeID:
		return openGitRepoCursor(ctx, ws, objectKey)
	case s4wave_git_world.GitWorktreeTypeID:
		return openGitWorktreeCursor(ctx, le, ws, objectKey)
	}

	// Resolve the filesystem type and open its World cursor.
	fsType, _, err := unixfs_world.LookupFsType(ctx, ws, objectKey)
	if err != nil {
		return nil, err
	}
	return unixfs_world.NewFSCursor(le, ws, objectKey, fsType, nil, false), nil
}

// openGitWorktreeCursor opens the working directory of a Git worktree.
func openGitWorktreeCursor(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	objectKey string,
) (unixfs.FSCursor, error) {
	workdirRef, err := git_world.WorktreeLookupWorkdirRef(ctx, ws, objectKey)
	if err != nil {
		return nil, err
	}
	return unixfs_world.FollowUnixfsRef(ctx, le, ws, workdirRef, "", false)
}

// openGitRepoCursor opens a read-only cursor over a Git repository.
func openGitRepoCursor(ctx context.Context, ws world.WorldState, objectKey string) (unixfs.FSCursor, error) {
	return git_repofs.OpenRepoFSCursor(ctx, ws, objectKey, false)
}
