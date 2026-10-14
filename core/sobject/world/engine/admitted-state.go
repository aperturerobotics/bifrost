package sobject_world_engine

import (
	"context"

	"github.com/aperturerobotics/util/ccontainer"
	"github.com/s4wave/spacewave/core/sobject"
)

// accessAdmittedState adds a reference to the state the engine replays and
// returns its container: the admitted state of a StatePublisher, which holds
// the operations admitted to pending publications, or else the shared object
// state.
func accessAdmittedState(ctx context.Context, so sobject.SharedObject, released func()) (ccontainer.Watchable[sobject.SharedObjectStateSnapshot], func(), error) {
	if publisher, ok := so.(sobject.StatePublisher); ok {
		return publisher.AccessAdmittedState(ctx, released)
	}
	return so.AccessSharedObjectState(ctx, released)
}

// getAdmittedState returns a snapshot of the state the engine replays.
func getAdmittedState(ctx context.Context, so sobject.SharedObject) (sobject.SharedObjectStateSnapshot, error) {
	// Retain the state until it supplies a snapshot.
	ctr, rel, err := accessAdmittedState(ctx, so, nil)
	if err != nil {
		return nil, err
	}
	defer rel()

	// Return the current snapshot once one is available.
	return ctr.WaitValue(ctx, nil)
}
