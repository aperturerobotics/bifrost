package transport_quic

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/pkg/errors"
)

// CheckAlreadyConnected checks if a address and peer id is already connected.
func CheckAlreadyConnected(t *Transport, addr string, peerID peer.ID) (bool, error) {
	lnk, ok := t.LookupLinkWithAddr(addr)
	if !ok {
		return false, nil
	}
	lnkPeer := lnk.GetRemotePeer().Pretty()
	desiredPeer := peerID.Pretty()
	if lnkPeer != desiredPeer {
		return false, errors.Errorf(
			"already connected to %s with different peer: %s != requested %s",
			addr,
			lnkPeer,
			desiredPeer,
		)
	}
	return true, nil
}
