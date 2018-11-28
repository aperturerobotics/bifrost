package api

import (
	"github.com/aperturerobotics/bifrost/peer"
)

// NewPeerInfo builds peer info from a peer.
func NewPeerInfo(p peer.Peer) (*PeerInfo, error) {
	pi := &PeerInfo{}
	pi.PeerId = peer.IDB58Encode(p.GetPeerID())
	return pi, nil
}
