package link_establish_controller

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

// ParsePeerIDs parses the peer ids field.
func (c *Config) ParsePeerIDs() ([]peer.ID, error) {
	return confparse.ParsePeerIDs(c.GetPeerIds(), false)
}

// SetPeerIDs sets the peer ids field.
func (c *Config) SetPeerIDs(ids []peer.ID) {
	pids := make([]string, len(ids))
	for i, pid := range ids {
		pids[i] = peer.IDB58Encode(pid)
	}
	c.PeerIds = pids
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if len(c.GetPeerIds()) == 0 {
		return errors.New("at least one peer id required")
	}
	if _, err := c.ParsePeerIDs(); err != nil {
		return errors.Wrap(err, "peer_ids")
	}
	if _, err := c.ParseSrcPeerId(); err != nil {
		return errors.Wrap(err, "src_peer_id")
	}
	return nil
}

// ParseSrcPeerId parses the source peer id.
// May return empty.
func (c *Config) ParseSrcPeerId() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetSrcPeerId())
}
