package kernel

import (
	"context"
	"encoding/binary"
	"math"
	"testing"
	"time"
)

// Guest memory addresses the epoll tests use.
const (
	epollTestPipe   = 1024
	epollTestEvent  = 2048
	epollTestEvents = 4096
	epollTestSpec   = 6144
)

// epollTestWait waits on epfd for up to timeout milliseconds and returns the
// events it reports.
func epollTestWait(t *testing.T, p *Process, mem *memMemory, epfd uint32, timeout int32) []epollEvent {
	// Wait.
	t.Helper()
	n := p.sysEpollWait(context.Background(), epfd, epollTestEvents, 8, timeout)
	if errnoFailed(n) {
		t.Fatalf("epoll_wait = %#x", n)
	}

	// Decode the events.
	events := make([]epollEvent, n)
	buf, _ := mem.Read(epollTestEvents, uint32(n)*epollEventSize)
	for i := range events {
		events[i].events = binary.LittleEndian.Uint32(buf[i*epollEventSize:])
		events[i].data = binary.LittleEndian.Uint64(buf[i*epollEventSize+8:])
	}
	return events
}

// epollTestCtl applies op for fd with events and data to epfd, returning the
// result.
func epollTestCtl(p *Process, mem *memMemory, epfd, op, fd, events uint32, data uint64) uint64 {
	ev := binary.LittleEndian.AppendUint32(nil, events)
	ev = binary.LittleEndian.AppendUint64(append(ev, 0, 0, 0, 0), data)
	mem.Write(epollTestEvent, ev)
	return p.sysEpollCtl(epfd, op, fd, epollTestEvent)
}

// counterTestRead reads fd as an 8-byte count, returning the count or the
// errno.
func counterTestRead(p *Process, fd uint32) (uint64, uint64) {
	data, errno := p.readFile(context.Background(), p.fds[fd], 8)
	if errno != 0 {
		return 0, errno
	}
	return binary.LittleEndian.Uint64(data), 0
}

// TestEpollPipe checks an epoll set reports a pipe's data and hangup, level
// triggered and edge triggered, and drops the pipe once it closes.
func TestEpollPipe(t *testing.T) {
	// Watch a pipe's read end.
	p, mem := fsTestProcess()
	epfd := uint32(p.sysEpollCreate1(guestEpollCloexec))
	p.sysPipe2(epollTestPipe, 0)
	rfd, _ := mem.ReadUint32Le(epollTestPipe)
	wfd, _ := mem.ReadUint32Le(epollTestPipe + 4)
	if res := epollTestCtl(p, mem, epfd, epollCtlAdd, rfd, epollIn, 7); res != 0 {
		t.Fatalf("add = %#x", res)
	}
	if res := epollTestCtl(p, mem, epfd, epollCtlAdd, rfd, epollIn, 7); res != negErrno(EEXIST) {
		t.Fatalf("second add = %#x, want EEXIST", res)
	}
	if got := epollTestWait(t, p, mem, epfd, 0); len(got) != 0 {
		t.Fatalf("empty pipe reported %v", got)
	}

	// Data reports on every wait while level triggered.
	p.writeOutBytes(context.Background(), wfd, []byte("x"))
	for range 2 {
		if got := epollTestWait(t, p, mem, epfd, 0); len(got) != 1 || got[0] != (epollEvent{epollIn, 7}) {
			t.Fatalf("level-triggered wait = %v, want IN with data 7", got)
		}
	}

	// Edge triggered, it reports once until the pipe changes.
	if res := epollTestCtl(p, mem, epfd, epollCtlMod, rfd, epollIn|epollET, 8); res != 0 {
		t.Fatalf("modify = %#x", res)
	}
	if got := epollTestWait(t, p, mem, epfd, 0); len(got) != 1 {
		t.Fatalf("edge-triggered wait = %v, want one event", got)
	}
	if got := epollTestWait(t, p, mem, epfd, 0); len(got) != 0 {
		t.Fatalf("unchanged edge-triggered wait = %v, want none", got)
	}

	// Closing the writer reports a hangup.
	p.sysClose(wfd)
	if got := epollTestWait(t, p, mem, epfd, 0); len(got) != 1 || got[0].events != epollIn|epollHup {
		t.Fatalf("wait after hangup = %v, want IN and HUP", got)
	}

	// Closing the reader drops the entry.
	p.sysClose(rfd)
	if got := epollTestWait(t, p, mem, epfd, 0); len(got) != 0 {
		t.Fatalf("wait after close = %v, want none", got)
	}
	if n := len(p.fds[epfd].epoll.entries); n != 0 {
		t.Fatalf("%d entries after close, want none", n)
	}
}

// TestEpollRejectsRegularFile checks a regular file cannot join a set, as on
// Linux.
func TestEpollRejectsRegularFile(t *testing.T) {
	p, mem := fsTestProcess()
	epfd := uint32(p.sysEpollCreate1(0))
	fd, _ := openTestFile(p, "f", nil)
	if res := epollTestCtl(p, mem, epfd, epollCtlAdd, fd, epollIn, 0); res != negErrno(EPERM) {
		t.Fatalf("add of a regular file = %#x, want EPERM", res)
	}
}

// TestEventfd checks an event file's reads take its count, one at a time
// for a semaphore, and that it refuses the reserved count.
func TestEventfd(t *testing.T) {
	// A counter reads whole, then would block.
	p, _ := fsTestProcess()
	fd := uint32(p.sysEventfd(3, guestEfdNonblock))
	p.writeOutBytes(context.Background(), fd, binary.LittleEndian.AppendUint64(nil, 2))
	if n, errno := counterTestRead(p, fd); n != 5 || errno != 0 {
		t.Fatalf("read = %d, %#x; want 5", n, errno)
	}
	if _, errno := counterTestRead(p, fd); errno != negErrno(EAGAIN) {
		t.Fatalf("read of zero = %#x, want EAGAIN", errno)
	}
	if res := p.writeOutBytes(context.Background(), fd, binary.LittleEndian.AppendUint64(nil, math.MaxUint64)); res != negErrno(EINVAL) {
		t.Fatalf("write of the reserved count = %#x, want EINVAL", res)
	}

	// A semaphore reads one at a time.
	sem := uint32(p.sysEventfd(2, guestEfdSemaphore|guestEfdNonblock))
	for range 2 {
		if n, errno := counterTestRead(p, sem); n != 1 || errno != 0 {
			t.Fatalf("semaphore read = %d, %#x; want 1", n, errno)
		}
	}
	if _, errno := counterTestRead(p, sem); errno != negErrno(EAGAIN) {
		t.Fatalf("read of an empty semaphore = %#x, want EAGAIN", errno)
	}
}

// TestTimerfd checks a periodic timer wakes a blocked epoll wait, counts its
// expirations, reports its setting and disarms.
func TestTimerfd(t *testing.T) {
	// Arm a 5ms periodic timer and watch it.
	p, mem := fsTestProcess()
	fd := uint32(p.sysTimerfdCreate(clockMonotonic, guestTfdNonblock|guestTfdCloexec))
	epfd := uint32(p.sysEpollCreate1(0))
	if res := epollTestCtl(p, mem, epfd, epollCtlAdd, fd, epollIn, 9); res != 0 {
		t.Fatalf("add = %#x", res)
	}
	writeItimerspec(mem, epollTestSpec, 5*time.Millisecond, 5*time.Millisecond)
	if res := p.sysTimerfdSettime(fd, 0, epollTestSpec, 0); res != 0 {
		t.Fatalf("settime = %#x", res)
	}

	// The wait blocks until the first expiration.
	if got := epollTestWait(t, p, mem, epfd, -1); len(got) != 1 || got[0] != (epollEvent{epollIn, 9}) {
		t.Fatalf("wait = %v, want IN with data 9", got)
	}
	if n, errno := counterTestRead(p, fd); n == 0 || errno != 0 {
		t.Fatalf("read = %d, %#x; want expirations", n, errno)
	}

	// The setting reads back.
	p.sysTimerfdGettime(fd, epollTestSpec)
	interval, value, _ := readItimerspec(mem, epollTestSpec)
	if interval != 5*time.Millisecond || value <= 0 || value > 5*time.Millisecond {
		t.Fatalf("gettime = %v, %v; want 5ms and the time left", interval, value)
	}

	// Disarming drops the expirations and stops the timer.
	writeItimerspec(mem, epollTestSpec, 0, 0)
	p.sysTimerfdSettime(fd, 0, epollTestSpec, 0)
	if _, errno := counterTestRead(p, fd); errno != negErrno(EAGAIN) {
		t.Fatalf("read after disarm = %#x, want EAGAIN", errno)
	}
	if got := epollTestWait(t, p, mem, epfd, 20); len(got) != 0 {
		t.Fatalf("disarmed timer reported %v", got)
	}
}

// TestEpollPwait checks a pending signal the call's mask unblocks ends the
// wait with EINTR, caught, and the old mask returns after it.
func TestEpollPwait(t *testing.T) {
	// Hold SIGINT, which has a handler, pending behind the mask.
	p, mem := fsTestProcess()
	p.sigActions[SIGINT] = sigAction{handler: 0x1000}
	p.setSigMask(sigSetmask, maskSet(SIGINT))
	p.deliverSignal(SIGINT)
	epfd := uint32(p.sysEpollCreate1(0))

	// A wait keeping the mask times out with SIGINT still pending.
	if n := p.sysEpollPwait(context.Background(), epfd, epollTestEvents, 8, 0, 0); n != 0 {
		t.Fatalf("epoll_pwait with the mask kept = %#x, want 0", n)
	}
	if p.sigPending != sigBit(SIGINT) {
		t.Fatalf("pending %#x, want SIGINT", p.sigPending)
	}

	// A wait with an empty mask catches it at once and restores the mask.
	mem.Write(epollTestSpec, make([]byte, 8))
	if n := p.sysEpollPwait(context.Background(), epfd, epollTestEvents, 8, -1, epollTestSpec); n != negErrno(EINTR) {
		t.Fatalf("epoll_pwait with an empty mask = %#x, want EINTR", n)
	}
	if p.sigMask != sigBit(SIGINT) || p.sigCaught != sigBit(SIGINT) {
		t.Fatalf("after the wait: mask %#x, caught %#x", p.sigMask, p.sigCaught)
	}
}
