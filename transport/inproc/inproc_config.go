package inproc

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/aperturerobotics/controllerbus/config"
	"google.golang.org/protobuf/proto"
)

// ConfigID is the string used to identify this config object.
const ConfigID = ControllerID

// Validate validates the configuration.
// This is a cursory validation to see if the values "look correct."
func (c *Config) Validate() error { return nil }

// ParseTransportPeerID parses the node peer ID if it is not empty.
func (c *Config) ParseTransportPeerID() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetTransportPeerId())
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
	return vals
}

// _ is a type assertion
var _ transport.Config = ((*Config)(nil))
