package udp

import (
	"context"

	"github.com/aperturerobotics/bifrost/transport"
	tc "github.com/aperturerobotics/bifrost/transport/controller"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
)

// Factory constructs a UDP transport.
type Factory struct {
	// bus is the controller bus
	bus bus.Bus
}

// NewFactory builds a UDP transport factory.
func NewFactory(bus bus.Bus) *Factory {
	return &Factory{bus: bus}
}

// GetConfigID returns the configuration ID for the controller.
func (t *Factory) GetConfigID() string {
	return ConfigID
}

// GetControllerID returns the unique ID for the controller.
func (t *Factory) GetControllerID() string {
	return ControllerID
}

// ConstructConfig constructs an instance of the controller configuration.
func (t *Factory) ConstructConfig() config.Config {
	return &Config{}
}

// Construct constructs the associated controller given configuration.
// The transport's identity (private key) comes from a GetNode lookup.
func (t *Factory) Construct(
	conf config.Config,
	opts controller.ConstructOpts,
) (controller.Controller, error) {
	le := opts.GetLogger()
	cc := conf.(*Config)

	peerIDConstraint, err := cc.ParseNodePeerID()
	if err != nil {
		return nil, err
	}

	// Construct the transport controller.
	return tc.NewController(
		le,
		t.bus,
		peerIDConstraint,
		func(
			ctx context.Context,
			le *logrus.Entry,
			pkey crypto.PrivKey,
			handler transport.TransportHandler,
		) (transport.Transport, error) {
			return NewUDP(
				ctx,
				le,
				pkey,
				handler,
				cc.GetPacketOpts(),
				0,
				cc.GetListenAddr(),
			)
		},
		TransportID,
		Version,
		cc.GetDialers(),
	), nil
}

// GetVersion returns the version of this controller.
func (t *Factory) GetVersion() semver.Version {
	return Version
}

// _ is a type assertion
var _ controller.Factory = ((*Factory)(nil))
