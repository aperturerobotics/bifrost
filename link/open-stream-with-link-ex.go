package link

import (
	"context"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/controllerbus/bus"
)

// OpenStreamViaLinkEx executes a OpenStreamViaLink directive.
func OpenStreamViaLinkEx(
	ctx context.Context,
	b bus.Bus,
	remotePeerID peer.ID,
	protocolID protocol.ID,
	linkUUID uint64,
	transportID uint64,
	openOpts stream.OpenOpts,
) (MountedStream, error) {
	mstrm, _, ref, err := bus.ExecWaitValue[MountedStream](
		ctx,
		b,
		NewOpenStreamViaLink(
			linkUUID,
			protocolID,
			openOpts,
			transportID,
		),
		false,
		nil,
		nil,
	)
	if ref != nil {
		ref.Release()
	}
	return mstrm, err
}
