package dex_solicit

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
)

// Factory constructs a solicitation-based DEX controller.
type Factory struct {
	// bus resolves the local buckets and peer streams.
	bus bus.Bus
	// authorizeRequest supplies the caller's incoming request policy.
	authorizeRequest RequestAuthorizer
}

// NewFactory builds a solicitation-based DEX controller factory.
func NewFactory(bus bus.Bus) *Factory {
	return NewFactoryWithRequestAuthorizer(bus, nil)
}

// NewFactoryWithRequestAuthorizer builds a factory with an incoming request policy.
func NewFactoryWithRequestAuthorizer(bus bus.Bus, authorizeRequest RequestAuthorizer) *Factory {
	// Retain the policy for every exchange constructed by this factory.
	return &Factory{bus: bus, authorizeRequest: authorizeRequest}
}

// GetConfigID returns the unique ID for the config.
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
func (t *Factory) Construct(
	ctx context.Context,
	conf config.Config,
	opts controller.ConstructOpts,
) (controller.Controller, error) {
	le := opts.GetLogger()
	cc := conf.(*Config)
	return NewControllerWithRequestAuthorizer(le, t.bus, cc, t.authorizeRequest)
}

// GetVersion returns the version of this controller.
func (t *Factory) GetVersion() controller.Version {
	return Version
}

// _ is a type assertion
var _ controller.Factory = (*Factory)(nil)
