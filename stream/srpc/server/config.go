package stream_srpc_server

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/sirupsen/logrus"
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
	return confparse.ParseProtocolIDs(c.GetProtocolIds(), false)
}

// GetDebugVals returns the directive arguments as key/value pairs.
// This should be something like param1="test", param2="test".
// This is not necessarily unique, and is primarily intended for display.
func (c *Config) GetDebugVals() config.DebugValues {
	return config.DebugValues{
		"peer-ids":     c.GetPeerIds(),
		"protocol-ids": c.GetProtocolIds(),
	}
}

// BuildServer constructs the server from the args.
func (c *Config) BuildServer(b bus.Bus, le *logrus.Entry, info *controller.Info, registerFns []RegisterFn) (*Server, error) {
	protocolIDs, err := c.ParseProtocolIDs()
	if err != nil {
		return nil, err
	}

	peerIDs, err := c.ParsePeerIDs()
	if err != nil {
		return nil, err
	}

	return NewServer(
		b,
		le,
		info,
		registerFns,
		protocolIDs,
		peer.IDsToString(peerIDs),
		c.GetDisableEstablishLink(),
	)
}
