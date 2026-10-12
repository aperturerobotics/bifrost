package kernel

import "testing"

// TestSetitimerRejectsArming checks setitimer fails to arm a timer, which the
// kernel does not support, and still disarms one.
func TestSetitimerRejectsArming(t *testing.T) {
	// Arm a one second interval.
	p, mem := newMemProcess(wasmPageSize)
	const newPtr, oldPtr = 256, 512
	mem.WriteUint32Le(newPtr+8, 1)
	if got := p.sysSetitimer(itimerReal, newPtr, oldPtr); got != enosys {
		t.Fatalf("arming setitimer = %#x, want ENOSYS", got)
	}
	if old, _ := mem.ReadUint64Le(oldPtr); old != 0 {
		t.Fatalf("old interval = %#x, want zero", old)
	}

	// Disarm it.
	mem.WriteUint32Le(newPtr+8, 0)
	if got := p.sysSetitimer(itimerReal, newPtr, oldPtr); got != 0 {
		t.Fatalf("disarming setitimer = %#x, want 0", got)
	}
}
