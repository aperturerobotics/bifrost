package stream_drpc_client

import (
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/util/confparse"
)

// ParseSrcPeerId parses the source peer id, if set.
func (c *Config) ParseSrcPeerId() (peer.ID, error) {
	return confparse.ParsePeerID(c.GetSrcPeerId())
}

// ParseServerPeerIds parses the destination peer ids
func (c *Config) ParseServerPeerIds() ([]peer.ID, error) {
	serverPeerIDs := c.GetServerPeerIds()
	ids := make([]peer.ID, len(serverPeerIDs))
	for i, serverPeerID := range serverPeerIDs {
		var err error
		ids[i], err = confparse.ParsePeerID(serverPeerID)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
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
