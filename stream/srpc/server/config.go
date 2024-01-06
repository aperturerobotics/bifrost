package stream_srpc_server

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/util/confparse"
)

// Validate checks the config.
func (c *Config) Validate() error {
	if _, err := c.ParsePeerIDs(); err != nil {
		return err
	}
	if _, err := c.ParseProtocolIDs(); err != nil {
		return err
	}
	return nil
}

// ParsePeerIDs parses the peer ids field.
func (c *Config) ParsePeerIDs() ([]peer.ID, error) {
	return confparse.ParsePeerIDs(c.GetPeerIds(), false)
}

// ParseProtocolIDs parses the protocol ids field.
func (c *Config) ParseProtocolIDs() ([]protocol.ID, error) {
	return confparse.ParseProtocolIDs(c.GetPeerIds(), false)
}
