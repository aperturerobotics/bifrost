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
		return nil, err
	}
	inst.Release()

	mstrm := val.GetValue().(MountedStream)
	return mstrm, nil
}
