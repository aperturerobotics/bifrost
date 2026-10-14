package sobject_world_engine

import (
	"context"

	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/world"
)

// participantPerson names the person a participant acts as: its entity, or its
// signing device for a participant without one.
func participantPerson(participant *sobject.SOParticipantConfig) string {
	if person := participant.GetEntityId(); person != "" {
		return person
	}
	return participant.GetPeerId()
}

// withOperationAuthor binds the person a participant's operation runs as and
// whether the participant owns the shared object.
func withOperationAuthor(ctx context.Context, participant *sobject.SOParticipantConfig) context.Context {
	ctx = world.WithOperationPerson(ctx, participantPerson(participant))
	return world.WithOperationOwner(ctx, sobject.IsOwner(participant.GetRole()))
}
