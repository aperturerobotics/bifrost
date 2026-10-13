package kernel

import (
	"bytes"
	"testing"
	"time"
)

// TestSetitimerRejectsArming checks setitimer fails to arm a timer, which the
// kernel does not support, and still disarms one.
func TestSetitimerRejectsArming(t *testing.T) {
	// Arm a one second timer, over an old value full of garbage.
	p, mem := newMemProcess(wasmPageSize)
	const newPtr, oldPtr = 256, 512
	mem.WriteUint64Le(newPtr+timevalSize, 1)
	mem.Write(oldPtr, bytes.Repeat([]byte{0xff}, 2*timevalSize))
	if got := p.sysSetitimer(itimerReal, newPtr, oldPtr); got != enosys {
		t.Fatalf("arming setitimer = %#x, want ENOSYS", got)
	}
	if old, _ := mem.Read(oldPtr, 2*timevalSize); !bytes.Equal(old, make([]byte, 2*timevalSize)) {
		t.Fatalf("old itimerval = %x, want zero", old)
	}

	// Disarm it.
	mem.WriteUint64Le(newPtr+timevalSize, 0)
	if got := p.sysSetitimer(itimerReal, newPtr, oldPtr); got != 0 {
		t.Fatalf("disarming setitimer = %#x, want 0", got)
	}
}

// TestClockSettime checks clock_settime reads a wasm32 struct timespec, a
// 64-bit tv_sec and then tv_nsec.
func TestClockSettime(t *testing.T) {
	// Set the clock to a time past 2038, which needs the 64-bit field.
	p, mem := newMemProcess(wasmPageSize)
	const tsPtr = 256
	want := time.Unix(1<<32, 5)
	mem.WriteUint64Le(tsPtr, uint64(want.Unix()))
	mem.WriteUint32Le(tsPtr+8, 5)
	if res := p.sysClockSettime(clockRealtime, tsPtr); res != 0 {
		t.Fatalf("clock_settime = %#x", res)
	}
	if got := p.kernel.realtimeNow(); got.Sub(want).Abs() > time.Minute {
		t.Fatalf("realtime = %v, want %v", got, want)
	}
}
