package stream_forwarding

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
)

// ControllerID identifies the API controller.
const ControllerID = "bifrost/stream/forwarding/1"

// Version is the controller version.
var Version = semver.MustParse("0.0.1")

// Factory constructs a forwarding controller
type Factory struct{}

// NewFactory builds a UDP transport factory.
func NewFactory() *Factory {
	return &Factory{}
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

	// Construct the controller.
	return NewController(le, cc)
}

// GetVersion returns the version of this controller.
func (t *Factory) GetVersion() semver.Version {
	return Version
}

// _ is a type assertion
var _ controller.Factory = ((*Factory)(nil))
