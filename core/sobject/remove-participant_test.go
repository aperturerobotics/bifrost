package sobject

import (
	"context"
	"testing"
)

// TestRemoveSOParticipantsAtomicallyRemovesAudience verifies that one owner
// transition removes a person's complete participant set.
func TestRemoveSOParticipantsAtomicallyRemovesAudience(t *testing.T) {
	ctx := context.Background()
	peers := createMockPeers(t, 3)
	owner, err := peers[0].GetPrivKey(ctx)
	if err != nil {
		t.Fatal(err)
	}
	initial := &SharedObjectConfig{Participants: []*SOParticipantConfig{
		{PeerId: peers[0].GetPeerID().String(), Role: SOParticipantRole_SOParticipantRole_OWNER},
		{PeerId: peers[1].GetPeerID().String(), Role: SOParticipantRole_SOParticipantRole_WRITER},
		{PeerId: peers[2].GetPeerID().String(), Role: SOParticipantRole_SOParticipantRole_WRITER},
	}}
	genesis, err := BuildSOConfigChange(initial, initial, SOConfigChangeType_SO_CONFIG_CHANGE_TYPE_GENESIS, owner, nil)
	if err != nil {
		t.Fatal(err)
	}
	checkpoint, err := VerifyConfigChange(initial, genesis)
	if err != nil {
		t.Fatal(err)
	}
	host, state := newLeaveTestHost(ctx, checkpoint)

	removed, err := RemoveSOParticipants(ctx, host, []string{
		peers[1].GetPeerID().String(), peers[2].GetPeerID().String(),
	}, owner, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 2 {
		t.Fatalf("removed %d participants, want 2", len(removed))
	}
	current := (*state).GetConfig()
	if len(current.GetParticipants()) != 1 || current.GetParticipants()[0].GetPeerId() != peers[0].GetPeerID().String() {
		t.Fatalf("atomic removal left audience %v", current.GetParticipants())
	}
	if current.GetConfigChainSeqno() != checkpoint.GetConfigChainSeqno()+1 {
		t.Fatalf("atomic removal advanced configuration by %d changes", current.GetConfigChainSeqno()-checkpoint.GetConfigChainSeqno())
	}
}
