//go:build linux

package fuse

import (
	"bytes"
	"context"
	"io"
	"testing"

	"bazil.org/fuse"
	hydra_testbed "github.com/s4wave/spacewave/db/testbed"
	"github.com/s4wave/spacewave/db/unixfs"
	unixfs_world "github.com/s4wave/spacewave/db/unixfs/world"
	unixfs_world_testbed "github.com/s4wave/spacewave/db/unixfs/world/testbed"
	"github.com/s4wave/spacewave/db/world"
	world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	"github.com/sirupsen/logrus"
)

// TestInodeOpenUsesSynchronousDirectIO checks the existing-file open path.
func TestInodeOpenUsesSynchronousDirectIO(t *testing.T) {
	// Open an existing inode of a writable mount.
	inode := &Inode{rfs: &RootFS{}}
	request := &fuse.OpenRequest{Flags: fuse.OpenReadWrite}
	response := &fuse.OpenResponse{}
	handle, err := inode.Open(context.Background(), request, response)
	if err != nil {
		t.Fatal(err)
	}

	// The open returns a Handle that bypasses the page cache.
	if _, ok := handle.(*Handle); !ok {
		t.Fatalf("expected *Handle, got %T", handle)
	}
	if response.Flags&fuse.OpenDirectIO == 0 {
		t.Fatal("expected OpenDirectIO response flag")
	}
}

// TestInodeOpenReadOnlyKeepsCache checks that a read-only mount serves opens
// through the kernel page cache, which memory-mapped files need.
func TestInodeOpenReadOnlyKeepsCache(t *testing.T) {
	// Open an existing inode of a read-only mount.
	inode := &Inode{rfs: &RootFS{readOnly: true}}
	response := &fuse.OpenResponse{}
	if _, err := inode.Open(context.Background(), &fuse.OpenRequest{}, response); err != nil {
		t.Fatal(err)
	}

	// The open keeps cached pages and does not bypass the cache.
	if response.Flags&fuse.OpenKeepCache == 0 {
		t.Fatal("expected OpenKeepCache response flag")
	}
	if response.Flags&fuse.OpenDirectIO != 0 {
		t.Fatal("expected no OpenDirectIO response flag")
	}
}

// TestInodeCreateUsesSynchronousDirectIO checks the create-and-open path.
func TestInodeCreateUsesSynchronousDirectIO(t *testing.T) {
	// Build a store.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	le := logrus.NewEntry(logrus.New())

	// Initialize the filesystem in a World.
	tb, err := hydra_testbed.NewTestbed(ctx, le)
	if err != nil {
		t.Fatal(err)
	}
	wtb, err := world_testbed.NewTestbed(tb)
	if err != nil {
		t.Fatal(err)
	}
	rootHandle, err := unixfs_world_testbed.InitTestbed(wtb, "test/fuse-create", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(rootHandle.Release)

	// Create a file through the root inode.
	rootFS := &RootFS{ctx: ctx, ctxCancel: cancel, le: le}
	inode := NewInode(rootFS, nil, rootHandle)
	request := &fuse.CreateRequest{
		Name:  "created.txt",
		Flags: fuse.OpenReadWrite,
		Mode:  0o644,
	}
	response := &fuse.CreateResponse{}
	created, handle, err := inode.Create(ctx, request, response)
	if err != nil {
		t.Fatal(err)
	}

	// Create returns the new Inode.
	createdInode, ok := created.(*Inode)
	if !ok {
		t.Fatalf("expected *Inode, got %T", created)
	}
	t.Cleanup(createdInode.h.Release)

	// The create returns a Handle that bypasses the page cache.
	opened, ok := handle.(*Handle)
	if !ok {
		t.Fatalf("expected *Handle, got %T", handle)
	}
	if response.Flags&fuse.OpenDirectIO == 0 {
		t.Fatal("expected OpenDirectIO response flag")
	}

	// The write is committed to the file before Write returns.
	data := []byte("created synchronously")
	writeResponse := &fuse.WriteResponse{}
	if err := opened.Write(ctx, &fuse.WriteRequest{Offset: 0, Data: data}, writeResponse); err != nil {
		t.Fatal(err)
	}
	if writeResponse.Size != len(data) {
		t.Fatalf("expected write size %d, got %d", len(data), writeResponse.Size)
	}
	got := make([]byte, len(data))
	if _, err := createdInode.h.ReadAt(ctx, 0, got); err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if string(got) != string(data) {
		t.Fatalf("expected %q, got %q", data, got)
	}
}

// TestHandleFlushCommitsSmallWrites checks that small writes stay buffered at
// the inode and that the flush on close commits them to the World.
func TestHandleFlushCommitsSmallWrites(t *testing.T) {
	// Build a store.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	le := logrus.NewEntry(logrus.New())

	// Initialize the filesystem in a World.
	tb, err := hydra_testbed.NewTestbed(ctx, le)
	if err != nil {
		t.Fatal(err)
	}
	wtb, err := world_testbed.NewTestbed(tb)
	if err != nil {
		t.Fatal(err)
	}
	rootHandle, err := unixfs_world_testbed.InitTestbed(wtb, "test/fuse-flush", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(rootHandle.Release)

	// Create a file through the root inode.
	rootFS := &RootFS{ctx: ctx, ctxCancel: cancel, le: le}
	inode := NewInode(rootFS, nil, rootHandle)
	created, handle, err := inode.Create(ctx, &fuse.CreateRequest{
		Name:  "flushed.txt",
		Flags: fuse.OpenReadWrite,
		Mode:  0o644,
	}, &fuse.CreateResponse{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(created.(*Inode).h.Release)

	// Write the file in 1 KiB pieces through the FUSE handle.
	opened := handle.(*Handle)
	piece := bytes.Repeat([]byte("x"), 1024)
	for i := range 16 {
		request := &fuse.WriteRequest{Offset: int64(i * len(piece)), Data: piece}
		if err := opened.Write(ctx, request, &fuse.WriteResponse{}); err != nil {
			t.Fatal(err)
		}
	}

	// None of the buffered writes are committed yet.
	if size := committedSize(t, ctx, le, wtb); size != 0 {
		t.Fatalf("committed %d bytes before the flush", size)
	}

	// The flush on close commits all 16 KiB.
	if err := opened.Flush(ctx, &fuse.FlushRequest{}); err != nil {
		t.Fatal(err)
	}
	if size := committedSize(t, ctx, le, wtb); size != uint64(16*len(piece)) {
		t.Fatalf("committed %d bytes after the flush, want %d", size, 16*len(piece))
	}
}

// committedSize returns the size of flushed.txt in a fresh tree over the
// World, which reads only committed bytes.
func committedSize(t *testing.T, ctx context.Context, le *logrus.Entry, wtb *world_testbed.Testbed) uint64 {
	// Open a tree with its own inodes and buffers.
	cursor := unixfs_world.NewFSCursorWithContext(
		ctx,
		le,
		world.NewEngineWorldState(wtb.Engine, true),
		"test/fuse-flush",
		unixfs_world.FSType_FSType_FS_NODE,
		nil,
		false,
	)
	readHandle, err := unixfs.NewFSHandle(cursor)
	if err != nil {
		t.Fatal(err)
	}
	defer readHandle.Release()

	// Read the file's size.
	file, err := readHandle.Lookup(ctx, "flushed.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Release()
	size, err := file.GetSize(ctx)
	if err != nil {
		t.Fatal(err)
	}
	return size
}
