package link

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/controllerbus/bus"
)

// OpenStreamViaLinkEx executes a OpenStreamViaLink directive.
// Returns a release function for the links used for the stream.
func OpenStreamViaLinkEx(
	ctx context.Context,
	b bus.Bus,
	remotePeerID peer.ID,
	protocolID protocol.ID,
	linkUUID uint64,
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
		NewOpenStreamViaLink(
			linkUUID,
			protocolID,
			openOpts,
			transportID,
		),
		nil,
	)
	if err != nil {
		estLinkInst.Release()
		return nil, func() {}, err
	}
	inst.Release()

	mstrm := val.GetValue().(MountedStream)
	return mstrm, estLinkInst.Release, nil
}
