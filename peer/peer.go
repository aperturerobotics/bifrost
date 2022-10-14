package peer

import (
	"context"
	"crypto/rand"

	"github.com/libp2p/go-libp2p/core/crypto"
	lpeer "github.com/libp2p/go-libp2p/core/peer"
)

// Peer is the common interface for a keypair-based identity.
type Peer interface {
	// GetPeerID returns the peer ID.
	GetPeerID() ID

	// GetPubKey returns the public key of the peer.
	GetPubKey() crypto.PubKey

	// GetPrivKey returns the private key.
	// This may require an extra lookup operation.
	// Returns ErrNoPrivKey if the private key is unavailable.
	GetPrivKey(ctx context.Context) (crypto.PrivKey, error)
}

// NewPeer builds a new Peer object with a private key.
// If privKey is nil, one will be generated.
func NewPeer(privKey crypto.PrivKey) (Peer, error) {
	if privKey == nil {
		var err error
		privKey, _, err = crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return nil, err
		}
	}

	id, err := lpeer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	return &peer{
		privKey: privKey,
		pubKey:  privKey.GetPublic(),
		peerID:  id,
	}, nil
}

// NewPeerWithPubKey builds a Peer with a public key.
func NewPeerWithPubKey(pubKey crypto.PubKey) (Peer, error) {
	id, err := lpeer.IDFromPublicKey(pubKey)
	if err != nil {
		return nil, err
	}
	return &peer{
		pubKey: pubKey,
		peerID: id,
	}, nil
}

// NewPeerWithID constructs a new Peer by extracting the pubkey from the ID.
func NewPeerWithID(id lpeer.ID) (Peer, error) {
	pubKey, err := id.ExtractPublicKey()
	if err != nil {
		return nil, err
	}
	return NewPeerWithPubKey(pubKey)
}

// peer implements Peer with an in-memory struct.
type peer struct {
	privKey crypto.PrivKey
	pubKey  crypto.PubKey
	peerID  ID
}

// GetPeerID returns the peer ID.
func (p *peer) GetPeerID() ID {
	return p.peerID
}

// GetPubKey returns the public key of the peer.
func (p *peer) GetPubKey() crypto.PubKey {
	return p.pubKey
}

// GetPrivKey returns the private key.
// May be empty if peer private key is unavailable.
func (p *peer) GetPrivKey(ctx context.Context) (crypto.PrivKey, error) {
	if p.privKey == nil {
		return nil, ErrNoPrivKey
	}
	return p.privKey, nil
}
