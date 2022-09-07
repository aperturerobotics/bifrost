package peer

import (
	"github.com/aperturerobotics/bifrost/hash"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
)

// NewSignature constructs a signature.
func NewSignature(
	privKey crypto.PrivKey,
	hashType hash.HashType,
	data []byte,
	inclPubKey bool,
) (*Signature, error) {
	h, err := hash.Sum(hashType, data)
	if err != nil {
		return nil, err
	}
	sd, err := privKey.Sign(h.GetHash())
	if err != nil {
		return nil, err
	}
	s := &Signature{HashType: hashType, SigData: sd}
	if inclPubKey {
		pkey, err := crypto.MarshalPublicKey(privKey.GetPublic())
		if err != nil {
			return nil, err
		}
		s.PubKey = pkey
	}
	return s, nil
}

// Validate checks the signature object (but not the signature itself).
func (s *Signature) Validate() error {
	if err := s.GetHashType().Validate(); err != nil {
		return err
	}
	if len(s.GetSigData()) == 0 {
		return ErrSignatureInvalid
	}
	if len(s.GetPubKey()) != 0 {
		if _, err := s.ParsePubKey(); err != nil {
			return errors.Wrap(err, "pub_key")
		}
	}
	return nil
}

// VerifyWithPublic checks a signature with a public key, hashing the data.
// Returns ok and any error interpeting the signature.
func (s *Signature) VerifyWithPublic(pubKey crypto.PubKey, data []byte) (bool, error) {
	ht := s.GetHashType()
	if ht == hash.HashType_HashType_UNKNOWN {
		return false, errors.New("hash type missing")
	}
	if len(s.GetSigData()) == 0 {
		return false, errors.New("signature empty")
	}
	if err := ht.Validate(); err != nil {
		return false, err
	}

	dataHash, err := hash.Sum(ht, data)
	if err != nil {
		return false, err
	}
	return pubKey.Verify(dataHash.GetHash(), s.GetSigData())
}

// ParsePubKey parses the incldued public key.
// Returns nil, nil if the pub key field was not set.
func (s *Signature) ParsePubKey() (crypto.PubKey, error) {
	pubKey := s.GetPubKey()
	if len(pubKey) == 0 {
		return nil, nil
	}
	return crypto.UnmarshalPublicKey(pubKey)
}
