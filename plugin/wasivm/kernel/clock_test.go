package kernel

import (
	"bytes"
	"testing"
	"time"
)

// TestSetitimerReal checks ITIMER_REAL sends SIGALRM when it expires and
// reports its old setting, while the CPU time timers can only be disarmed.
func TestSetitimerReal(t *testing.T) {
	// Arm a ten second timer, over an old value full of garbage.
	p, mem := newMemProcess(wasmPageSize)
	p.sigActions[SIGALRM] = sigAction{handler: 0x1000}
	const newPtr, oldPtr = 256, 512
	mem.WriteUint64Le(newPtr+timevalSize, 10)
	mem.Write(oldPtr, bytes.Repeat([]byte{0xff}, 2*timevalSize))
	if got := p.sysSetitimer(itimerReal, newPtr, oldPtr); got != 0 {
		t.Fatalf("arming setitimer = %#x, want 0", got)
	}
	if old, _ := mem.Read(oldPtr, 2*timevalSize); !bytes.Equal(old, make([]byte, 2*timevalSize)) {
		t.Fatalf("old itimerval = %x, want zero", old)
	}

	// Replace it with one expiring in a millisecond, which reports the ten
	// second timer, then wait for its signal.
	mem.WriteUint64Le(newPtr+timevalSize, 0)
	mem.WriteUint64Le(newPtr+timevalSize+8, 1000)
	if got := p.sysSetitimer(itimerReal, newPtr, oldPtr); got != 0 {
		t.Fatalf("rearming setitimer = %#x, want 0", got)
	}
	if sec, _ := mem.ReadUint64Le(oldPtr + timevalSize); sec != 9 && sec != 10 {
		t.Fatalf("old value %d s, want about 10", sec)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		p.sigMu.Lock()
		caught := p.sigCaught
		p.sigMu.Unlock()
		if caught&sigBit(SIGALRM) != 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("SIGALRM not caught")
		}
		time.Sleep(time.Millisecond)
	}

	// The expired timer reads disarmed, and a CPU time timer cannot arm.
	if got := p.sysGetitimer(itimerReal, oldPtr); got != 0 {
		t.Fatalf("getitimer = %#x, want 0", got)
	}
	if old, _ := mem.Read(oldPtr, 2*timevalSize); !bytes.Equal(old, make([]byte, 2*timevalSize)) {
		t.Fatalf("expired itimerval = %x, want zero", old)
	}
	if got := p.sysSetitimer(itimerVirtual, newPtr, 0); got != enosys {
		t.Fatalf("arming ITIMER_VIRTUAL = %#x, want ENOSYS", got)
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
