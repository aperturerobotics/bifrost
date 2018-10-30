package node

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/bus"
)

// GetNodeWithPeerID executes a GetNodeSingleton directive.
// If peer ID is empty, selects any node.
func GetNodeWithPeerID(
	ctx context.Context,
	b bus.Bus,
	peerIDConstraint peer.ID,
) (Node, error) {
	v, ref, err := bus.ExecOneOff(ctx, b, NewGetNodeSingleton(peerIDConstraint), nil)
	if err != nil {
		return nil, err
	}
	ref.Release()

	return v.(Node), nil
}
