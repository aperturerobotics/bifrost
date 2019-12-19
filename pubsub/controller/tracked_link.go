package pubsub_controller

import (
	"context"
	"errors"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/sirupsen/logrus"
)

type trackedLink struct {
	c         *Controller
	tpl       pubsub.PeerLinkTuple
	lnk       link.Link
	le        *logrus.Entry
	ctxCancel context.CancelFunc
}

// trackLink executes tracking the link.
func (t *trackedLink) trackLink(ctx context.Context) error {
	// hack: decide which side opens stream using whoever's peer id is greater
	if t.lnk.GetLocalPeer().Pretty() > t.lnk.GetRemotePeer().Pretty() {
		t.le.Debug("expecting peer to open stream")
		return nil
	}
	t.le.Debug("link tracking starting")
	ps, err := t.c.GetPubSub(ctx)
	if err != nil {
		return err
	}

	av, avRef, err := bus.ExecOneOff(
		ctx,
		t.c.bus,
		link.NewOpenStreamViaLink(
			t.lnk.GetUUID(),
			t.c.protocolID,
			stream.OpenOpts{Reliable: true, Encrypted: true},
			t.lnk.GetTransportUUID(),
		),
		nil,
	)
	if err != nil {
		return err
	}
	defer avRef.Release()

	mtStrm, ok := av.GetValue().(link.MountedStream)
	if !ok {
		return errors.New("open stream via link returned non-stream value")
	}

	t.le.WithField("protocol-id", mtStrm.GetProtocolID()).
		Info("pubsub stream opened (by us)")
	ps.AddPeerStream(t.tpl, true, mtStrm)
	return nil
}
