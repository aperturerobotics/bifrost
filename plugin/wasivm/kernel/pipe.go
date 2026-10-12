package kernel

import (
	"context"
	"sync"
)

// pipeEnd is one side of a pipe. It counts the descriptor table entries that
// refer to it, and closes its side of the pipe when the last one closes.
type pipeEnd struct {
	pipe *bufferedPipe

	// mtx guards refs and holds.
	mtx sync.Mutex
	// refs counts the descriptor table entries that refer to this end.
	refs int
	// holds counts vfork parents that keep this end open while the child
	// changes the descriptor table.
	holds int
	// reader is set when this end owns the read side.
	reader bool
}

// newPipeEnds returns the read and write ends of a new pipe, each with one
// reference.
func newPipeEnds() (*pipeEnd, *pipeEnd) {
	pipe := newBufferedPipe()
	return &pipeEnd{pipe: pipe, refs: 1, reader: true}, &pipeEnd{pipe: pipe, refs: 1}
}

// ref adds a reference, reopening the end if its last reference had closed.
func (pe *pipeEnd) ref() {
	// Count the reference.
	pe.mtx.Lock()
	reopen := pe.refs == 0
	pe.refs++
	pe.mtx.Unlock()

	// Reopen the end after a close.
	if reopen {
		pe.open()
	}
}

// unref drops a reference, closing the end when no reference or hold keeps
// it open.
func (pe *pipeEnd) unref() {
	// Drop the reference.
	pe.mtx.Lock()
	if pe.refs == 0 {
		pe.mtx.Unlock()
		return
	}
	pe.refs--
	closed := pe.refs == 0 && pe.holds == 0
	pe.mtx.Unlock()

	// Close the end after the last one.
	if !closed {
		return
	}
	if pe.reader {
		pe.pipe.closeRead()
		return
	}
	pe.pipe.closeWrite()
}

// hold keeps the end open for a vfork parent. A held read end accepts writes
// even while no reader is open, so a pipeline's writer can run before its
// reader starts.
func (pe *pipeEnd) hold() {
	pe.mtx.Lock()
	pe.holds++
	pe.mtx.Unlock()
	if pe.reader {
		pe.pipe.allowDetachedRead()
	}
}

// restoreHeld turns a hold back into a reference when the vfork parent
// resumes.
func (pe *pipeEnd) restoreHeld() {
	// Move the hold to the references.
	pe.mtx.Lock()
	if pe.holds == 0 {
		pe.mtx.Unlock()
		return
	}
	reopen := pe.refs == 0
	pe.holds--
	pe.refs++
	pe.mtx.Unlock()

	// Reopen the end if the child closed it.
	if reopen {
		pe.open()
	}
}

// open reopens this end's side of the pipe.
func (pe *pipeEnd) open() {
	if pe.reader {
		pe.pipe.openRead()
		return
	}
	pe.pipe.openWrite()
}

// read reads from the pipe. It reports true at end of file or when ctx ends.
func (pe *pipeEnd) read(ctx context.Context, dat []byte) (int, bool) {
	return pe.pipe.read(ctx, dat)
}

// write writes to the pipe. It reports false when the read side is closed.
func (pe *pipeEnd) write(dat []byte) (int, bool) {
	return pe.pipe.write(dat)
}

// bufferedPipe is a pipe whose buffer grows without bound, so a writer never
// blocks. A pipeline's stages can then run one after another under vfork.
type bufferedPipe struct {
	// mtx guards the fields below.
	mtx sync.Mutex
	// buf holds the bytes written and not yet read.
	buf []byte
	// readClosed is set when the read side has no open reference.
	readClosed bool
	// detachedRead lets writes succeed while the read side is closed, so a
	// vfork pipeline's writer can run before its reader starts.
	detachedRead bool
	// writeClosed is set when the write side has no open reference.
	writeClosed bool
	// waitCh closes when the buffer or the close state changes.
	waitCh chan struct{}
}

// newBufferedPipe returns an empty open pipe.
func newBufferedPipe() *bufferedPipe {
	return &bufferedPipe{waitCh: make(chan struct{})}
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
		select {
		case <-ctx.Done():
			return 0, true
		case <-waitCh:
		}
	}
}

// write appends dat to the buffer. It reports false when no reader can
// receive it.
func (p *bufferedPipe) write(dat []byte) (int, bool) {
	// Buffer the data while a reader can still take it.
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if p.readClosed && !p.detachedRead {
		return 0, false
	}
	p.buf = append(p.buf, dat...)
	p.notifyLocked()
	return len(dat), true
}

// closeRead marks the read side closed.
func (p *bufferedPipe) closeRead() {
	p.mtx.Lock()
	p.readClosed = true
	p.mtx.Unlock()
}

// closeWrite marks the write side closed and wakes the readers.
func (p *bufferedPipe) closeWrite() {
	p.mtx.Lock()
	p.writeClosed = true
	p.notifyLocked()
	p.mtx.Unlock()
}

// openRead reopens the read side.
func (p *bufferedPipe) openRead() {
	p.mtx.Lock()
	p.readClosed = false
	p.mtx.Unlock()
}

// openWrite reopens the write side.
func (p *bufferedPipe) openWrite() {
	p.mtx.Lock()
	p.writeClosed = false
	p.mtx.Unlock()
}

// allowDetachedRead lets writes succeed while the read side is closed.
func (p *bufferedPipe) allowDetachedRead() {
	p.mtx.Lock()
	p.detachedRead = true
	p.mtx.Unlock()
}

// notifyLocked wakes the readers waiting on a change. The caller holds mtx.
func (p *bufferedPipe) notifyLocked() {
	close(p.waitCh)
	p.waitCh = make(chan struct{})
}
