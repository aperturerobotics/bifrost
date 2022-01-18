package stream_drpc_server

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/util/confparse"
)

// Validate checks the config.
func (c *Config) Validate() error {
	if _, err := c.ParsePeerIDs(); err != nil {
		return err
	}
	if err := c.GetDrpcOpts().Validate(); err != nil {
		return err
	}
	// TODO
	return nil
}

// ParsePeerIDs parses the peer ids field.
func (c *Config) ParsePeerIDs() ([]peer.ID, error) {
	return confparse.ParsePeerIDs(c.GetPeerIds(), false)
}
