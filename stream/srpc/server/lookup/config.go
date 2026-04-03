package stream_srpc_server_lookup

import (
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/aperturerobotics/controllerbus/config"
)

// Validate checks the config.
func (c *Config) Validate() error {
	if _, err := confparse.ParsePeerIDs(c.GetPeerIds(), false); err != nil {
		return err
	}
	if _, err := confparse.ParseProtocolIDs(c.GetProtocolIds(), false); err != nil {
		return err
	}
	return nil
}

// GetDebugVals returns the directive arguments as key/value pairs.
func (c *Config) GetDebugVals() config.DebugValues {
	vals := config.DebugValues{
		"protocol-ids": c.GetProtocolIds(),
	}
	if pids := c.GetPeerIds(); len(pids) > 0 {
		vals["peer-ids"] = pids
	}
	if sid := c.GetServerId(); sid != "" {
		vals["server-id"] = []string{sid}
	}
	return vals
}
