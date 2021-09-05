package websocket

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
	return confparse.ParsePeerID(c.GetTransportPeerId())
}

// ParseRestrictPeerID parses the remote peer ID restriction if it is not empty.
func (c *Config) ParseRestrictPeerID() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetRestrictPeerId())
}

// GetConfigID returns the unique string for this configuration type.
// This string is stored with the encoded config.
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks if the other config is equal.
func (c *Config) EqualsConfig(other config.Config) bool {
	return proto.Equal(c, other)
}

// SetTransportPeerId sets the node peer ID field.
func (c *Config) SetTransportPeerId(peerID string) {
	c.TransportPeerId = peerID
}

// GetDebugVals returns the directive arguments as key/value pairs.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (c *Config) GetDebugVals() config.DebugValues {
	vals := make(config.DebugValues)
	if tp := c.GetTransportPeerId(); tp != "" {
		vals["peer-id"] = []string{tp}
	}
	if la := c.GetListenAddr(); la != "" {
		vals["listen-addr"] = []string{la}
	}
	return vals
}

// _ is a type assertion
var _ transport.Config = ((*Config)(nil))

// _ is a type assertion
var _ config.Debuggable = ((*Config)(nil))
