package s2s

import (
	"context"
	"crypto/rand"

	"github.com/libp2p/go-libp2p-crypto"
	"golang.org/x/crypto/nacl/box"
	// "golang.org/x/crypto/nacl/secretbox"
	// "golang.org/x/crypto/ripemd160"
)

// Handshaker implements the Station to Station handshake protocol.
type Handshaker struct {
	// privKey is the local private key.
	privKey crypto.PrivKey
}

// NewHandshaker builds a new handshaker.
func NewHandshaker(privKey crypto.PrivKey) *Handshaker {
	return &Handshaker{privKey: privKey}
}

// HandshakeIdentity performs a key-exchange and secret negotiation handshake.
// Returns the generated secret and any error.
func (h *Handshaker) HandshakeIdentity(ctx context.Context) ([]byte, crypto.PubKey, error) {
	// hPub is the handshake public key.
	hPub, hPriv, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	_ = hPub
	_ = hPriv
	return nil, nil, nil
}
