package resource_sobject

import (
	"testing"

	"github.com/s4wave/spacewave/core/sobject"
)

func TestAuthoritativeSyncDenied(t *testing.T) {
	config := &sobject.SharedObjectConfig{Participants: []*sobject.SOParticipantConfig{
		{PeerId: "owner", Role: sobject.SOParticipantRole_SOParticipantRole_OWNER},
		{PeerId: "writer", Role: sobject.SOParticipantRole_SOParticipantRole_WRITER},
	}}
	if sobject.AuthoritativeSyncDenied(config, &sobject.SharedObjectHealth{SyncDeniedPeerIds: []string{"writer"}}) {
		t.Fatal("writer denial revoked another participant")
	}
	if !sobject.AuthoritativeSyncDenied(config, &sobject.SharedObjectHealth{SyncDeniedPeerIds: []string{"owner"}}) {
		t.Fatal("owner denial did not revoke live body access")
	}
}
