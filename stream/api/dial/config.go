package stream_api_dial

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/util/confparse"
)

// Validate validates the configuration.
// This is a cursory validation to see if the values "look correct."
func (c *Config) Validate() error {
	if c.GetPeerId() == "" {
		return peer.ErrPeerIDEmpty
	}

	if _, err := c.ParseLocalPeerID(); err != nil {
		return err
	}

	pid := protocol.ID(c.GetProtocolId())
	if err := pid.Validate(); err != nil {
		return err
	}

	return nil
}

// ParseLocalPeerID parses the local peer ID constraint.
// may be empty.
func (c *Config) ParseLocalPeerID() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetLocalPeerId())
}

// ParsePeerID parses the target peer ID constraint.
func (c *Config) ParsePeerID() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetPeerId())
}
