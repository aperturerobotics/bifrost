package peer

import (
	"crypto/rand"

	"github.com/libp2p/go-libp2p-crypto"
)

// Peer implements common functionalities between peer types.
// Includes: identity.
type Peer struct {
	// privKey is the private key
	privKey crypto.PrivKey
	// peerID is the local peer id
	peerID ID
}

// NewPeer builds a new Peer object.
// If privKey is nil, one will be generated.
func NewPeer(privKey crypto.PrivKey) (*Peer, error) {
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

	return &Peer{
		privKey: privKey,
		peerID:  id,
	}, nil
}

// GetPubKey returns the public key of the node.
func (p *Peer) GetPubKey() crypto.PubKey {
	return p.privKey.GetPublic()
}

// GetPrivKey returns the private key.
func (p *Peer) GetPrivKey() crypto.PrivKey {
	return p.privKey
}

// GetPeerID returns the peer ID.
func (p *Peer) GetPeerID() ID {
	return p.peerID
}
