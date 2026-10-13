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

// pipeEnd is one side of a pipe. It counts the descriptor table entries that
// refer to it, and closes its side of the pipe when the last one closes.
type pipeEnd struct {
	pipe *bufferedPipe

	// mtx guards refs.
	mtx sync.Mutex
	// refs counts the descriptor table entries that refer to this end.
	refs int
	// reader is set when this end owns the read side.
	reader bool
}

// newPipeEnds returns the read and write ends of a new pipe with inode number
// ino, each with one reference.
func newPipeEnds(ino uint64) (*pipeEnd, *pipeEnd) {
	pipe := &bufferedPipe{ino: ino, waitCh: make(chan struct{})}
	return &pipeEnd{pipe: pipe, refs: 1, reader: true}, &pipeEnd{pipe: pipe, refs: 1}
}

// ref adds a descriptor table entry's reference.
func (pe *pipeEnd) ref() {
	pe.mtx.Lock()
	pe.refs++
	pe.mtx.Unlock()
}

// unref drops a reference, closing the end after the last one.
func (pe *pipeEnd) unref() {
	// Drop the reference.
	pe.mtx.Lock()
	if pe.refs == 0 {
		pe.mtx.Unlock()
		return
	}
	pe.refs--
	closed := pe.refs == 0
	pe.mtx.Unlock()

	// Close the end after the last one.
	if closed {
		pe.pipe.close(pe.reader)
	}
}

// read reads from the pipe. It reports true at end of file or when ctx ends.
func (pe *pipeEnd) read(ctx context.Context, dat []byte) (int, bool) {
	return pe.pipe.read(ctx, dat)
}

// write writes to the pipe, waiting while it is full. It reports false when
// the read side is closed.
func (pe *pipeEnd) write(ctx context.Context, dat []byte) (int, bool) {
	return pe.pipe.write(ctx, dat)
}

// bufferedPipe is a pipe whose buffer holds up to pipeCapacity bytes. A
// writer waits while it is full, and a reader while it is empty.
type bufferedPipe struct {
	// ino is the pipe's inode number.
	ino uint64

	// mtx guards the fields below.
	mtx sync.Mutex
	// buf holds the bytes written and not yet read.
	buf []byte
	// readClosed and writeClosed are set when that side has no open
	// reference.
	readClosed  bool
	writeClosed bool
	// waitCh closes when the buffer or the close state changes.
	waitCh chan struct{}
}

// read waits for data and copies it to dat. It reports true at end of file
// or when ctx ends.
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
			return 0, true
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
