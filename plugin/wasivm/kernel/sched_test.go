package kernel

import (
	"testing"
)

func TestPriorityRoundTrip(t *testing.T) {
	p, _ := newMemProcess(wasmPageSize)
	if got := p.sysSetpriority(prioProcess, 0, uint32(0xfffffff6)); got != 0 {
		t.Fatalf("setpriority failed: %#x", got)
	}
	if got := uint32(p.sysGetpriority(prioProcess, 0)); got != 0xfffffff6 {
		t.Fatalf("getpriority = %#x, want -10", got)
	}
}
