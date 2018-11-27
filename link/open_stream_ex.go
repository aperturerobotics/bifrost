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
	_, estLinkInst, err := b.AddDirective(
		NewEstablishLinkWithPeer(remotePeerID),
		nil,
	)
	if err != nil {
		return nil, func() {}, err
	}

	val, inst, err := bus.ExecOneOff(
		ctx,
		b,
		NewOpenStreamWithPeer(
			protocolID,
			localPeerID, remotePeerID,
			transportID,
			openOpts,
		),
		nil,
	)
	if err != nil {
		estLinkInst.Release()
		return nil, func() {}, err
	}
	inst.Release()

	mstrm := val.(MountedStream)
	return mstrm, inst.Release, nil
}
