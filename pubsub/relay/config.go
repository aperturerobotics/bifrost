package pubsub_relay

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
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
	if len(c.GetTopicIds()) == 0 {
		return errors.New("at least one topic id required")
	}
	peerID, err := c.ParsePeerID()
	if err != nil {
		return err
	}
	if peerID == "" {
		return errors.New("peer id must be specified")
	}

	return nil
}

// ParsePeerID parses the peer ID if it is not empty.
func (c *Config) ParsePeerID() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetPeerId())
}
