package link

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/bus"
)

// EstablishLinkWithPeerEx executes a EstablishLinkWithPeer directive.
// Returns a release function.
func EstablishLinkWithPeerEx(
	ctx context.Context,
	b bus.Bus,
	localPeerID, remotePeerID peer.ID,
	returnIfIdle bool,
) (Link, func(), error) {
	estl, _, ref, err := bus.ExecWaitValue[EstablishLinkWithPeerValue](
		ctx,
		b,
		NewEstablishLinkWithPeer(
			localPeerID, remotePeerID,
		),
		bus.ReturnIfIdle(returnIfIdle),
		nil,
		nil,
	)
	if err != nil {
		return nil, func() {}, err
	}

	return estl, ref.Release, nil
}
