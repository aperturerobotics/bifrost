package nats

import (
	"errors"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/nats-io/nkeys"
)

// KeyPair wraps a Bifrost key pair.
type KeyPair struct {
	privKey crypto.PrivKey
	pubKey  crypto.PubKey
}

// NewKeyPair builds a new keypair. Public key must be specified unless private
// key is specified. Private key must be specified if public key is not.
func NewKeyPair(privKey crypto.PrivKey, pubKey crypto.PubKey) (*KeyPair, error) {
	// derive pub key if priv is given
	if privKey != nil && pubKey == nil {
		pubKey = privKey.GetPublic()
	}
	if pubKey == nil {
		if privKey != nil {
			return nil, errors.New("could not derive public key from priv key")
		}
		return nil, errors.New("no key was specified")
	}
	return &KeyPair{privKey: privKey, pubKey: pubKey}, nil
}

// Seed returns the seed used to make the key.
func (k *KeyPair) Seed() ([]byte, error) {
	return nil, errors.New("BIFROST TODO: nkeys: seed unavailable")
}

// PublicKey will return the encoded public key associated with the KeyPair.
// All KeyPairs have a public key.
func (k *KeyPair) PublicKey() (string, error) {
	// public key is embedded in peer ID.
	// for NATS we want server ID == peer ID.
	/*
		mpk, err := crypto.MarshalPublicKey(k.pubKey)
		if err != nil {
			return "", err
		}
		return b58.Encode(mpk), nil
	*/
	peerID, err := peer.IDFromPublicKey(k.pubKey)
	if err != nil {
		return "", err
	}
	return peerID.Pretty(), nil
}

// PrivateKey will return an error since this is not available for public key only KeyPairs.
func (k *KeyPair) PrivateKey() ([]byte, error) {
	if k.privKey == nil {
		return nil, nkeys.ErrPublicKeyOnly
	}
	mpk, err := crypto.MarshalPrivateKey(k.privKey)
	if err != nil {
		return nil, err
	}
	return mpk, nil
}

// Sign will return an error since this is not available for public key only KeyPairs.
func (k *KeyPair) Sign(input []byte) ([]byte, error) {
	if k.privKey == nil {
		return nil, nkeys.ErrPublicKeyOnly
	}
	return k.privKey.Sign(input)
}

// Verify will verify the input against a signature utilizing the public key.
func (k *KeyPair) Verify(input []byte, sig []byte) error {
	if k.pubKey == nil {
		return nkeys.ErrInvalidPublicKey
	}
	ok, err := k.pubKey.Verify(input, sig)
	if err != nil {
		return err
	}
	if !ok {
		return nkeys.ErrInvalidSignature
	}
	return nil
}

func (k *KeyPair) Wipe() {
	// finalizers should wipe
	k.privKey = nil
	k.pubKey = nil
}

// _ is a type assertion
var _ nkeys.KeyPair = ((*KeyPair)(nil))
