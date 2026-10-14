package kernel

import "context"

// OpenFile represents an open file description. Descriptors made by dup,
// exec and fork share one OpenFile, and so share its offset and flags. The
// kernel lock guards its fields.
type OpenFile struct {
	// node is the file's node, or nil for a pipe or a socket.
	node *VNode
	// path is the absolute path at the time of the open.
	path string
	// offset is the read and write position of a regular file.
	offset int64
	// flags holds the open flags, in the kernel's values.
	flags uint32

	// stderr routes console writes to the kernel's stderr stream.
	stderr bool

	// refs counts the descriptor table entries that refer to the file.
	refs int
	// rx is the pipe the file reads: a pipe's read end, or a connected
	// socket's incoming direction.
	rx *bufferedPipe
	// tx is the pipe the file writes: a pipe's write end, or a connected
	// socket's outgoing direction.
	tx *bufferedPipe
	// sock is the socket the file is, or nil.
	sock *unixSocket

	// data is a synthesized file's content, generated at open.
	data []byte

	// dirNames holds the directory reading state for getdents64.
	dirNames []string
}

// ref adds a descriptor table entry's reference.
func (of *OpenFile) ref() {
	of.refs++
}

// unref drops a descriptor table entry's reference, and releases the file
// after the last one.
func (of *OpenFile) unref() {
	of.refs--
	if of.refs == 0 {
		of.release()
	}
}

// release closes the file's sides of its pipes and its socket.
func (of *OpenFile) release() {
	if of.rx != nil {
		of.rx.close(true)
	}
	if of.tx != nil {
		of.tx.close(false)
	}
	if of.sock != nil {
		of.sock.close()
	}
}

// pipe returns the pipe a pipe file reads or writes, or nil.
func (of *OpenFile) pipe() *bufferedPipe {
	if of.rx != nil {
		return of.rx
	}
	return of.tx
}

// ioContext returns the context of a read or write of the file, which ends at
// once for a nonblocking file.
func (of *OpenFile) ioContext(ctx context.Context) context.Context {
	if of.flags&oNONBLOCK == 0 {
		return ctx
	}
	return nonblocking(ctx)
}

// nonblocking returns a child of ctx that has already ended, so a read or
// write under it takes what is ready and does not wait.
func nonblocking(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	return ctx
}
