package inproc

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/golang/protobuf/proto"
)

// ConfigID is the string used to identify this config object.
const ConfigID = ControllerID

// Validate validates the configuration.
// This is a cursory validation to see if the values "look correct."
func (c *Config) Validate() error { return nil }

// ParseNodePeerID parses the node peer ID if it is not empty.
func (c *Config) ParseNodePeerID() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetNodePeerId())
}

// GetConfigID returns the unique string for this configuration type.
// This string is stored with the encoded config.
// Example: bifrost/transport/xbee/1
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks if the other config is equal.
func (c *Config) EqualsConfig(other config.Config) bool {
	return proto.Equal(c, other)
}

// SetNodePeerId sets the node peer ID field.
func (c *Config) SetNodePeerId(peerID string) {
	c.NodePeerId = peerID
}

// _ is a type assertion
var _ transport.Config = ((*Config)(nil))
