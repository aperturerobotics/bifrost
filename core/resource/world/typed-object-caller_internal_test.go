package resource_world

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/s4wave/spacewave/sdk/world/objecttype"
)

// TestBindCallerDispatchesNonParticipantWithoutParticipant proves a caller
// that is not an accepted participant still reaches typed dispatch, so each
// consumer denies its mutations instead of failing the typed object.
func TestBindCallerDispatchesNonParticipantWithoutParticipant(t *testing.T) {
	// Mount a typed resource whose participant lookup rejects the caller.
	caller, err := peer.NewPeer(nil)
	if err != nil {
		t.Fatal(err)
	}
	r := &TypedObjectResource{
		participantLookup: func(context.Context, peer.ID) (*sobject.SOParticipantConfig, error) {
			return nil, sobject.ErrNotParticipant
		},
	}
	ctx := objecttype.WithCaller(t.Context(), objecttype.Caller{PeerID: caller.GetPeerID().String()})

	// Bind the non-participant caller for dispatch.
	ctx, err = r.bindCaller(ctx)
	if err != nil {
		t.Fatal(err)
	}
	bound := objecttype.CallerFromContext(ctx)
	if bound.PeerID != caller.GetPeerID().String() || bound.Participant != nil {
		t.Fatalf("caller = %+v, want the peer without a participant", bound)
	}
}
