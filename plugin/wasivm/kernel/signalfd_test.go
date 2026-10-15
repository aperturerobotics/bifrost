package kernel

import (
	"context"
	"encoding/binary"
	"testing"
)

// TestSignalfd checks a signal file wakes an epoll wait when a blocked
// signal of its mask arrives, reads each pending signal as a record, and
// leaves the signals outside its mask pending.
func TestSignalfd(t *testing.T) {
	// Block SIGHUP, SIGINT and SIGTERM, and watch a file reading the last
	// two.
	p, mem := fsTestProcess()
	p.setSigMask(sigSetmask, maskSet(SIGINT, SIGTERM, SIGHUP))
	mem.Write(epollTestSpec, binary.LittleEndian.AppendUint64(nil, *maskSet(SIGINT, SIGTERM)))
	fd := uint32(p.sysSignalfd(-1, epollTestSpec, guestSfdNonblock|guestSfdCloexec))
	epfd := uint32(p.sysEpollCreate1(0))
	if res := epollTestCtl(p, mem, epfd, epollCtlAdd, fd, epollIn, 3); res != 0 {
		t.Fatalf("add = %#x", res)
	}

	// A signal outside the mask leaves the file idle, and one in it wakes
	// a blocked wait.
	p.deliverSignal(SIGHUP)
	if got := epollTestWait(t, p, mem, epfd, 0); len(got) != 0 {
		t.Fatalf("SIGHUP reported %v", got)
	}
	go p.deliverSignal(SIGTERM)
	if got := epollTestWait(t, p, mem, epfd, -1); len(got) != 1 || got[0] != (epollEvent{epollIn, 3}) {
		t.Fatalf("wait = %v, want IN with data 3", got)
	}

	// A read takes every pending signal of the mask, lowest first, then
	// would block.
	p.deliverSignal(SIGINT)
	data, errno := p.readFile(context.Background(), p.fds[fd], 4*signalfdSiginfoSize)
	if errno != 0 || len(data) != 2*signalfdSiginfoSize {
		t.Fatalf("read = %d bytes, %#x; want two records", len(data), errno)
	}
	first, second := binary.LittleEndian.Uint32(data), binary.LittleEndian.Uint32(data[signalfdSiginfoSize:])
	if first != SIGINT || second != SIGTERM {
		t.Fatalf("read signals %d and %d, want SIGINT and SIGTERM", first, second)
	}
	if _, errno := p.readFile(context.Background(), p.fds[fd], signalfdSiginfoSize); errno != negErrno(EAGAIN) {
		t.Fatalf("read of none = %#x, want EAGAIN", errno)
	}
	if p.sigPending != sigBit(SIGHUP) {
		t.Fatalf("pending %#x, want SIGHUP", p.sigPending)
	}
}
