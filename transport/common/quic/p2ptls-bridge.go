package transport_quic

import (
	"github.com/aperturerobotics/bifrost/crypto"
	"github.com/aperturerobotics/bifrost/peer"
	lp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	lp2ppeer "github.com/libp2p/go-libp2p/core/peer"
)

// PrivKeyToLP2P converts a bifrost crypto.PrivKey to a go-libp2p crypto.PrivKey.
func PrivKeyToLP2P(key crypto.PrivKey) (lp2pcrypto.PrivKey, error) {
	data, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return nil, err
	}
	return lp2pcrypto.UnmarshalPrivateKey(data)
}

// pubKeyFromLP2P converts a go-libp2p crypto.PubKey to a bifrost crypto.PubKey.
func pubKeyFromLP2P(key lp2pcrypto.PubKey) (crypto.PubKey, error) {
	data, err := lp2pcrypto.MarshalPublicKey(key)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPublicKey(data)
}

// peerIDToLP2P converts a bifrost peer.ID to a go-libp2p peer.ID.
func peerIDToLP2P(id peer.ID) lp2ppeer.ID {
	return lp2ppeer.ID(id)
}
