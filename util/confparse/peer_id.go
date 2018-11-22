package confparse

import "github.com/aperturerobotics/bifrost/peer"

// ParsePeerID parses the peer ID if it is not empty.
func ParsePeerID(peerID string) (peer.ID, error) {
	if peerID == "" {
		return "", nil
	}

	return peer.IDB58Decode(peerID)
}
