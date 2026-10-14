package kernel

import (
	"encoding/binary"
	"time"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// timerfd flags as the guest's headers number them: Linux's values, not
// wasi-libc's O_NONBLOCK.
const (
	guestTfdCloexec      = 0o2000000
	guestTfdNonblock     = 0o4000
	guestTfdTimerAbstime = 0x1
	// guestTfdTimerCancelOnSet is accepted and ignored: a clock set never
	// cancels a timer.
	guestTfdTimerCancelOnSet = 0x2
)

// timespecSize is the size of a wasm32 struct timespec, a 64-bit tv_sec and
// a 32-bit tv_nsec padded to 8 bytes. A struct itimerspec holds two: the
// interval, then the value.
const timespecSize = 16

// sysTimerfdCreate opens a disarmed timer file following clockID.
func (p *Process) sysTimerfdCreate(clockID, flags uint32) uint64 {
	// Check the clock and the flags.
	switch {
	case clockID != clockRealtime && clockID != clockMonotonic && clockID != clockBoottime:
		return negErrno(EINVAL)
	case flags&^(guestTfdCloexec|guestTfdNonblock) != 0:
		return negErrno(EINVAL)
	}

	// Open the file.
	var status uint32
	if flags&guestTfdNonblock != 0 {
		status = oNONBLOCK
	}
	timer := newTimerFile(p.kernel, clockID)
	return uint64(p.installFD(&OpenFile{
		path:    "anon_inode:[timerfd]",
		flags:   oRDWR | status,
		counter: timer.count,
		timer:   timer,
	}))
}

// sysTimerfdSettime arms or disarms the timer of fd from the itimerspec at
// newPtr, writing the old setting to oldPtr when it is set.
func (p *Process) sysTimerfdSettime(fd, flags, newPtr, oldPtr uint32) uint64 {
	// Find the timer and read the setting.
	of, ok := p.fds[fd]
	switch {
	case !ok:
		return negErrno(EBADF)
	case of.timer == nil:
		return negErrno(EINVAL)
	case flags&^(guestTfdTimerAbstime|guestTfdTimerCancelOnSet) != 0:
		return negErrno(EINVAL)
	}
	mem := p.inst.Memory()
	interval, value, errno := readItimerspec(mem, newPtr)
	if errno != 0 {
		return negErrno(errno)
	}

	// Apply it and report the old one.
	oldValue, oldInterval := of.timer.set(value, interval, flags&guestTfdTimerAbstime != 0)
	if oldPtr != 0 && !writeItimerspec(mem, oldPtr, oldInterval, oldValue) {
		return negErrno(EFAULT)
	}
	return 0
}

// sysTimerfdGettime writes the timer of fd's time until its next expiration
// and its period to currPtr.
func (p *Process) sysTimerfdGettime(fd, currPtr uint32) uint64 {
	// Find the timer.
	of, ok := p.fds[fd]
	switch {
	case !ok:
		return negErrno(EBADF)
	case of.timer == nil:
		return negErrno(EINVAL)
	}

	// Report the setting.
	value, interval := of.timer.get()
	if !writeItimerspec(p.inst.Memory(), currPtr, interval, value) {
		return negErrno(EFAULT)
	}
	return 0
}

// readItimerspec reads a wasm32 struct itimerspec's interval and value as
// durations. It returns an errno for an unreadable pointer or a negative or
// out-of-range field.
func readItimerspec(mem engine.Memory, ptr uint32) (time.Duration, time.Duration, uint32) {
	var d [2]time.Duration
	for i := range d {
		t, errno := readTimespec(mem, ptr+uint32(i)*timespecSize)
		if errno != 0 {
			return 0, 0, errno
		}
		if d[i] = t.Sub(time.Unix(0, 0)); d[i] < 0 {
			return 0, 0, EINVAL
		}
	}
	return d[0], d[1], 0
}

// writeItimerspec writes interval and value as a wasm32 struct itimerspec,
// reporting false for an unwritable pointer.
func writeItimerspec(mem engine.Memory, ptr uint32, interval, value time.Duration) bool {
	buf := make([]byte, 2*timespecSize)
	for i, d := range []time.Duration{interval, value} {
		binary.LittleEndian.PutUint64(buf[i*timespecSize:], uint64(d/time.Second))
		binary.LittleEndian.PutUint32(buf[i*timespecSize+8:], uint32(d%time.Second))
	}
	return mem.Write(ptr, buf)
}
