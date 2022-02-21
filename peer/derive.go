package peer

import (
	"crypto/ed25519"
	"crypto/rsa"

	"github.com/aperturerobotics/bifrost/util/crypto/extra25519"
	curve25519_ecdh "github.com/aperturerobotics/bifrost/util/crypto/extra25519/ecdh"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
)

// DeriveKey derives a crypto key using a private key.
//
// Not all private key types are supported.
// Context string must be unique to the situation.
// Data is written to out.
func DeriveKey(context string, privKey crypto.PrivKey, out []byte) error {
	spKey, err := crypto.PrivKeyToStdKey(privKey)
	if err != nil {
		return err
	}

	var material []byte
	switch t := spKey.(type) {
	case *rsa.PrivateKey:
		rawKey, err := privKey.Raw()
		if err != nil {
			return err
		}
		rawKeyHash := blake3.Sum512(rawKey)
		material = rawKeyHash[:]
	case *ed25519.PrivateKey:
		// convert private key to curve25519
		var recvPrivKey [32]byte
		var tPrivKey [64]byte
		copy(tPrivKey[:], (*t)[:])
		extra25519.PrivateKeyToCurve25519(&recvPrivKey, &tPrivKey)

		// hash the context + the curve25519 key
		secret := append(tPrivKey[:], []byte(context)...)
		seed := blake3.Sum256(secret)

		// generate a new ephemeral private / public key
		ephPrivKey := ed25519.NewKeyFromSeed(seed[:])
		var ephPubKey [32]byte
		copy(ephPubKey[:], ephPrivKey[32:])
		var msgPubKey [32]byte
		valid := extra25519.PublicKeyToCurve25519(&msgPubKey, &ephPubKey)
		if !valid {
			return errors.New("generated invalid ed25519 key")
		}

		// derive a shared secret w/ ephemeral & private key
		material, err = curve25519_ecdh.ComputeSharedSecret(recvPrivKey[:], msgPubKey[:])
		if err != nil {
			return err
		}
	default:
		return errors.Errorf("unhandled private key type: %s", privKey.Type().String())
	}

	// xor the material with context
	contextb := []byte(context)
	for i := range material {
		material[i] = material[i] ^ contextb[i%len(contextb)]
	}

	// derive key with blake3
	blake3.DeriveKey(context, material, out)
	return nil
}

// DeriveEd25519Key derives a ed25519 private key from an existing private key.
//
// The context string will be mixed to determine which key is generated.
// Not all private key types are supported.
func DeriveEd25519Key(context string, privKey crypto.PrivKey) (crypto.PrivKey, crypto.PubKey, error) {
	seed := make([]byte, ed25519.SeedSize)
	if err := DeriveKey(context, privKey, seed); err != nil {
		return nil, nil, err
	}

	key := ed25519.NewKeyFromSeed(seed)
	return crypto.KeyPairFromStdKey(&key)
}
