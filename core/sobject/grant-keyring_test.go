package sobject

import "testing"

// TestSOGrantKeyringReuse checks that a keyring decrypts a grant once for the
// copies of a state, and decrypts a grant with other bytes again.
func TestSOGrantKeyringReuse(t *testing.T) {
	// Start from a state granting a key to two peers.
	peers := createMockPeers(t, 2)
	privs := mustPrivKeys(t, peers)
	state, _ := newTestSOState(t, peers)
	keyring := NewSOGrantKeyring(mockSharedObjectID)

	// A copy of the state reuses the first decryption.
	handle := NewSOStateParticipantHandle(nil, nil, mockSharedObjectID, state, privs[1], peers[1].GetPeerID())
	grant := handle.localGrant(state.CurrentKeyEpoch())
	first, err := keyring.Inner(grant, privs[1])
	if err != nil {
		t.Fatal(err)
	}
	copied := state.CloneVT()
	again, err := keyring.Inner(handle.localGrant(copied.CurrentKeyEpoch()), privs[1])
	if err != nil {
		t.Fatal(err)
	}
	if again != first {
		t.Fatal("copied grant decrypted again")
	}

	// A grant of a rotated key decrypts to its own key.
	_, epoch, err := RotateTransformKey(privs[0], mockSharedObjectID, state.GetConfig().GetParticipants(), 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	rotated, err := keyring.Inner(handle.localGrant(epoch), privs[1])
	if err != nil {
		t.Fatal(err)
	}
	if rotated == first || rotated.EqualVT(first) {
		t.Fatal("rotated grant returned the first key")
	}
}
