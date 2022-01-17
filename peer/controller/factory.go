package peer_controller

import (
	"errors"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/blang/semver"
)

// Factory constructs a Peer controller.
type Factory struct {
}

// NewFactory builds a peer controller factory.
func NewFactory() *Factory {
	return &Factory{}
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
		return nil, errors.New("private key must be configured")
		/*
			le.Info("generating private key, none configured")
			privKey, _, err = keypem.GeneratePrivKey()
			if err != nil {
				return nil, err
			}
		*/
	}

	return NewController(le, privKey)
}

// GetVersion returns the version of this controller.
func (t *Factory) GetVersion() semver.Version {
	return Version
}

// _ is a type assertion
var _ controller.Factory = ((*Factory)(nil))
