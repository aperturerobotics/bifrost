package kernel

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"maps"
	"math"
	goruntime "runtime"
	"slices"
	"time"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// Every process starts with the root directory open as descriptor 3, the
// one directory it finds with fd_prestat_get.
const (
	wasiPreopenRootFD = 3
	wasiPreopenRoot   = "/"
)

// WASI errno values the kernel returns directly. wasiErrno translates the
// rest from Linux.
const (
	wasiErrnoSuccess  = 0
	wasiErrnoBadf     = 8
	wasiErrnoFault    = 21
	wasiErrnoInval    = 28
	wasiErrnoIo       = 29
	wasiErrnoNosys    = 52
	wasiErrnoOverflow = 61
)

// poll_oneoff event types, the subscription flag for an absolute clock
// timeout, and the sizes of a subscription and an event.
const (
	wasiEventtypeClock      = 0
	wasiEventtypeFDRead     = 1
	wasiEventtypeFDWrite    = 2
	wasiSubclockAbstime     = 1
	wasiSubscriptionSize    = 48
	wasiEventSize           = 32
	wasiMaxDurationNanosecs = math.MaxInt64
)

// wasiFunctions returns the wasi_snapshot_preview1 module.
func wasiFunctions() map[string]hostFunc {
	fns := map[string]hostFunc{
		"args_get": wasiFunc("iii", func(_ context.Context, p *Process, mem engine.Memory, a []uint64) uint32 {
			return wasiWriteStrings(mem, p.args, uint32(a[0]), uint32(a[1]))
		}),
		"args_sizes_get": wasiFunc("iii", func(_ context.Context, p *Process, mem engine.Memory, a []uint64) uint32 {
			return wasiWriteStringSizes(mem, p.args, uint32(a[0]), uint32(a[1]))
		}),
		"clock_res_get": wasiFunc("iii", func(_ context.Context, p *Process, mem engine.Memory, a []uint64) uint32 {
			return p.wasiClockResGet(mem, uint32(a[0]), uint32(a[1]))
		}),
		"clock_time_get": wasiFunc("iiji", func(_ context.Context, p *Process, mem engine.Memory, a []uint64) uint32 {
			return p.wasiClockTimeGet(mem, uint32(a[0]), uint32(a[2]))
		}),
		"environ_get": wasiFunc("iii", func(_ context.Context, p *Process, mem engine.Memory, a []uint64) uint32 {
			return wasiWriteStrings(mem, p.wasiEnviron(), uint32(a[0]), uint32(a[1]))
		}),
		"environ_sizes_get": wasiFunc("iii", func(_ context.Context, p *Process, mem engine.Memory, a []uint64) uint32 {
			return wasiWriteStringSizes(mem, p.wasiEnviron(), uint32(a[0]), uint32(a[1]))
		}),
		"poll_oneoff": wasiFunc("iiiii", func(ctx context.Context, p *Process, mem engine.Memory, a []uint64) uint32 {
			return p.wasiPollOneoff(ctx, mem, uint32(a[0]), uint32(a[1]), uint32(a[2]), uint32(a[3]))
		}),
		"proc_exit": {"vi", exitFunc},
		"random_get": wasiFunc("iii", func(_ context.Context, _ *Process, mem engine.Memory, a []uint64) uint32 {
			return wasiRandomGet(mem, uint32(a[0]), uint32(a[1]))
		}),
		"sched_yield": wasiFunc("i", func(context.Context, *Process, engine.Memory, []uint64) uint32 {
			goruntime.Gosched()
			return wasiErrnoSuccess
		}),
		"sock_accept": wasiFunc("iiii", func(context.Context, *Process, engine.Memory, []uint64) uint32 {
			return wasiErrnoNosys
		}),
	}
	maps.Copy(fns, wasiFSFunctions())
	return fns
}

// wasiFunc builds a host function from a WASI call that returns an errno.
func wasiFunc(sig string, call func(ctx context.Context, p *Process, mem engine.Memory, a []uint64) uint32) hostFunc {
	return hostFunc{sig, func(ctx context.Context, p *Process, a []uint64) uint64 {
		return uint64(call(ctx, p, p.inst.Memory(), a))
	}}
}

// wasiEnviron returns the environment as KEY=VALUE strings sorted by key.
func (p *Process) wasiEnviron() []string {
	env := make([]string, 0, len(p.env))
	for _, key := range slices.Sorted(maps.Keys(p.env)) {
		env = append(env, key+"="+p.env[key])
	}
	return env
}

// wasiClockTimeGet writes clock id's time in nanoseconds to timePtr.
func (p *Process) wasiClockTimeGet(mem engine.Memory, id, timePtr uint32) uint32 {
	t, ok := p.kernel.nowForClock(id)
	switch {
	case !ok:
		return wasiErrnoInval
	case !mem.WriteUint64Le(timePtr, uint64(t.UnixNano())):
		return wasiErrnoFault
	}
	return wasiErrnoSuccess
}

// wasiClockResGet writes clock id's resolution in nanoseconds to resPtr.
func (p *Process) wasiClockResGet(mem engine.Memory, id, resPtr uint32) uint32 {
	sec, nsec, ok := clockResolution(id)
	switch {
	case !ok:
		return wasiErrnoInval
	case !mem.WriteUint64Le(resPtr, uint64(sec)*uint64(time.Second)+uint64(nsec)):
		return wasiErrnoFault
	}
	return wasiErrnoSuccess
}

// wasiRandomGet fills bufLen bytes at bufPtr with random data.
func wasiRandomGet(mem engine.Memory, bufPtr, bufLen uint32) uint32 {
	buf := make([]byte, bufLen)
	_, _ = rand.Read(buf)
	if !mem.Write(bufPtr, buf) {
		return wasiErrnoFault
	}
	return wasiErrnoSuccess
}

// wasiPollEvent is an event poll_oneoff reports.
type wasiPollEvent struct {
	userdata  uint64
	errno     uint16
	eventType uint8
	// nbytes is the bytes a descriptor can read or write.
	nbytes uint64
}

// wasiPoll is the state of a poll_oneoff call after reading its
// subscriptions.
type wasiPoll struct {
	// ready holds the events that have already happened.
	ready []wasiPollEvent
	// stdin is a read event waiting for console input, which stdinCh
	// signals.
	stdin   *wasiPollEvent
	stdinCh <-chan struct{}
	// clock is the earliest clock event not yet due, due after wait.
	clock *wasiPollEvent
	wait  time.Duration
}

// wasiPollOneoff writes to outPtr the events of nsubscriptions subscriptions
// at inPtr that happen first, and their count to neventsPtr. Descriptors are
// always ready except a console with no input, so the call waits only for
// console input or a clock. A caught signal ends the wait with EINTR.
func (p *Process) wasiPollOneoff(ctx context.Context, mem engine.Memory, inPtr, outPtr, nsubscriptions, neventsPtr uint32) uint32 {
	// Report the events that already happened.
	if nsubscriptions == 0 {
		return wasiErrnoInval
	}
	poll, errno := p.wasiReadPoll(mem, inPtr, nsubscriptions)
	if errno != wasiErrnoSuccess {
		return errno
	}
	if len(poll.ready) != 0 || (poll.stdin == nil && poll.clock == nil) {
		return wasiWritePollEvents(mem, poll.ready, outPtr, neventsPtr)
	}

	// Wait for console input or the clock. A nil channel never fires.
	ctx, release := p.interruptible(ctx)
	defer release()
	var stdinCh <-chan struct{}
	if poll.stdin != nil {
		stdinCh = poll.stdinCh
	}
	var timerCh <-chan time.Time
	if poll.clock != nil {
		timer := time.NewTimer(poll.wait)
		defer timer.Stop()
		timerCh = timer.C
	}
	select {
	case <-ctx.Done():
		return wasiErrno(EINTR)
	case <-stdinCh:
		return wasiWritePollEvents(mem, []wasiPollEvent{*poll.stdin}, outPtr, neventsPtr)
	case <-timerCh:
		return wasiWritePollEvents(mem, []wasiPollEvent{*poll.clock}, outPtr, neventsPtr)
	}
}

// wasiReadPoll reads n poll_oneoff subscriptions at inPtr into the events
// ready now and the ones to wait for.
func (p *Process) wasiReadPoll(mem engine.Memory, inPtr, n uint32) (*wasiPoll, uint32) {
	stdinReady, stdinCh := p.kernel.stdin.ready()
	poll := &wasiPoll{stdinCh: stdinCh}
	for i := range n {
		// Read the subscription's header.
		subPtr := inPtr + i*wasiSubscriptionSize
		userdata, ok1 := mem.ReadUint64Le(subPtr)
		eventType, ok2 := mem.ReadByteAt(subPtr + 8)
		fd, ok3 := mem.ReadUint32Le(subPtr + 16)
		if !ok1 || !ok2 || !ok3 {
			return nil, wasiErrnoFault
		}
		event := wasiPollEvent{userdata: userdata, eventType: eventType}

		// File a clock event as ready or as the earliest wait.
		if eventType == wasiEventtypeClock {
			wait, errno := p.wasiPollClockWait(mem, subPtr)
			switch {
			case errno != wasiErrnoSuccess:
				return nil, errno
			case wait <= 0:
				poll.ready = append(poll.ready, event)
			case poll.clock == nil || wait < poll.wait:
				poll.clock, poll.wait = &event, wait
			}
			continue
		}

		// File a descriptor event as ready unless it reads a console with no
		// input.
		event.nbytes = 1
		_, open := p.fds[fd]
		switch {
		case eventType != wasiEventtypeFDRead && eventType != wasiEventtypeFDWrite:
			return nil, wasiErrnoInval
		case !open:
			event.errno = wasiErrnoBadf
			poll.ready = append(poll.ready, event)
		case eventType == wasiEventtypeFDRead && !stdinReady && p.isConsoleFD(fd):
			poll.stdin = &event
		default:
			poll.ready = append(poll.ready, event)
		}
	}
	return poll, wasiErrnoSuccess
}

// wasiPollClockWait returns how long until the clock subscription at subPtr
// is due.
func (p *Process) wasiPollClockWait(mem engine.Memory, subPtr uint32) (time.Duration, uint32) {
	// Read the clock, the timeout and the flags.
	id, ok1 := mem.ReadUint32Le(subPtr + 16)
	timeout, ok2 := mem.ReadUint64Le(subPtr + 24)
	flags, ok3 := mem.ReadUint32Le(subPtr + 40)
	if !ok1 || !ok2 || !ok3 {
		return 0, wasiErrnoFault
	}
	if flags&wasiSubclockAbstime == 0 {
		return time.Duration(min(timeout, wasiMaxDurationNanosecs)), wasiErrnoSuccess
	}

	// Measure an absolute timeout from the clock's current time.
	var now uint64
	switch id {
	case clockRealtime:
		now = uint64(p.kernel.realtimeNow().UnixNano())
	case clockMonotonic, clockProcessCPUTimeID, clockThreadCPUTimeID:
		now = uint64(p.kernel.monotonicElapsed())
	default:
		return 0, wasiErrnoInval
	}
	if timeout <= now {
		return 0, wasiErrnoSuccess
	}
	return time.Duration(min(timeout-now, wasiMaxDurationNanosecs)), wasiErrnoSuccess
}

// wasiWritePollEvents writes events to outPtr and their count to neventsPtr.
func wasiWritePollEvents(mem engine.Memory, events []wasiPollEvent, outPtr, neventsPtr uint32) uint32 {
	for i, event := range events {
		var buf [wasiEventSize]byte
		binary.LittleEndian.PutUint64(buf[0:], event.userdata)
		binary.LittleEndian.PutUint16(buf[8:], event.errno)
		buf[10] = event.eventType
		binary.LittleEndian.PutUint64(buf[16:], event.nbytes)
		if !mem.Write(outPtr+uint32(i)*wasiEventSize, buf[:]) {
			return wasiErrnoFault
		}
	}
	if !mem.WriteUint32Le(neventsPtr, uint32(len(events))) {
		return wasiErrnoFault
	}
	return wasiErrnoSuccess
}

// wasiWriteStringSizes writes the count of values to countPtr and the bytes
// they take as NUL-terminated strings to bufSizePtr.
func wasiWriteStringSizes(mem engine.Memory, values []string, countPtr, bufSizePtr uint32) uint32 {
	// Total the strings.
	var bufSize uint64
	for _, value := range values {
		bufSize += uint64(len(value)) + 1
	}
	if bufSize > math.MaxUint32 {
		return wasiErrnoOverflow
	}

	// Write the sizes.
	if !mem.WriteUint32Le(countPtr, uint32(len(values))) || !mem.WriteUint32Le(bufSizePtr, uint32(bufSize)) {
		return wasiErrnoFault
	}
	return wasiErrnoSuccess
}

// wasiWriteStrings writes values as NUL-terminated strings to bufPtr and a
// pointer to each to the array at ptrsPtr.
func wasiWriteStrings(mem engine.Memory, values []string, ptrsPtr, bufPtr uint32) uint32 {
	off := bufPtr
	for i, value := range values {
		if !mem.WriteUint32Le(ptrsPtr+uint32(i)*4, off) || !mem.Write(off, append([]byte(value), 0)) {
			return wasiErrnoFault
		}
		off += uint32(len(value) + 1)
	}
	return wasiErrnoSuccess
}

// wasiReadPath reads a path of length bytes at ptr.
func wasiReadPath(mem engine.Memory, ptr, length uint32) (string, uint32) {
	buf, ok := mem.Read(ptr, length)
	if !ok {
		return "", wasiErrnoFault
	}
	return string(buf), wasiErrnoSuccess
}
