package kernel

import (
	"context"
	"encoding/binary"
	"math"
)

// eventfd flags as the guest's headers number them: EFD_NONBLOCK is
// O_NONBLOCK, and EFD_CLOEXEC is wasi-libc's O_CLOEXEC, zero.
const (
	guestEfdSemaphore = 0x1
	guestEfdNonblock  = guestONONBLOCK
)

// sysEventfd opens an event file whose counter starts at initval.
func (p *Process) sysEventfd(initval, flags uint32) uint64 {
	if flags&^(guestEfdSemaphore|guestEfdNonblock) != 0 {
		return negErrno(EINVAL)
	}
	return uint64(p.installFD(&OpenFile{
		path:    "anon_inode:[eventfd]",
		flags:   oRDWR | kernelStatusFlags(flags&guestEfdNonblock),
		counter: newCounter(uint64(initval), flags&guestEfdSemaphore != 0),
	}))
}

// readCounter reads an event or timer file: it waits for a nonzero count and
// takes it as an 8-byte value. A take of nothing returns EINTR when a signal
// interrupted it, or EAGAIN when the file is nonblocking.
func (p *Process) readCounter(ctx context.Context, of *OpenFile, count int) ([]byte, uint64) {
	if count < 8 {
		return nil, negErrno(EINVAL)
	}
	n := of.counter.take(of.ioContext(ctx))
	if n == 0 {
		return nil, readStatus(ctx, of, 0, false)
	}
	return binary.LittleEndian.AppendUint64(nil, n), 0
}

// writeCounter adds the 8-byte value in buf to an event file's counter,
// waiting while the sum would overflow. A timer file refuses writes.
func (p *Process) writeCounter(ctx context.Context, of *OpenFile, buf []byte) uint64 {
	// Read the value.
	if of.timer != nil || len(buf) < 8 {
		return negErrno(EINVAL)
	}
	n := binary.LittleEndian.Uint64(buf)
	if n == math.MaxUint64 {
		return negErrno(EINVAL)
	}

	// Add it, letting a caught signal interrupt the wait.
	ctx, release := p.interruptible(ctx)
	defer release()
	switch {
	case of.counter.add(of.ioContext(ctx), n):
		return 8
	case interrupted(ctx):
		return negErrno(EINTR)
	}
	return negErrno(EAGAIN)
}
