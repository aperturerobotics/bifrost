//go:build tinygo

package space_unixfs

import (
	"context"

	"github.com/s4wave/spacewave/db/unixfs"
	unixfs_errors "github.com/s4wave/spacewave/db/unixfs/errors"
	unixfs_world "github.com/s4wave/spacewave/db/unixfs/world"
	"github.com/s4wave/spacewave/db/world"
	world_types "github.com/s4wave/spacewave/db/world/types"
	"github.com/sirupsen/logrus"
)

// supportsProjectedType reports whether objects of typeID can be projected.
func supportsProjectedType(typeID string) bool {
	switch typeID {
	case unixfs_world.FSNodeTypeID, unixfs_world.FSObjectTypeID, unixfs_world.FSHostVolumeTypeID:
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

	// Resolve the filesystem type and open its World cursor.
	fsType, _, err := unixfs_world.LookupFsType(ctx, ws, objectKey)
	if err != nil {
		return nil, err
	}
	return unixfs_world.NewFSCursor(le, ws, objectKey, fsType, nil, false), nil
}
