package s4wave_flowgraph

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	world_types "github.com/s4wave/spacewave/db/world/types"
	"github.com/s4wave/spacewave/net/peer"
	"github.com/sirupsen/logrus"
)

// Validate requires an independently addressed Flowgraph key.
func (o *CreateFlowgraphOp) Validate() error {
	_, err := ParseFlowgraphObjectKey(o.GetObjectKey())
	return err
}

// ApplyWorldOp creates the authored body and its type edge in the caller's transaction.
func (o *CreateFlowgraphOp) ApplyWorldOp(ctx context.Context, _ *logrus.Entry, ws world.WorldState, _ peer.ID) (bool, error) {
	// Require a valid identity before creating the object.
	if err := o.Validate(); err != nil {
		return false, err
	}

	// Store the initial authored body and release the returned object handle.
	obj, _, err := world.CreateWorldObject(ctx, ws, o.GetObjectKey(), func(cursor *block.Cursor) error {
		cursor.SetBlock(&Flowgraph{Name: o.GetName()}, true)
		return nil
	})
	defer world.ReleaseObjectState(obj)
	if err != nil {
		return false, err
	}

	// Register the object's type through the World graph.
	return false, world_types.SetObjectType(ctx, ws, o.GetObjectKey(), FlowgraphTypeID)
}

// _ is a type assertion.
var _ world.Operation = (*CreateFlowgraphOp)(nil)
