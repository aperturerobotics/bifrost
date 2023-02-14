package stream_listening

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/aperturerobotics/controllerbus/config"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// ConfigID is the string used to identify this config object.
const ConfigID = ControllerID

// Validate validates the configuration.
// This is a cursory validation to see if the values "look correct."
func (c *Config) Validate() error {
	if c.GetLocalPeerId() != "" {
		if _, err := c.ParseLocalPeerID(); err != nil {
			return errors.Wrap(err, "local")
		}
	}
	if c.GetRemotePeerId() == "" {
		return errors.Wrap(peer.ErrEmptyPeerID, "remote")
	}
	if _, err := c.ParseRemotePeerID(); err != nil {
		return errors.Wrap(err, "remote")
	}
	if c.GetListenMultiaddr() == "" {
		return errors.New("listen multiaddr cannot be empty")
	}

	pid := protocol.ID(c.GetProtocolId())
	if err := pid.Validate(); err != nil {
		return err
	}

	if _, err := c.ParseListenMultiaddr(); err != nil {
		return err
	}

	return nil
}

// ParseLocalPeerID parses the local peer ID.
// may return nil.
func (c *Config) ParseLocalPeerID() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetLocalPeerId())
}

// ParseRemotePeerID parses the remote peer ID.
// may return nil.
func (c *Config) ParseRemotePeerID() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetRemotePeerId())
}

// ParseListenMultiaddr parses the multiaddress.
func (c *Config) ParseListenMultiaddr() (ma.Multiaddr, error) {
	return ma.NewMultiaddr(c.GetListenMultiaddr())
}

// GetConfigID returns the unique string for this configuration type.
// This string is stored with the encoded config.
func (c *Config) GetConfigID() string {
	return ConfigID
}

// EqualsConfig checks if the config is equal to another.
func (c *Config) EqualsConfig(other config.Config) bool {
	ot, ok := other.(*Config)
	if !ok {
		return false
	}

	return proto.Equal(ot, c)
}

var _ config.Config = ((*Config)(nil))
