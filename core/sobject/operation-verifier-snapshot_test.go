package sobject

import "testing"

// TestSOOperationVerifierSnapshots checks shared and independently decoded
// snapshots, changed operation bytes, and cache eviction after removal.
func TestSOOperationVerifierSnapshots(t *testing.T) {
	// Verify a signed operation before copying the state.
	peers := createMockPeers(t, 1)
	state, _ := newTestSOState(t, peers)
	op := writeTestOp(t, state, mustPrivKeys(t, peers)[0], "data")
	verifier := NewSOOperationVerifier(mockSharedObjectID)
	first, err := verifier.OperationSet(state)
	if err != nil {
		t.Fatal(err)
	}

	// Both copy forms preserve the verified body and operation identity.
	for _, next := range []*SOState{state.Clone(), state.CloneVT()} {
		set, err := verifier.OperationSet(next)
		if err != nil {
			t.Fatal(err)
		}
		if set.Get(op.Hash()) != first.Get(op.Hash()) || len(verifier.hashes) != 1 {
			t.Fatal("snapshot lost the verified operation or retained another hash")
		}
	}

	// Changed bytes in an independent operation still require authentication.
	forged := state.CloneVT()
	forged.Ops[0].Inner[len(forged.Ops[0].Inner)-1] ^= 1
	if _, err := verifier.OperationSet(forged); err == nil {
		t.Fatal("verifier accepted changed operation bytes")
	}

	// Removing operations drops their hashes as well as their verified bodies.
	next := state.Clone()
	next.Ops = nil
	if _, err := verifier.OperationSet(next); err != nil {
		t.Fatal(err)
	}
	if len(verifier.hashes) != 0 || len(verifier.verified) != 0 {
		t.Fatal("verifier retained removed operations")
	}
}
