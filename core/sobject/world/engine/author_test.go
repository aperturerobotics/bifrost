package sobject_world_engine

import (
	"testing"

	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/world"
)

// TestOperationAuthor binds two signing devices to one accepted entity, falls
// back to the signing device for a local-only participant, and marks owners.
func TestOperationAuthor(t *testing.T) {
	for _, test := range []struct {
		participant *sobject.SOParticipantConfig
		person      string
		owner       bool
	}{
		{&sobject.SOParticipantConfig{PeerId: "device-a", EntityId: "person", Role: sobject.SOParticipantRole_SOParticipantRole_OWNER}, "person", true},
		{&sobject.SOParticipantConfig{PeerId: "device-b", EntityId: "person", Role: sobject.SOParticipantRole_SOParticipantRole_WRITER}, "person", false},
		{&sobject.SOParticipantConfig{PeerId: "local-only", Role: sobject.SOParticipantRole_SOParticipantRole_WRITER}, "local-only", false},
	} {
		ctx := withOperationAuthor(t.Context(), test.participant)
		if person, owner := world.OperationPersonFromContext(ctx), world.OperationOwnerFromContext(ctx); person != test.person || owner != test.owner {
			t.Fatalf("author for %s = %q, %v", test.participant.GetPeerId(), person, owner)
		}
	}
}
