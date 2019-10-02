package link_establish_controller

import (
	"github.com/aperturerobotics/bifrost/peer"
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
	sl := c.GetPeerIds()
	pids := make([]peer.ID, len(sl))
	var err error
	for i, pidStr := range sl {
		pids[i], err = peer.IDB58Decode(pidStr)
		if err != nil {
			return nil, errors.Wrapf(err, "peer_ids[%d]", i)
		}
	}
	return pids, nil
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
	return nil
}
