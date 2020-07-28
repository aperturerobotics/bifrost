package nats_controller

import (
	"context"
	"errors"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/aperturerobotics/bifrost/pubsub/controller"
	"github.com/aperturerobotics/bifrost/pubsub/nats"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// Factory constructs a GossipSub controller.
type Factory struct {
	// bus is the controller bus
	bus bus.Bus
}

// NewFactory builds a GossipSub controller factory.
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

	pid, err := cc.ParsePeerID()
	if err != nil {
		return nil, err
	}

	if len(pid) == 0 {
		return nil, errors.New("nats requires a peer id to be set")
	}

	// Construct the EntityGraph controller.
	return pubsub_controller.NewController(
		le,
		t.bus,
		controller.NewInfo(ControllerID, Version, "nats controller"),
		pid,
		nats.NatsRouterID,
		func(
			ctx context.Context,
			le *logrus.Entry,
			peer peer.Peer,
			handler pubsub.PubSubHandler,
		) (pubsub.PubSub, error) {
			return nats.NewNats(
				ctx,
				le,
				handler,
				cc.GetNatsConfig(),
				peer,
			)
		},
	), nil
}

// GetVersion returns the version of this controller.
func (t *Factory) GetVersion() semver.Version {
	return Version
}

// _ is a type assertion
var _ controller.Factory = ((*Factory)(nil))
