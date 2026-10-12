package native

import (
	"github.com/tetratelabs/wazero/experimental"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// reusableSnapshot is a wazero snapshot that can be restored more than once.
// wazero installs a snapshot's saved stack as the live engine stack, so the
// first restore would corrupt the stack a later restore reuses. ash restores
// one setjmp snapshot from several vfork children in turn.
type reusableSnapshot struct {
	// snapshot is the captured state, never restored directly.
	snapshot experimental.Snapshot
}

// Restore resumes execution at a fresh copy of the capture.
func (s reusableSnapshot) Restore(returnValues []uint64) {
	cloneSnapshot(s.snapshot).Restore(returnValues)
}

// _ is a type assertion
var _ engine.Snapshot = reusableSnapshot{}
