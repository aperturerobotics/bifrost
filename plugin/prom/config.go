// Package prom_plugin stores Prometheus samples in a World, scrapes exporters
// into it from approved Devices, and answers PromQL over it.
//
// Each scrape target is one World object. Its root is a block-backed key-value
// tree holding the target's labels and scrape config, its series table, and
// one immutable segment per commit window. A Store appends samples to an
// in-memory window and commits the window as one segment, and a Queryable
// serves the fork's storage.Queryable interface from the committed segments of
// several targets. The spacewave-prom plugin registers the prom/target
// ObjectType, whose process scrapes the target on the Device its
// ProcessBinding approves.
package prom_plugin

import (
	"time"

	"github.com/aperturerobotics/controllerbus/config"
)

// ConfigID is the config identifier.
const ConfigID = ControllerID

// DefaultCommitWindow is the commit window of a config that sets none.
const DefaultCommitWindow = time.Minute

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
	return config.EqualsConfig(c, other)
}

// CommitWindow returns the time between a Device's commits in one Space.
func (c *Config) CommitWindow() time.Duration {
	if ms := c.GetCommitWindowMs(); ms != 0 {
		return time.Duration(ms) * time.Millisecond
	}
	return DefaultCommitWindow
}

// _ is a type assertion.
var _ config.Config = (*Config)(nil)
