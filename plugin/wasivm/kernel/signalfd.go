package kernel

import (
	"context"
	"encoding/binary"
)

// signalfd flags as the guest's headers number them: Linux's values, as for
// timerfd.
const (
	guestSfdCloexec  = 0o2000000
	guestSfdNonblock = 0o4000
)

// signalfdSiginfoSize is the size of a struct signalfd_siginfo, whose first
// field is the signal number.
const signalfdSiginfoSize = 128

// signalFile is a signalfd file. A read takes the signals of its mask held
// pending for its process, which blocks them so they are not delivered.
type signalFile struct {
	// p is the process whose signals the file reads: the one that opened
	// it. Unlike Linux, which reads the reader's signals, a child that
	// inherits the file reads its parent's.
	p *Process
	// mask is the signals the file reads, in the Linux layout. The kernel
	// lock guards it.
	mask uint64
}

// sysSignalfd opens a signal file reading the signals in the set at
// maskPtr, or, when fd is a signal file and not -1, replaces its mask.
func (p *Process) sysSignalfd(fd int32, maskPtr, flags uint32) uint64 {
	// Read the mask and check the flags.
	mask, ok := p.inst.Memory().ReadUint64Le(maskPtr)
	switch {
	case !ok:
		return negErrno(EFAULT)
	case flags&^(guestSfdCloexec|guestSfdNonblock) != 0:
		return negErrno(EINVAL)
	}
	mask &^= sigKillStopMask

	// Open a file.
	if fd == -1 {
		var status uint32
		if flags&guestSfdNonblock != 0 {
			status = oNONBLOCK
		}
		return uint64(p.installFD(&OpenFile{
			path:    "anon_inode:[signalfd]",
			flags:   oRDONLY | status,
			signals: &signalFile{p: p, mask: mask},
		}, flags&guestSfdCloexec != 0))
	}

	// Replace the file's mask, waking its pollers to check it.
	of, ok := p.fds[uint32(fd)]
	switch {
	case !ok:
		return negErrno(EBADF)
	case of.signals == nil:
		return negErrno(EINVAL)
	}
	sf := of.signals
	sf.mask = mask
	sf.p.sigMu.Lock()
	sf.p.wakeSigWaitersLocked()
	sf.p.sigMu.Unlock()
	return uint64(fd)
}

// state reports whether a signal of the file's mask is pending, and returns
// the channel that closes at the process's next signal delivery.
func (sf *signalFile) state() (ready bool, change <-chan struct{}) {
	p := sf.p
	p.sigMu.Lock()
	defer p.sigMu.Unlock()
	return p.sigPending&sf.mask != 0, p.sigWaitChLocked()
}

// readSignals reads a signal file: it waits for a pending signal of its
// mask, then takes it and the others pending, as many as count bytes of
// signalfd_siginfo records hold. A take of nothing returns EINTR when a
// signal interrupted it, or EAGAIN when the file is nonblocking.
func (p *Process) readSignals(ctx context.Context, of *OpenFile, count int) ([]byte, uint64) {
	// Wait for the first signal.
	if count < signalfdSiginfoSize {
		return nil, negErrno(EINVAL)
	}
	sf := of.signals
	sig := sf.p.takeSignal(of.ioContext(ctx), sf.mask)
	if sig == 0 {
		return nil, readStatus(ctx, of, 0, false)
	}

	// Record it and each other pending signal that fits.
	var buf []byte
	for sig != 0 {
		info := make([]byte, signalfdSiginfoSize)
		binary.LittleEndian.PutUint32(info, sig)
		buf = append(buf, info...)
		if len(buf)+signalfdSiginfoSize > count {
			break
		}
		sig = sf.p.takeSignal(nonblocking(ctx), sf.mask)
	}
	return buf, 0
}
