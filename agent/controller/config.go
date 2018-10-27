package agent_controller

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/util/confparse"
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
	return confparse.ParsePrivateKey(privKeyDat)
}

// ParseNodePeerID parses the node peer ID if it is not empty.
func (c *Config) ParseNodePeerID() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetNodePeerId())
}
