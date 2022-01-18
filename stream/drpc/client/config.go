package stream_drpc_client

import (
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/pkg/errors"
)

// Validate validates the config.
func (c *Config) Validate() error {
	if err := c.GetDrpcOpts().Validate(); err != nil {
		return err
	}
	if _, err := c.ParseSrcPeerId(); err != nil {
		return errors.Wrap(err, "src_peer_id")
	}
	if _, err := c.ParseServerPeerIds(); err != nil {
		return errors.Wrap(err, "server_peer_ids")
	}
	if _, err := c.ParseTimeoutDur(); err != nil {
		return errors.Wrap(err, "timeout_dur")
	}
	return nil
}

// ParseSrcPeerId parses the source peer id, if set.
func (c *Config) ParseSrcPeerId() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetSrcPeerId())
}

// ParseServerPeerIds parses the destination peer ids
func (c *Config) ParseServerPeerIds() ([]peer.ID, error) {
	return confparse.ParsePeerIDs(c.GetServerPeerIds(), false)
}

// ParseTimeoutDur parses the timeout duration.
// returns zero if empty
func (c *Config) ParseTimeoutDur() (time.Duration, error) {
	durStr := c.GetTimeoutDur()
	if durStr == "" {
		return 0, nil
	}
	return time.ParseDuration(durStr)
}
