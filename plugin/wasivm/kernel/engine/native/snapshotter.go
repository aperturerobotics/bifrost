package native

import (
	"github.com/tetratelabs/wazero/experimental"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine"
)

// snapshotsKey marks a context passed through Runtime.WithSnapshots.
type snapshotsKey struct{}

// snapshotter captures the state of a wazero host call.
type snapshotter struct {
	// s is the wazero snapshotter of the host call.
	s experimental.Snapshotter
}

// Snapshot captures a state that can be restored any number of times.
func (s *snapshotter) Snapshot() engine.Snapshot {
	return reusableSnapshot{snapshot: s.s.Snapshot()}
}

// SnapshotOnce captures a state that is restored at most once.
func (s *snapshotter) SnapshotOnce() engine.Snapshot {
	return s.s.Snapshot()
}

// _ is a type assertion
var _ engine.Snapshotter = (*snapshotter)(nil)
