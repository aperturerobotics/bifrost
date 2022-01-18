package confparse

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/pkg/errors"
)

// ParsePeerID parses the peer ID if it is not empty.
func ParsePeerID(peerID string) (peer.ID, error) {
	if peerID == "" {
		return "", nil
	}

	return peer.IDB58Decode(peerID)
}

// ParsePeerIDs parses a list of peer IDs.
func ParsePeerIDs(ids []string, allowEmpty bool) ([]peer.ID, error) {
	pids := make([]peer.ID, 0, len(ids))
	for i, pidStr := range ids {
		if len(pidStr) == 0 {
			if allowEmpty {
				continue
			}
			return nil, errors.Wrapf(peer.ErrPeerIDEmpty, "peer_ids[%d]", i)
		}
		v, err := peer.IDB58Decode(pidStr)
		if err != nil {
			return nil, errors.Wrapf(err, "peer_ids[%d]", i)
		}
		pids = append(pids, v)
	}
	return pids, nil
}
