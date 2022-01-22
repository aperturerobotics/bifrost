package peer

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"

	"github.com/aperturerobotics/bifrost/util/crypto/extra25519"
	curve25519_ecdh "github.com/aperturerobotics/bifrost/util/crypto/extra25519/ecdh"
	"github.com/klauspost/compress/s2"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptToPubKey encrypts a message to a public key
//
// Not all public key types are supported.
//
// RSA: The message must be no longer than the length of the public modulus
// minus twice the hash length, minus a further 2.
func EncryptToPubKey(pubKey crypto.PubKey, msgSrc []byte) ([]byte, error) {
	spKey, err := crypto.PubKeyToStdKey(pubKey)
	if err != nil {
		return nil, err
	}

	msg := s2.EncodeBetter(nil, msgSrc)

	switch t := spKey.(type) {
	case *rsa.PublicKey:
		encHash := blake3.New()
		return rsa.EncryptOAEP(encHash, rand.Reader, t, msg, nil)
	case ed25519.PublicKey:
		// Encrypt using ed25519:
		//  - convert to curve25519 key
		//  - hash the message with blake3
		//  - generate a nonce (NonceSizeX=24 bytes) with blake3 derivekey
		//  - blake3 hash the nonce to generate a 32 byte seed
		/// - generate a temporary message ed25519 key with seed
		//  - generate shared secret from message key + target key
		//  - use shared secret to perform XCHACHA20_POLY1305 crypt
		// s2: snappy2 compression by klauspost/compress
		// Ciphertext: chacha20poly1305(s2(message))
		// Return value: s2(nonce (32 bytes) + ciphertext)
		if len(t) != 32 {
			return nil, errors.Errorf("unexpected ed25519 public key len: %d", len(t))
		}

		var pubKey [32]byte
		copy(pubKey[:], t)
		var encPubKey [32]byte
		valid := extra25519.PublicKeyToCurve25519(&encPubKey, &pubKey)
		if !valid {
			return nil, errors.New("invalid ed25519 key")
		}

		// hash message
		msgh := blake3.Sum256(msg)
		var msgNonce [chacha20poly1305.NonceSizeX]byte
		blake3.DeriveKey("bifrost/peer encrypt", msgh[:], msgNonce[:])

		msgKeySeed := blake3.Sum256(msgNonce[:])
		msgEdPrivKey := ed25519.NewKeyFromSeed(msgKeySeed[:])
		if err != nil {
			return nil, err
		}

		var msgEdPrivKeyArr [64]byte
		copy(msgEdPrivKeyArr[:], msgEdPrivKey)

		var msgPrivKey [32]byte
		extra25519.PrivateKeyToCurve25519(&msgPrivKey, &msgEdPrivKeyArr)

		// sharedSecret is 32 bytes
		sharedSecret, err := curve25519_ecdh.ComputeSharedSecret(&msgPrivKey, &encPubKey)
		if err != nil {
			return nil, err
		}

		cipher, err := chacha20poly1305.NewX(sharedSecret)
		if err != nil {
			return nil, err
		}
		enc := cipher.Seal(msgNonce[:], msgNonce[:], msg, nil)
		enc = s2.Encode(nil, enc)
		return enc, nil
	default:
		return nil, errors.Errorf("unhandled public key type: %s", pubKey.Type().String())
	}
}

// DecryptWithPrivKey decrypts with the given private key.
//
// Not all private key types are supported.
//
// RSA: The message must be no longer than the length of the public modulus
// minus twice the hash length, minus a further 2.
func DecryptWithPrivKey(privKey crypto.PrivKey, ciphertext []byte) ([]byte, error) {
	spKey, err := crypto.PrivKeyToStdKey(privKey)
	if err != nil {
		return nil, err
	}

	switch t := spKey.(type) {
	case *rsa.PrivateKey:
		encHash := blake3.New()
		decData, err := rsa.DecryptOAEP(encHash, rand.Reader, t, ciphertext, nil)
		if err != nil {
			return nil, err
		}
		return s2.Decode(nil, decData)
	case *ed25519.PrivateKey:
		encMsg, err := s2.Decode(nil, ciphertext)
		if err != nil {
			return nil, err
		}
		if len(encMsg) < chacha20poly1305.NonceSizeX+1 {
			return nil, errors.Errorf("message is too short: %d", len(encMsg))
		}

		// nonce is 24 bytes
		msgNonce := encMsg[:chacha20poly1305.NonceSizeX]
		encTxt := encMsg[chacha20poly1305.NonceSizeX:]

		// derive message temporary sender key from nonce
		msgKeySeed := blake3.Sum256(msgNonce)
		msgEdPrivKey := ed25519.NewKeyFromSeed(msgKeySeed[:])
		if err != nil {
			return nil, err
		}

		var msgEdPubKey [32]byte
		copy(msgEdPubKey[:], msgEdPrivKey[32:])

		var msgPubKey [32]byte
		valid := extra25519.PublicKeyToCurve25519(&msgPubKey, &msgEdPubKey)
		if !valid {
			return nil, errors.New("generated invalid ed25519 key")
		}

		// decode message using private key
		var recvPrivKey [32]byte
		var tPrivKey [64]byte
		copy(tPrivKey[:], (*t)[:])
		extra25519.PrivateKeyToCurve25519(&recvPrivKey, &tPrivKey)

		// sharedSecret is 32 bytes
		sharedSecret, err := curve25519_ecdh.ComputeSharedSecret(&recvPrivKey, &msgPubKey)
		if err != nil {
			return nil, err
		}

		cipher, err := chacha20poly1305.NewX(sharedSecret)
		if err != nil {
			return nil, err
		}

		decData, err := cipher.Open(nil, msgNonce, encTxt, nil)
		if err != nil {
			return nil, err
		}

		// re-use encMsg memory
		return s2.Decode(encMsg[:cap(encMsg)], decData)
	default:
		return nil, errors.Errorf("unhandled private key type: %s", privKey.Type().String())
	}
}
