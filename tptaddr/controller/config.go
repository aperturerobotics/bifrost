package tptaddr_controller

import (
	"github.com/aperturerobotics/controllerbus/config"
)

// ConfigID is the string used to identify this config object.
const ConfigID = ControllerID

// Validate validates the configuration.
func (c *Config) Validate() error {
	return nil
}

// GetConfigID returns the unique string for this configuration type.
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks if the config is equal to another.
func (c *Config) EqualsConfig(other config.Config) bool {
	ot, ok := other.(*Config)
	if !ok {
		return false
	}

	return ot.EqualVT(c)
}

var _ config.Config = ((*Config)(nil))
