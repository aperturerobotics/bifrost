package keypem

import (
	"crypto/rand"
	"encoding/pem"
	"errors"

	"github.com/libp2p/go-libp2p-crypto"
)

// PemType is the expected header type on private keys.
const PemType = "LIBP2P PRIVATE KEY"

// ParsePrivKeyPem parses a private key in pem format.
// If none is found returns nil
func ParsePrivKeyPem(pemDat []byte) (crypto.PrivKey, error) {
	b, _ := pem.Decode(pemDat)
	if b == nil {
		return nil, nil
	}

	if b.Type != PemType {
		return nil, errors.New("unexpected pem type for private key")
	}

	return crypto.UnmarshalPrivateKey(b.Bytes)
}

// MarshalPrivKeyPem marshals a private key to pem.
func MarshalPrivKeyPem(key crypto.PrivKey) ([]byte, error) {
	dat, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  PemType,
		Bytes: dat,
	}), nil
}

// GeneratePrivKey generates a new private key to use.
// The kind of private key is not specified.
func GeneratePrivKey() (crypto.PrivKey, crypto.PubKey, error) {
	return crypto.GenerateEd25519Key(rand.Reader)
}
