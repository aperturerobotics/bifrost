package flowgraph_plugin

import "github.com/aperturerobotics/controllerbus/config"

// ConfigID selects the Stream Echo plugin controller.
const ConfigID = ControllerID

// Validate accepts the plugin's empty startup configuration.
func (c *Config) Validate() error {
	return nil
}

// GetConfigID returns the plugin controller config identifier.
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig compares plugin startup configurations.
func (c *Config) EqualsConfig(other config.Config) bool {
	return config.EqualsConfig(c, other)
}

// _ is a type assertion.
var _ config.Config = (*Config)(nil)
