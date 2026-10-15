//go:build !goscript

package plugin_host_wazero_quickjs

import (
	"path/filepath"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/pkg/errors"
)

// ConfigID is the config identifier.
const ConfigID = ControllerID

// NewConfig constructs a new controller config.
// Sets the most important fields only.
func NewConfig() *Config {
	return &Config{}
}

// GetConfigID returns the unique string for this configuration type.
// This string is stored with the encoded config.
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks if the config is equal to another.
func (c *Config) EqualsConfig(other config.Config) bool {
	return config.EqualsConfig(c, other)
}

// Validate checks that a set cache directory is an absolute path.
func (c *Config) Validate() error {
	if c.GetCacheDir() != "" && !filepath.IsAbs(c.GetCacheDir()) {
		return errors.New("cache_dir must be an absolute path")
	}
	return nil
}

// _ is a type assertion
var _ config.Config = (*Config)(nil)
