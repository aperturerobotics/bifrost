package kernel

import (
	"context"
	"sync"
)

// pipeCapacity is the size of a pipe's buffer, Linux's default.
const pipeCapacity = 65536

// pipeAtomic is the largest write a pipe never interleaves with another
// writer's, Linux's PIPE_BUF.
const pipeAtomic = 4096

// bufferedPipe is a pipe whose buffer holds up to pipeCapacity bytes. A
// writer waits while it is full, and a reader while it is empty. A pipe
// carries one direction of a pipe(2) pair or of a connected socket.
type bufferedPipe struct {
	// ino is the pipe's inode number.
	ino uint64

	// mtx guards the fields below.
	mtx sync.Mutex
	// buf holds the bytes written and not yet read.
	buf []byte
	// readClosed is set when no open file reads the pipe.
	readClosed bool
	// writeClosed is set when no open file writes the pipe.
	writeClosed bool
	// waitCh closes when the buffer or the close state changes.
	waitCh chan struct{}
}

// newBufferedPipe returns an empty pipe with inode number ino.
func newBufferedPipe(ino uint64) *bufferedPipe {
	return &bufferedPipe{ino: ino, waitCh: make(chan struct{})}
}

// read waits for data and copies it to dat. It reports true at end of file.
// When ctx ends first it returns 0 bytes without end of file, so a done ctx
// makes a read that does not wait.
func (p *bufferedPipe) read(ctx context.Context, dat []byte) (int, bool) {
	for {
		// Take buffered data, or end of file once the writers closed.
		p.mtx.Lock()
		if len(p.buf) != 0 {
			n := copy(dat, p.buf)
			p.buf = p.buf[n:]
			if len(p.buf) == 0 {
				p.buf = nil
			}
			p.notifyLocked()
			p.mtx.Unlock()
			return n, false
		}
		if p.writeClosed {
			p.mtx.Unlock()
			return 0, true
		}
		waitCh := p.waitCh
		p.mtx.Unlock()

		// Wait for a write or a close.
		if !wait(ctx, waitCh) {
			return 0, false
		}
	}
}

// write copies dat to the buffer, waiting for room. A write of at most
// pipeAtomic bytes waits until it fits whole. It returns the bytes written,
// fewer when ctx ends, and reports false when the read side closed before
// any byte was written.
func (p *bufferedPipe) write(ctx context.Context, dat []byte) (int, bool) {
	written := 0
	for {
		// Fail once no reader can receive the data.
		p.mtx.Lock()
		if p.readClosed {
			p.mtx.Unlock()
			return written, written != 0
		}

		// Copy what fits.
		room := pipeCapacity - len(p.buf)
		if room > 0 && (len(dat) > pipeAtomic || room >= len(dat)) {
			n := min(room, len(dat))
			p.buf = append(p.buf, dat[:n]...)
			dat = dat[n:]
			written += n
			p.notifyLocked()
		}
		if len(dat) == 0 {
			p.mtx.Unlock()
			return written, true
		}
		waitCh := p.waitCh
		p.mtx.Unlock()

		// Wait for a read or a close.
		if !wait(ctx, waitCh) {
			return written, true
		}
	}
}

// state reports whether a read, or a write when write is set, would proceed
// without waiting, whether the other side has closed, and a channel that
// closes at the pipe's next change.
func (p *bufferedPipe) state(write bool) (ready, hangup bool, change <-chan struct{}) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if write {
		return p.readClosed || len(p.buf) < pipeCapacity, p.readClosed, p.waitCh
	}
	return len(p.buf) != 0 || p.writeClosed, p.writeClosed, p.waitCh
}

// close marks the read or write side closed and wakes the other side.
func (p *bufferedPipe) close(read bool) {
	p.mtx.Lock()
	if read {
		p.readClosed = true
		p.buf = nil
	} else {
		p.writeClosed = true
	}
	p.notifyLocked()
	p.mtx.Unlock()
}

// notifyLocked wakes the waiters on a change. The caller holds mtx.
func (p *bufferedPipe) notifyLocked() {
	close(p.waitCh)
	p.waitCh = make(chan struct{})
}
