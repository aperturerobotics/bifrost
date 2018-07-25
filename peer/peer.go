package peer

import (
	"crypto/rand"

	"github.com/libp2p/go-libp2p-crypto"
)

// Peer implements common functionalities between peer types.
// Includes: identity.
type Peer interface {
	// GetPubKey returns the public key of the node.
	GetPubKey() crypto.PubKey

	// GetPrivKey returns the private key.
	GetPrivKey() crypto.PrivKey

	// GetPeerID returns the peer ID.
	GetPeerID() ID
}

// NewPeer builds a new Peer object.
// If privKey is nil, one will be generated.
func NewPeer(privKey crypto.PrivKey) (Peer, error) {
	if privKey == nil {
		var err error
		privKey, _, err = crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return nil, err
		}
	}

	id, err := IDFromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	return &peer{
		privKey: privKey,
		peerID:  id,
	}, nil
}

type peer struct {
	privKey crypto.PrivKey
	peerID  ID
}

// GetPubKey returns the public key of the node.
func (p *peer) GetPubKey() crypto.PubKey {
	return p.privKey.GetPublic()
}

// GetPrivKey returns the private key.
func (p *peer) GetPrivKey() crypto.PrivKey {
	return p.privKey
}

// GetPeerID returns the peer ID.
func (p *peer) GetPeerID() ID {
	return p.peerID
}
