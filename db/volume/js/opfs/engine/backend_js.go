//go:build js

package engine

import (
	"context"
	"io"
	"io/fs"
	"runtime/trace"
	"syscall/js"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/opfs"
)

// browserBackend adapts browser OPFS operations to the engine's durable file
// and origin-wide lock contract.
type browserBackend struct {
	// channel sends publication hints for this volume identity.
	channel js.Value
	// driver owns the browser resources opened by this backend.
	driver opfs.Driver
	// dir contains every file owned by this engine.
	dir js.Value
	// lockPrefix separates this engine's Web Locks from other volume locks.
	lockPrefix string
}

// NewBrowserBackend constructs an engine backend in one OPFS directory.
func NewBrowserBackend(driver opfs.Driver, dir js.Value, lockPrefix string) Backend {
	channel, _ := driver.NewBroadcastChannel(lockPrefix + "/engine/changed")
	return &browserBackend{
		channel:    channel,
		driver:     driver,
		dir:        dir,
		lockPrefix: lockPrefix + "/engine/",
	}
}

// Read opens one immutable snapshot and reads the requested bounded range.
func (b *browserBackend) Read(
	ctx context.Context,
	name string,
	offset int64,
	length int,
) (data []byte, retErr error) {
	ctx, task := trace.NewTask(ctx, "hydra/opfs-engine/read")
	defer task.End()
	defer func() { trace.Logf(ctx, "hydra/opfs-engine/read/shape", "bytes=%d files=1", len(data)) }()

	// Validate the request before opening a browser resource.
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if offset < 0 {
		return nil, errors.New("opfs engine read has negative offset")
	}
	if length < 0 && length != readAll {
		return nil, errors.New("opfs engine read has invalid length")
	}
	if length == readAll && offset != 0 {
		return nil, errors.New("opfs engine whole read has nonzero offset")
	}
	if length > maxPublicationBytes {
		return nil, errors.Wrap(ErrLimit, "opfs engine read exceeds publication limit")
	}

	// Retain one immutable File for the size check and every requested byte.
	snapshot, err := b.driver.OpenReadSnapshot(b.dir, name)
	if err != nil {
		return nil, b.classifyError(err)
	}
	defer func() {
		if err := snapshot.Close(); retErr == nil && err != nil {
			data = nil
			retErr = b.classifyError(err)
		}
	}()

	// Resolve a whole-file request before allocating its bounded result.
	if length == readAll {
		size, err := snapshot.Size()
		if err != nil {
			return nil, b.classifyError(err)
		}
		if size < 0 || size > maxPublicationBytes {
			return nil, errors.Wrapf(ErrLimit, "opfs engine file %s has invalid size %d", name, size)
		}
		length = int(size)
	}
	if length == 0 {
		return []byte{}, nil
	}

	// Read exactly the requested range from the retained immutable File.
	data = make([]byte, length)
	n, err := snapshot.ReadAt(data, offset)
	if err != nil {
		return data[:n], b.classifyError(err)
	}
	if n != length {
		return data[:n], io.ErrUnexpectedEOF
	}
	return data, nil
}

// Write durably replaces one complete file.
func (b *browserBackend) Write(ctx context.Context, name string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return b.classifyError(b.driver.WriteFile(b.dir, name, data))
}

// Remove idempotently deletes one file.
func (b *browserBackend) Remove(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := b.driver.DeleteEntry(b.dir, name, false)
	if b.driver.ClassifyError(err) == opfs.ErrorKindNotFound {
		return nil
	}
	return err
}

// Lock acquires one origin-wide engine Web Lock until the returned release runs.
func (b *browserBackend) Lock(ctx context.Context, name string, exclusive bool) (func(), error) {
	result, err := b.driver.AcquireWebLock(ctx, b.lockPrefix+name, exclusive)
	if err != nil {
		return nil, err
	}
	if result.Outcome != opfs.WebLockOutcomeAcquired {
		return nil, errors.Errorf("opfs engine Web Lock %s ended with outcome %d", name, result.Outcome)
	}
	return result.Release, nil
}

// classifyError normalizes missing browser entries for the storage contract.
func (b *browserBackend) classifyError(err error) error {
	if b.driver.ClassifyError(err) == opfs.ErrorKindNotFound {
		return fs.ErrNotExist
	}
	return err
}

// _ is a type assertion
var _ Backend = (*browserBackend)(nil)

// Subscribe wakes on cross-runtime hints; durable roots remain authoritative.
func (b *browserBackend) Subscribe() (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	channel, err := b.driver.NewBroadcastChannel(b.lockPrefix + "changed")
	if err != nil {
		return ch, func() {}
	}
	callback := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		select {
		case ch <- struct{}{}:
		default:
		}
		return nil
	})
	channel.Set("onmessage", callback)
	return ch, func() {
		channel.Set("onmessage", js.Null())
		_ = b.driver.CloseBroadcastChannel(channel)
		callback.Release()
	}
}

// Notify publishes a best-effort hint after the durable root has advanced.
func (b *browserBackend) Notify(generation uint64) {
	_ = b.driver.SendBroadcastChannel(b.channel, opfs.BroadcastMessage{Generation: generation})
}

// Close releases the instance's publication channel.
func (b *browserBackend) Close() error {
	return b.driver.CloseBroadcastChannel(b.channel)
}
