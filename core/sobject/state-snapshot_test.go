package sobject

import (
	"slices"
	"testing"
)

// TestSOStateClone checks immutable operation sharing and independent state edits.
func TestSOStateClone(t *testing.T) {
	// Copy every state field, including an unknown field retained on the wire.
	state := &SOState{
		Config:          &SharedObjectConfig{ConfigChainHash: []byte("config")},
		Checkpoint:      &SOCheckpoint{Inner: []byte("checkpoint")},
		KeyEpochs:       []*SOKeyEpoch{{Epoch: 1}},
		Ops:             []*SOOperation{{Inner: []byte("first")}, {Inner: []byte("second")}},
		Invites:         []*SOInvite{{}},
		Sequence:        []*SOSequence{{Inner: []byte("sequence")}},
		ControlMessages: []*SOControlMessage{{Inner: []byte("control")}},
		unknownFields:   []byte{0x40, 1},
	}
	before := state.CloneVT()
	next := state.Clone()
	if !next.EqualVT(before) || next.Ops[0] != state.Ops[0] || next.Ops[1] != state.Ops[1] {
		t.Fatal("copy changed state or copied immutable operations")
	}

	// Mutating the state copy never changes an earlier snapshot.
	next.Config.ConfigChainHash[0]++
	next.Checkpoint.Inner[0]++
	next.KeyEpochs[0].Epoch++
	next.Invites[0].Uses++
	next.Sequence[0].Inner[0]++
	next.ControlMessages[0].Inner[0]++
	next.unknownFields[1]++

	// Replace operation membership without editing any shared operation.
	next.Ops = slices.Delete(next.Ops, 0, 1)
	next.Ops = append(next.Ops, &SOOperation{Inner: []byte("third")})
	if !state.EqualVT(before) {
		t.Fatal("editing the copy changed the original state")
	}

	// Nil snapshots remain nil, and deep copies still isolate operation edits.
	if (*SOState)(nil).Clone() != nil {
		t.Fatal("copy of nil state is nonnil")
	}
	before.Ops[0].Inner[0]++
	if state.Ops[0].Inner[0] == before.Ops[0].Inner[0] {
		t.Fatal("deep copy shares operation bytes")
	}
}

// BenchmarkStateSnapshot measures copying a state and rebuilding its operation
// set through the verifier used across published snapshots.
func BenchmarkStateSnapshot(b *testing.B) {
	// Hold the same signed operations throughout the measurement.
	state, _ := newBenchState(b, 1024)

	// Measure a state copy separately from the cached set construction.
	b.Run("Clone", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if next := state.Clone(); !next.EqualVT(state) {
				b.Fatal("copy changed the state")
			}
		}
	})
	b.Run("OperationSet", func(b *testing.B) {
		// Prime the verifier before timing unchanged snapshots.
		verifier := NewSOOperationVerifier(mockSharedObjectID)
		if _, err := verifier.OperationSet(state); err != nil {
			b.Fatal(err)
		}

		// Build each snapshot with the already verified operations.
		b.ReportAllocs()
		for b.Loop() {
			set, err := verifier.OperationSet(state)
			if err != nil {
				b.Fatal(err)
			}
			if set.Len() != len(state.GetOps()) {
				b.Fatal("set lost operations")
			}
		}
	})
}
