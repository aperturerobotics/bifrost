package engine

import "context"

// Backend provides immutable files, atomic whole-file replacement, and locks.
// Lock names are local to one volume and shared by all of its runtime instances.
// Read returns io/fs.ErrNotExist for missing files. Write acknowledges only
// after durable whole-file publication. Remove is idempotent.
type Backend interface {
	// Read returns an owned bounded range, or a complete file when length is readAll.
	Read(ctx context.Context, name string, offset int64, length int) ([]byte, error)
	// Write acknowledges durable replacement; interrupted first creation may leave an empty entry.
	Write(ctx context.Context, name string, data []byte) error
	// Remove deletes a file and succeeds when it is already absent.
	Remove(ctx context.Context, name string) error
	// Lock holds origin-wide shared or exclusive access until its release is called.
	Lock(ctx context.Context, name string, exclusive bool) (func(), error)
	// Subscribe registers a lossy publication hint; its release closes resources.
	Subscribe() (<-chan struct{}, func())
	// Notify announces a durable generation without affecting commit success.
	Notify(generation uint64)
	// Close releases instance-local notification resources.
	Close() error
}

// readAll requests a whole bounded immutable file from a backend.
const readAll = -1
