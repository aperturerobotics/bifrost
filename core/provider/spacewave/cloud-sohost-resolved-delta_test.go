package provider_spacewave

import (
	"testing"

	api "github.com/s4wave/spacewave/core/provider/spacewave/api"
	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/net/peer"
)

// TestDelayedOperationDeltaKeepsOperationsResolved exercises the echo
// received after a local root write has resolved accepted and rejected operations.
func TestDelayedOperationDeltaKeepsOperationsResolved(t *testing.T) {
	// Build a signed operation and a root that already acknowledges it.
	writer, err := peer.NewPeer(nil)
	if err != nil {
		t.Fatal(err)
	}
	priv, err := writer.GetPrivKey(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	op := buildTestSOOperation(t, priv, 1)
	root := buildTestSORoot(t, priv, 2, []*sobject.SOAccountNonce{{
		PeerId: writer.GetPeerID().String(),
		Nonce:  1,
	}})
	rejected := buildTestSOOperation(t, priv, 2)
	pending := buildTestSOOperation(t, priv, 3)
	state := &sobject.SOState{
		Root: root,
		OpRejections: []*sobject.SOPeerOpRejections{{
			PeerId: writer.GetPeerID().String(),
			Rejections: []*sobject.SOOperationRejection{
				buildTestSOOperationRejection(t, priv, writer.GetPeerID(), 2, rejected),
			},
		}},
	}
	rootData, err := (&api.PostRootRequest{Root: root}).MarshalVT()
	if err != nil {
		t.Fatal(err)
	}

	// Delayed accepted and rejected echoes stay terminal; a new operation and
	// its duplicate produce exactly one pending entry.
	for _, operation := range []*sobject.SOOperation{op, rejected, pending, pending} {
		data, err := operation.MarshalVT()
		if err != nil {
			t.Fatal(err)
		}
		if err := applyChangeLogEntry(testSharedObjectID, state, &api.SOStateDeltaEntry{
			ChangeType: "op",
			ChangeData: data,
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := applyChangeLogEntry(testSharedObjectID, state, &api.SOStateDeltaEntry{
		ChangeType: "root",
		ChangeData: rootData,
	}); err != nil {
		t.Fatal(err)
	}
	if len(state.GetOps()) != 1 {
		t.Fatalf("expected one unresolved operation, got %d", len(state.GetOps()))
	}
	if !state.GetOps()[0].EqualVT(pending) {
		t.Fatal("delayed echo replaced the unresolved operation")
	}
}
