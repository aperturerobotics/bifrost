package simulate

import (
	"context"

	bp "github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/sim/graph"
	"github.com/aperturerobotics/bifrost/testbed"
	transport_controller "github.com/aperturerobotics/bifrost/transport/controller"
	"github.com/aperturerobotics/bifrost/transport/inproc"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

// Peer holds state about an executing Peer
type Peer struct {
	// graphPeer is the underlying graph peer.
	graphPeer *graph.Peer
	// le is the root log entry
	le *logrus.Entry
	// ctx is the context for the peer
	ctx context.Context
	// rel releases all of this peer's resources
	rel func()
	// testbed is the peer's testbed.
	testbed *testbed.Testbed
	// inproc is the in-process transport instance.
	inproc *inproc.Inproc
	// transportController is the transport controller
	transportController *transport_controller.Controller
}

// newPeer constructs a new Peer.
func newPeer(ctx context.Context, le *logrus.Entry, gp *graph.Peer) (*Peer, error) {
	var rels []func()
	rel := func() {
		for _, r := range rels {
			r()
		}
	}
	np := &Peer{
		graphPeer: gp,
		rel:       rel,
		le:        le.WithField("sim-peer", gp.ID()),
		// WithField("sim-peer", gp.GetPeerID().Pretty()).
	}
	var ctxCancel func()
	np.ctx, ctxCancel = context.WithCancel(ctx)
	rels = append(rels, ctxCancel)
	var err error
	np.testbed, err = testbed.NewTestbed(
		np.ctx,
		np.le, // np.le,
		testbed.TestbedOpts{PrivKey: gp.GetPeerPriv()},
	)
	if err != nil {
		rel()
		return nil, err
	}
	np.testbed.StaticResolver.AddFactory(inproc.NewFactory(np.testbed.Bus))
	for _, extraFactoryCtor := range gp.GetExtraFactories() {
		np.testbed.StaticResolver.AddFactory(extraFactoryCtor(np.testbed.Bus))
	}
	for _, extraFactoryCtor := range gp.GetExtraFactoryAdders() {
		extraFactoryCtor(np.testbed.Bus, np.testbed.StaticResolver)
	}
	tp1, _, tp1Ref, err := loader.WaitExecControllerRunning(
		np.ctx,
		np.testbed.Bus,
		resolver.NewLoadControllerWithConfig(&inproc.Config{
			TransportPeerId: np.GetPeerID().Pretty(),
		}),
		nil,
	)
	if err != nil {
		rel()
		return nil, err
	}
	rels = append(rels, tp1Ref.Release)
	tpc := tp1.(*transport_controller.Controller)
	tp, err := tpc.GetTransport(np.ctx)
	if err != nil {
		rel()
		return nil, err
	}
	np.inproc = tp.(*inproc.Inproc)
	np.transportController = tpc
	_, _, tp2Ref, err := bus.ExecOneOff(
		np.ctx,
		np.testbed.Bus,
		resolver.NewLoadControllerWithConfig(&configset_controller.Config{}),
		nil,
		nil,
	)
	if err != nil {
		rel()
		return nil, err
	}
	rels = append(rels, tp2Ref.Release)
	_, tp3Ref, err := np.testbed.Bus.AddDirective(
		configset.NewApplyConfigSet(gp.GetConfigSet()),
		nil,
	)
	if err != nil {
		rel()
		return nil, err
	}
	rels = append(rels, tp3Ref.Release)
	return np, nil
}

// GetPeerID returns the Peer ID.
func (p *Peer) GetPeerID() bp.ID {
	return p.graphPeer.GetPeerID()
}

// GetPeerPriv returns the Peer private key.
func (p *Peer) GetPeerPriv() crypto.PrivKey {
	return p.graphPeer.GetPeerPriv()
}

// GetTestbed returns the testbed.
func (p *Peer) GetTestbed() *testbed.Testbed {
	return p.testbed
}

// GetInproc returns the in-proc transport.
func (p *Peer) GetInproc() *inproc.Inproc {
	return p.inproc
}

// GetTransportController returns the transport controller.
func (p *Peer) GetTransportController() *transport_controller.Controller {
	return p.transportController
}

// Close cancels the Peer's subroutines.
func (p *Peer) Close() {
	p.rel()
}
