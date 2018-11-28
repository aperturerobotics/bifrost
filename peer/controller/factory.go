package peer_controller

import (
	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
)

// Factory constructs a Peer controller.
type Factory struct {
	// bus is the controller bus
	bus bus.Bus
}

// NewFactory builds a websocket transport factory.
func NewFactory(bus bus.Bus) *Factory {
	return &Factory{bus: bus}
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
func (t *Factory) Construct(
	conf config.Config,
	opts controller.ConstructOpts,
) (controller.Controller, error) {
	le := opts.GetLogger()
	cc := conf.(*Config)

	privKey, err := cc.ParsePrivateKey()
	if err != nil {
		return nil, err
	}

	if privKey == nil {
		le.Info("generating private key, none configured")
		privKey, _, err = keypem.GeneratePrivKey()
		if err != nil {
			return nil, err
		}
	}

	return NewController(le, privKey)
}

// GetVersion returns the version of this controller.
func (t *Factory) GetVersion() semver.Version {
	return Version
}

// _ is a type assertion
var _ controller.Factory = ((*Factory)(nil))
