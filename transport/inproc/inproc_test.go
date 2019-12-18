package inproc

import (
	"context"
	"testing"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/testbed"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/bifrost/transport/controller"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// buildTestbed builds a new testbed for udp.
func buildTestbed(t *testing.T, ctx context.Context) (*testbed.Testbed, *logrus.Entry) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	tb, err := testbed.NewTestbed(ctx, le, testbed.TestbedOpts{})
	if err != nil {
		t.Fatal(err.Error())
	}
	tb.StaticResolver.AddFactory(NewFactory(tb.Bus))
	return tb, le
}

func execPeer(ctx context.Context, t *testing.T, tb *testbed.Testbed, conf *Config) (
	*transport_controller.Controller,
	*Inproc,
	directive.Reference,
) {
	peerId, err := peer.IDFromPrivateKey(tb.PrivKey)
	if err != nil {
		t.Fatal(err.Error())
	}

	if conf == nil {
		conf = &Config{}
	}
	conf.TransportPeerId = peerId.Pretty()

	tp1, tp1Ref, err := bus.ExecOneOff(
		ctx,
		tb.Bus,
		resolver.NewLoadControllerWithConfig(conf),
		nil,
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	tpc1 := tp1.GetValue().(*transport_controller.Controller)
	tpt1, err := tpc1.GetTransport(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	return tpc1, tpt1.(*Inproc), tp1Ref
}

// TestEstablishLink tests creating a UDP link with two in-memory nodes.
func TestEstablishLink(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	tb1, le1 := buildTestbed(t, ctx)
	le1 = le1.WithField("testbed", 0)
	tb2, le2 := buildTestbed(t, ctx)
	le2 = le2.WithField("testbed", 1)

	tpc1, tp1, tp1Ref := execPeer(ctx, t, tb1, nil)
	peerId1 := tp1.GetPeerID()
	defer tp1Ref.Release()

	tpc2, tp2, tp2Ref := execPeer(ctx, t, tb2, &Config{
		Dialers: map[string]*dialer.DialerOpts{
			peerId1.Pretty(): &dialer.DialerOpts{
				Address: tp1.LocalAddr().String(),
			},
		},
	})
	peerId2 := tp2.GetPeerID()
	defer tp2Ref.Release()

	le1.Infof("constructed peer 1 with id %s", peerId1.Pretty())
	le2.Infof("constructed peer 2 with id %s", peerId2.Pretty())

	tp2.ConnectToInproc(ctx, tp1)
	tp1.ConnectToInproc(ctx, tp2)

	// Attempt to open a link between them.
	lnk2to1, lnk1Ref, err := bus.ExecOneOff(
		ctx,
		tb2.Bus,
		link.NewEstablishLinkWithPeer(peerId1),
		nil,
	)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer lnk1Ref.Release()

	le1.Infof(
		"opened link from 2 -> 1 with id %v",
		lnk2to1.GetValue().(link.Link).GetUUID(),
	)

	_ = tpc1
	_ = tpc2
}
