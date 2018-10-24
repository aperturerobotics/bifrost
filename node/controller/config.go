package node_controller

import (
	"errors"

	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-crypto"
)

// ConfigID is the identifier for the config type.
const ConfigID = ControllerID

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

	return proto.Equal(c, oc)
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
	privKeyDat := []byte(c.GetPrivKey())
	if len(privKeyDat) == 0 {
		return nil, nil
	}

	key, err := keypem.ParsePrivKeyPem(privKeyDat)
	if err != nil {
		return nil, err
	}

	if key == nil {
		return nil, errors.New("no pem data found")
	}

	return key, nil
}
