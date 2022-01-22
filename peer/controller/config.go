package peer_controller

import (
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// ConfigID is the identifier for the config type.
const ConfigID = ControllerID

// NewConfigWithPrivKey builds a new configuration with a private key
func NewConfigWithPrivKey(pk crypto.PrivKey) (*Config, error) {
	privKeyStr, err := confparse.MarshalPrivateKey(pk)
	if err != nil {
		return nil, err
	}

	return &Config{
		PrivKey: privKeyStr,
	}, nil
}

// GetConfigID returns the config identifier.
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks equality between two configs.
func (c *Config) EqualsConfig(c2 config.Config) bool {
	oc, ok := c2.(*Config)
	if !ok {
		return false
	}

	if c.GetPrivKey() == oc.GetPrivKey() {
		return true
	}

	pk1, err := c.ParsePrivateKey()
	if err != nil {
		return false
	}

	pk2, err := oc.ParsePrivateKey()
	if err != nil {
		return false
	}

	return pk2.Equals(pk1)
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if _, err := c.ParsePrivateKey(); err != nil {
		return err
	}

	return nil
}

// ParsePrivateKey parses the private key from the configuration.
func (c *Config) ParsePrivateKey() (crypto.PrivKey, error) {
	return confparse.ParsePrivateKey(c.GetPrivKey())
}
