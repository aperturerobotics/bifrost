package link

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/controllerbus/bus"
)

// OpenStreamWithPeerEx executes a OpenStreamWithPeer directive.
// Returns a release function for the links used for the stream.
func OpenStreamWithPeerEx(
	ctx context.Context,
	b bus.Bus,
	protocolID protocol.ID,
	localPeerID, remotePeerID peer.ID,
	transportID uint64,
	openOpts stream.OpenOpts,
) (MountedStream, func(), error) {
	_, estLinkRef, err := b.AddDirective(
		NewEstablishLinkWithPeer(localPeerID, remotePeerID),
		nil,
	)
	if err != nil {
		return nil, func() {}, err
	}

	mstrm, _, ref, err := bus.ExecWaitValue[MountedStream](
		ctx,
		b,
		NewOpenStreamWithPeer(
			protocolID,
			localPeerID, remotePeerID,
			transportID,
			openOpts,
		),
		nil,
		nil,
		nil,
	)
	if err != nil {
		estLinkRef.Release()
		return nil, func() {}, err
	}
	ref.Release()

	return mstrm, estLinkRef.Release, nil
}
