package peer

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
)

// GetPeerWithID gets a peer.
// If peer ID is empty, selects any peer.
// returnIfIdle if set, will return if the directive becomes idle.
// valDisposeCallback is called when the value is no longer valid.
// valDisposeCallback can be nil.
func GetPeerWithID(
	ctx context.Context,
	b bus.Bus,
	peerIDConstraint ID,
	returnIfIdle bool,
	valDisposeCallback func(),
) (Peer, directive.Instance, directive.Reference, error) {
	return bus.ExecWaitValue[Peer](
		ctx,
		b,
		NewGetPeer(peerIDConstraint),
		returnIfIdle,
		valDisposeCallback,
		nil,
	)
}
