//+build !js

package api

import (
	"github.com/aperturerobotics/bifrost/peer"
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
	peerID := c.GetNodePeerId()
	if peerID == "" {
		return "", nil
	}

	return peer.IDFromString(peerID)
}

// GetConfigID returns the unique string for this configuration type.
// This string is stored with the encoded config.
// Example: bifrost/daemon/api/1
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks if the other config is equal.
func (c *Config) EqualsConfig(other config.Config) bool {
	return proto.Equal(c, other)
}

// _ is a type assertion
var _ config.Config = ((*Config)(nil))
