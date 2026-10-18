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

	// stderr routes console writes to the console's stderr stream.
	stderr bool
	// tty is the terminal the file reads and writes, or nil.
	tty *tty
	// ptm is the pseudo-terminal whose master the file is, or nil. Its rx
	// reads the slave's output.
	ptm *pty

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

	// counter is the count an eventfd or timerfd file reads, or nil.
	counter *counter
	// timer is the timer of a timerfd file, which adds to its counter, or
	// nil.
	timer *intervalTimer
	// epoll is the interest list of an epoll file, or nil.
	epoll *epollSet
	// signals is the signal source of a signalfd file, or nil.
	signals *signalFile

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

// release closes the file's sides of its pipes, its socket and its
// pseudo-terminal, and stops its timer: a slave's last close ends the
// master's input, and the master's close hangs up the slave.
func (of *OpenFile) release() {
	// Close the pipes and the socket, and stop the timer.
	if of.rx != nil {
		of.rx.close(true)
	}
	if of.tx != nil {
		of.tx.close(false)
	}
	if of.sock != nil {
		of.sock.close()
	}
	if of.timer != nil {
		of.timer.stop()
	}

	// Close the pseudo-terminal's side.
	if of.tty != nil && of.tty.pty != nil {
		of.tty.pty.closeSlave()
	}
	if of.ptm != nil {
		of.ptm.hangup()
	}
}

// terminal returns the terminal of a terminal file: the console, a pty slave,
// or the slave of a pty master. It returns nil for any other file, and for
// the console when it is not a terminal.
func (of *OpenFile) terminal() *tty {
	t := of.tty
	if of.ptm != nil {
		t = of.ptm.term
	}
	if t == nil || !t.terminal {
		return nil
	}
	return t
}

// readiness reports whether a read of the file, or a write when write is
// set, would proceed without waiting, and whether the other side closed.
// change closes at the file's next change while it is not ready. A regular
// file is always ready and has no change. An epoll file reports neither
// readiness nor a change, since its readiness is not watched.
func (of *OpenFile) readiness(write bool) (ready, hangup bool, change <-chan struct{}) {
	switch {
	case of.counter != nil:
		ready, change := of.counter.state(write)
		return ready, false, change
	case of.signals != nil:
		ready, change := of.signals.state()
		return ready && !write, false, change
	case of.epoll != nil:
		return false, false, nil
	case write && of.tx != nil:
		return of.tx.state(true)
	case write:
		return true, false, nil
	case of.rx != nil:
		return of.rx.state(false)
	case of.sock != nil && of.sock.listener != nil:
		ready, change := of.sock.listener.state()
		return ready, false, change
	case of.tty != nil:
		ready, change := of.tty.in.ready()
		return ready, false, change
	}
	return true, false, nil
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

// unrefFiles drops one reference to each file, as a close of each would.
func unrefFiles(files []*OpenFile) {
	for _, of := range files {
		of.unref()
	}
}

// nonblocking returns a child of ctx that has already ended, so a read or
// write under it takes what is ready and does not wait.
func nonblocking(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	return ctx
}
