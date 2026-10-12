package kernel

import (
	"context"
	"testing"
)

// TestWait4ReapSecondChild checks wait4 removes the child it reaps, so a
// later wait4 for any child returns the next one rather than the first again.
func TestWait4ReapSecondChild(t *testing.T) {
	// Start a parent.
	ctx := context.Background()
	parent, _ := newMemProcess(wasmPageSize)
	k := parent.kernel

	// Reap one exited child by its PID, then another as any child.
	for _, wait := range []struct{ pid, waitPID int32 }{{2, 2}, {3, -1}} {
		child := newProcess(k, int(wait.pid), nil)
		child.ppid = parent.pid
		close(child.exitCh)
		k.procs.set(child.pid, child)
		if got := int32(parent.sysWait4(ctx, wait.waitPID, 0, 0, 0)); got != wait.pid {
			t.Fatalf("wait4(%d) = %d, want %d", wait.waitPID, got, wait.pid)
		}
		if k.procs.get(child.pid) != nil {
			t.Fatalf("wait4 left pid %d in the process table", child.pid)
		}
	}
}
