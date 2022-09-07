package keypem

import (
	"encoding/pem"
	"errors"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// PrivPemType is the expected header type on private keys.
const PrivPemType = "LIBP2P PRIVATE KEY"

// PubPemType is the expected header type on public keys.
const PubPemType = "LIBP2P PUBLIC KEY"

// ParsePrivKeyPem parses a private key in pem format.
// If none is found returns nil
func ParsePrivKeyPem(pemDat []byte) (crypto.PrivKey, error) {
	b, _ := pem.Decode(pemDat)
	if b == nil {
		return nil, nil
	}

	if b.Type != PrivPemType {
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
		Type:  PrivPemType,
		Bytes: dat,
	}), nil
}

// ParsePubKeyPem parses a public key in pem format.
// If none is found returns nil
func ParsePubKeyPem(pemDat []byte) (crypto.PubKey, error) {
	b, _ := pem.Decode(pemDat)
	if b == nil {
		return nil, nil
	}

	if b.Type != PubPemType {
		return nil, errors.New("unexpected pem type for public key")
	}

	return crypto.UnmarshalPublicKey(b.Bytes)
}

// MarshalPubKeyPem marshals a public key to pem.
func MarshalPubKeyPem(key crypto.PubKey) ([]byte, error) {
	dat, err := crypto.MarshalPublicKey(key)
	if err != nil {
		return nil, err
	}

	return pem.EncodeToMemory(&pem.Block{
		Type:  PubPemType,
		Bytes: dat,
	}), nil
}
