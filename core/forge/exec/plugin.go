package space_exec

import (
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/pkg/errors"
)

// PluginExecConfigID is the config ID for the plugin bridge handler.
const PluginExecConfigID = "space-exec/plugin"

// GetConfigID returns the config ID.
func (c *PluginExecConfig) GetConfigID() string {
	return PluginExecConfigID
}

// Validate checks that the plugin bridge has enough routing information.
func (c *PluginExecConfig) Validate() error {
	if c.GetPluginId() == "" {
		return errors.New("plugin_id is required")
	}
	if c.GetControllerId() == "" {
		return errors.New("controller_id is required")
	}
	return nil
}

// EqualsConfig checks equality with another plugin bridge config.
func (c *PluginExecConfig) EqualsConfig(other config.Config) bool {
	oc, ok := other.(*PluginExecConfig)
	if !ok {
		return false
	}
	return c.EqualVT(oc)
}

// _ is a type assertion
var _ config.Config = (*PluginExecConfig)(nil)
