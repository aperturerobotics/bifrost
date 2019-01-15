package peer

import (
	"github.com/aperturerobotics/bifrost/hash"
	"github.com/libp2p/go-libp2p-crypto"
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
		pkey, err := privKey.Bytes()
		if err != nil {
			return nil, err
		}
		s.PubKey = pkey
	}
	return s, nil
}
