package peer

import (
	"crypto/ed25519"
	"crypto/rsa"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
)

// EncryptToPubKey encrypts a message to a public key.
//
// Supported types: Ed25519, RSA
// Context must be same whem decrypting, optional.
func EncryptToPubKey(pubKey crypto.PubKey, context string, msgSrc []byte) ([]byte, error) {
	spKey, err := crypto.PubKeyToStdKey(pubKey)
	if err != nil {
		return nil, err
	}

	switch t := spKey.(type) {
	case *rsa.PublicKey:
		return EncryptToRSA(t, context, msgSrc)
	case ed25519.PublicKey:
		return EncryptToEd25519(t, context, msgSrc)
	default:
		return nil, errors.Errorf("unhandled public key type: %s", pubKey.Type().String())
	}
}

// DecryptWithPrivKey decrypts with the given private key.
//
// Supported types: Ed25519, RSA
// Context must be same as when encrypting.
func DecryptWithPrivKey(privKey crypto.PrivKey, context string, ciphertext []byte) ([]byte, error) {
	spKey, err := crypto.PrivKeyToStdKey(privKey)
	if err != nil {
		return nil, err
	}

	switch t := spKey.(type) {
	case *rsa.PrivateKey:
		return DecryptWithRSA(t, context, ciphertext)
	case *ed25519.PrivateKey:
		return DecryptWithEd25519(*t, context, ciphertext)
	default:
		return nil, errors.Errorf("unhandled private key type: %s", privKey.Type().String())
	}
}
