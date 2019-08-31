package confparse

import (
	"errors"

	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// ParsePrivateKey parses the private key from a configuration.
// If there is no private key specified, returns nil, nil.
func ParsePrivateKey(privKeyDat []byte) (crypto.PrivKey, error) {
	if len(privKeyDat) == 0 {
		return nil, nil
	}

	key, err := keypem.ParsePrivKeyPem(privKeyDat)
	if err != nil {
		return nil, err
	}

	if key == nil {
		return nil, errors.New("no pem data found")
	}

	return key, nil
}
