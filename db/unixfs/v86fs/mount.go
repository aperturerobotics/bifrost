package unixfs_v86fs

import (
	"context"

	"github.com/s4wave/spacewave/db/unixfs"
)

// MountResolver resolves a mount name to an FSHandle root.
// Called when the client sends a MOUNT request.
type MountResolver func(ctx context.Context, name string) (*unixfs.FSHandle, error)

// MountEntry is a dynamic mount table entry.
type MountEntry struct {
	// Name is the mount name used for v86fs MOUNT resolution.
	Name string
	// Path is the guest filesystem path to mount at.
	Path string
	// Handle is the FSHandle root for this mount.
	Handle *unixfs.FSHandle
}
