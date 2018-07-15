package libp2p

import (
	"errors"
	"github.com/aperturerobotics/controllerbus/config"
	ma "github.com/multiformats/go-multiaddr"
)

// TransportConfig is an identified transport configuration.
type TransportConfig struct {
	TransportConfigInner

	// controllerID is the id of the transport controller
	// this is used to differentiate config objects
	controllerID string

	// configID is the computed config id
	configID string
}

// NewTransportConfig constructs a new transport config.
func NewTransportConfig(controllerID string, tci TransportConfigInner) *TransportConfig {
	configID := "/libp2p/transport/" + controllerID

	return &TransportConfig{
		controllerID:         controllerID,
		configID:             configID,
		TransportConfigInner: tci,
	}
}

// ParseListenMultiaddr parses the listen multiaddress field.
func (c *TransportConfigInner) ParseListenMultiaddr() (ma.Multiaddr, error) {
	lm := c.GetListenMultiaddr()
	if lm == "" {
		return nil, errors.New("listen multiaddr cannot be empty")
	}

	return ma.NewMultiaddr(lm)
}

// Validate validates the configuration.
// This is a cursory validation to see if the values "look correct."
func (c *TransportConfig) Validate() error {
	if _, err := c.ParseListenMultiaddr(); err != nil {
		return err
	}

	return nil
}

// GetConfigID returns the unique string for this configuration type.
// This string is stored with the encoded config.
// Example: bifrost/transport/udp/1
func (c *TransportConfig) GetConfigID() string {
	return c.configID
}

// _ is a type assertion
var _ config.Config = ((*TransportConfig)(nil))
