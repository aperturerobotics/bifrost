package space_world_ops

import (
	"context"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/net/peer"
	s4wave_canvas "github.com/s4wave/spacewave/sdk/canvas"
	"github.com/sirupsen/logrus"
)

// NewCanvasAddNodeOp constructs a new CanvasAddNodeOp block.
func NewCanvasAddNodeOp(objKey string, node *s4wave_canvas.CanvasNode) *CanvasAddNodeOp {
	return &CanvasAddNodeOp{
		ObjectKey: objKey,
		Node:      node,
	}
}

// Validate performs cursory checks on the op.
func (o *CanvasAddNodeOp) Validate() error {
	// Require a target object and a canvas node with an ID.
	if len(o.GetObjectKey()) == 0 {
		return world.ErrEmptyObjectKey
	}
	node := o.GetNode()
	if node == nil {
		return ErrNodeRequired
	}
	if len(node.GetId()) == 0 {
		return ErrNodeIdRequired
	}
	return nil
}

// ApplyWorldOp applies the operation as a world operation.
func (o *CanvasAddNodeOp) ApplyWorldOp(
	ctx context.Context,
	le *logrus.Entry,
	worldHandle world.WorldState,
	sender peer.ID,
) (sysErr bool, err error) {
	// Validate the canvas node addition before accessing its object.
	if err := o.Validate(); err != nil {
		return false, err
	}

	// Select the canvas object and node to add.
	objKey := o.GetObjectKey()
	node := o.GetNode()

	// Update the canvas object with the requested node.
	_, _, err = world.AccessWorldObject(ctx, worldHandle, objKey, true, func(bcs *block.Cursor) error {
		// Load the canvas state or prepare an empty state.
		state, uerr := s4wave_canvas.UnmarshalCanvasState(ctx, bcs)
		if uerr != nil {
			return uerr
		}
		if state == nil {
			state = &s4wave_canvas.CanvasState{}
		}

		// Preserve the previous canvas state and persist the new node.
		previous := state.CloneVT()
		if state.Nodes == nil {
			state.Nodes = make(map[string]*s4wave_canvas.CanvasNode)
		}
		state.Nodes[node.GetId()] = node
		return s4wave_canvas.WriteCanvasState(ctx, bcs, previous, state)
	})
	if err != nil {
		return false, err
	}

	return false, nil
}

// _ is a type assertion
var _ world.Operation = (*CanvasAddNodeOp)(nil)
