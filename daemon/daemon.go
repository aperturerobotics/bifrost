package daemon

import (
	"context"

	"github.com/aperturerobotics/bifrost/core"
	api_controller "github.com/aperturerobotics/bifrost/daemon/api/controller"
	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/bifrost/peer"
	nctr "github.com/aperturerobotics/bifrost/peer/controller"
	nats_controller "github.com/aperturerobotics/bifrost/pubsub/nats/controller"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
)

// Daemon implements the Bifrost daemon.
type Daemon struct {
	// bus is the controller bus.
	bus bus.Bus
	// staticResolver is the static controller factory resolver.
	staticResolver *static.Resolver

	// nodePriv is the primary node private key
	nodePriv crypto.PrivKey
	// nodePeerID is the primary node ID
	nodePeerID peer.ID
	// nodePeerIDPretty is the node peer ID as a b58 address
	nodePeerIDPretty string

	// closeCbs are funcs to call when we close the daemon
	closeCbs []func()
}

// ConstructOpts are extra options passed to the daemon constructor.
type ConstructOpts struct {
	// LogEntry is the root logger to use.
	// If unset, will use a default logger.
	LogEntry *logrus.Entry
	// ExtraControllerFactories is a set of extra controller factories to
	// make available to the daemon.
	ExtraControllerFactories []func(bus.Bus) controller.Factory
}

// NewDaemon constructs a new daemon.
func NewDaemon(
	ctx context.Context,
	nodePriv crypto.PrivKey,
	opts ConstructOpts,
) (*Daemon, error) {
	le := opts.LogEntry
	if le == nil {
		log := logrus.New()
		log.SetLevel(logrus.DebugLevel)
		le = logrus.NewEntry(log)
	}

	// Construct the controller bus.
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return nil, err
	}

	sr.AddFactory(api_controller.NewFactory(b))
	sr.AddFactory(nats_controller.NewFactory(b))
	for _, ctor := range opts.ExtraControllerFactories {
		if ctor != nil {
			sr.AddFactory(ctor(b))
		}
	}

	// Construct the node controller.
	peerID, err := peer.IDFromPrivateKey(nodePriv)
	if err != nil {
		return nil, err
	}

	peerIDPretty := peerID.Pretty()
	nodePrivKeyPem, err := keypem.MarshalPrivKeyPem(nodePriv)
	if err != nil {
		return nil, err
	}

	dir := resolver.NewLoadControllerWithConfig(&nctr.Config{
		PrivKey: string(nodePrivKeyPem),
	})
	_, _, ncRef, err := loader.WaitExecControllerRunning(ctx, b, dir, nil)
	if err != nil {
		return nil, err
	}
	le.Debugf("node controller resolved w/ ID: %s", peerIDPretty)

	return &Daemon{
		bus: b,

		closeCbs: []func(){
			ncRef.Release,
		},
		nodePriv:         nodePriv,
		nodePeerID:       peerID,
		staticResolver:   sr,
		nodePeerIDPretty: peerIDPretty,
	}, nil
}

// GetStaticResolver returns the underlying static resolver for controller impl lookups.
func (d *Daemon) GetStaticResolver() *static.Resolver {
	return d.staticResolver
}

// GetControllerBus returns the controller bus.
func (d *Daemon) GetControllerBus() bus.Bus {
	return d.bus
}
