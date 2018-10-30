//+build !js

package daemon

import (
	"context"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/keypem"
	nctr "github.com/aperturerobotics/bifrost/node/controller"
	"github.com/aperturerobotics/bifrost/peer"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	"github.com/aperturerobotics/controllerbus/core"
	"github.com/aperturerobotics/objstore/db"
	"github.com/aperturerobotics/objstore/db/inmem"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// Daemon implements the Bifrost daemon.
type Daemon struct {
	// bus is the controller bus.
	bus bus.Bus
	// staticResolver is the static controller factory resolver.
	staticResolver *static.Resolver
	// db is the database
	db db.Db

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
	// Database is the key-value storage database to use.
	// If nil, will use an in-memory database.
	Database db.Db
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
	controllerBus, staticResolver, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return nil, err
	}

	staticResolver.AddFactory(api.NewFactory(controllerBus))
	staticResolver.AddFactory(wtpt.NewFactory(controllerBus))
	staticResolver.AddFactory(udptpt.NewFactory(controllerBus))
	staticResolver.AddFactory(nctr.NewFactory(controllerBus))

	for _, factory := range opts.ExtraControllerFactories {
		if con := factory(controllerBus); con != nil {
			staticResolver.AddFactory(con)
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

	dir := resolver.NewLoadControllerWithConfigSingleton(&nctr.Config{
		PrivKey: string(nodePrivKeyPem),
	})
	val, valRef, err := bus.ExecOneOff(ctx, controllerBus, dir, nil)
	if err != nil {
		return nil, err
	}
	_ = val
	le.Infof("node controller resolved w/ ID: %s", peerIDPretty)

	ddb := opts.Database
	if ddb == nil {
		ddb = inmem.NewInmemDb()
	}

	return &Daemon{
		bus: controllerBus,
		db:  ddb,

		closeCbs:         []func(){valRef.Release},
		nodePriv:         nodePriv,
		nodePeerID:       peerID,
		staticResolver:   staticResolver,
		nodePeerIDPretty: peerIDPretty,
	}, nil
}

// GetControllerBus returns the controller bus.
func (d *Daemon) GetControllerBus() bus.Bus {
	return d.bus
}
