package kernel

import (
	"time"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// nowForClock returns the current time of a clock. Monotonic clocks, and the
// CPU-time clocks they stand in for, count from boot. It reports false for an
// unknown clock.
func (k *Kernel) nowForClock(clockID uint32) (time.Time, bool) {
	switch clockID {
	case clockRealtime, clockRealtimeCoarse:
		return k.realtimeNow(), true
	case clockMonotonic, clockMonotonicRaw, clockMonotonicCoarse, clockBoottime,
		clockProcessCPUTimeID, clockThreadCPUTimeID:
		return time.Unix(0, int64(k.monotonicElapsed())), true
	default:
		return time.Time{}, false
	}
}

// realtimeNow returns the host time shifted by the kernel's realtime offset.
func (k *Kernel) realtimeNow() time.Time {
	k.timeMu.Lock()
	offset := k.realtimeOffset
	k.timeMu.Unlock()
	return time.Now().Add(offset)
}

// setRealtime sets the kernel's realtime clock to target without changing
// the host clock.
func (k *Kernel) setRealtime(target time.Time) {
	k.timeMu.Lock()
	k.realtimeOffset = time.Until(target)
	k.timeMu.Unlock()
}

// monotonicElapsed returns the time since the kernel booted.
func (k *Kernel) monotonicElapsed() time.Duration {
	return time.Since(k.bootTime)
}

// clockResolution returns a clock's resolution as seconds and nanoseconds. It
// reports false for an unknown clock.
func clockResolution(clockID uint32) (uint32, uint32, bool) {
	switch clockID {
	case clockRealtimeCoarse, clockMonotonicCoarse:
		return 0, uint32(time.Millisecond), true
	case clockRealtime, clockMonotonic, clockMonotonicRaw, clockBoottime,
		clockProcessCPUTimeID, clockThreadCPUTimeID:
		return 0, 1, true
	default:
		return 0, 0, false
	}
}

// readTimespec32 reads a timespec of two 32-bit fields from guest memory. It
// returns an errno for an unreadable pointer or an out-of-range nanosecond
// field.
func readTimespec32(mem engine.Memory, ptr uint32) (time.Time, uint32) {
	// Read both fields and range-check the nanoseconds.
	sec, ok := mem.ReadUint32Le(ptr)
	if !ok {
		return time.Time{}, EFAULT
	}
	nsec, ok := mem.ReadUint32Le(ptr + 4)
	if !ok {
		return time.Time{}, EFAULT
	}
	if nsec >= uint32(time.Second) {
		return time.Time{}, EINVAL
	}
	return time.Unix(int64(int32(sec)), int64(nsec)), 0
}

// writeZeroItimerval writes a disarmed itimerval to guest memory.
func writeZeroItimerval(mem engine.Memory, ptr uint32) bool {
	return mem.WriteUint64Le(ptr, 0) && mem.WriteUint64Le(ptr+8, 0)
}

// readItimerval32 reads an itimerval's interval and value from guest memory.
func readItimerval32(mem engine.Memory, ptr uint32) (time.Duration, time.Duration, uint32) {
	// Read the interval, then the value after it.
	interval, errno := readTimevalDuration32(mem, ptr)
	if errno != 0 {
		return 0, 0, errno
	}
	value, errno := readTimevalDuration32(mem, ptr+8)
	if errno != 0 {
		return 0, 0, errno
	}
	return interval, value, 0
}

// readTimevalDuration32 reads a timeval of two 32-bit fields from guest
// memory as a duration.
func readTimevalDuration32(mem engine.Memory, ptr uint32) (time.Duration, uint32) {
	// Read both fields and range-check the microseconds.
	sec, ok := mem.ReadUint32Le(ptr)
	if !ok {
		return 0, EFAULT
	}
	usec, ok := mem.ReadUint32Le(ptr + 4)
	if !ok {
		return 0, EFAULT
	}
	if usec >= uint32(time.Second/time.Microsecond) {
		return 0, EINVAL
	}
	return time.Duration(sec)*time.Second + time.Duration(usec)*time.Microsecond, 0
}

// sysClockSettime sets the kernel's realtime clock from a timespec.
func (p *Process) sysClockSettime(clockID, tpPtr uint32) uint64 {
	// Only the realtime clock can be set.
	if clockID != clockRealtime {
		return negErrno(EINVAL)
	}
	if tpPtr == 0 {
		return negErrno(EFAULT)
	}

	// Read the time and set it.
	t, errno := readTimespec32(p.inst.Memory(), tpPtr)
	if errno != 0 {
		return negErrno(errno)
	}
	p.kernel.setRealtime(t)
	return 0
}

// sysSetitimer reports every interval timer as disarmed and accepts only a
// request to disarm one, since the kernel has no timer signal delivery.
func (p *Process) sysSetitimer(which, newPtr, oldPtr uint32) uint64 {
	// Report the old timer as disarmed.
	if which > itimerProf {
		return negErrno(EINVAL)
	}
	mem := p.inst.Memory()
	if oldPtr != 0 && !writeZeroItimerval(mem, oldPtr) {
		return negErrno(EFAULT)
	}

	// Accept the new timer only when it disarms.
	if newPtr == 0 {
		return negErrno(EFAULT)
	}
	interval, value, errno := readItimerval32(mem, newPtr)
	if errno != 0 {
		return negErrno(errno)
	}
	if interval != 0 || value != 0 {
		return enosys
	}
	return 0
}
