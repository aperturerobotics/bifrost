package peer

import (
	"bytes"
	"crypto/ed25519"

	"github.com/aperturerobotics/bifrost/util/crypto/extra25519"
	curve25519_ecdh "github.com/aperturerobotics/bifrost/util/crypto/extra25519/ecdh"
	"github.com/klauspost/compress/s2"
	b58 "github.com/mr-tron/base58/base58"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
	"golang.org/x/crypto/chacha20poly1305"
)

// EncryptToEd25519 encrypts to a ed25519 key using curve25519.
//
// convert destination key to curve25519
// mix pub key into seed: blake3(context + msgSrc + encPubKey)
// generate the message priv key (ed25519) from seed
// derive the curve25519 public key from the message priv key
// use the first 24 bytes of msg pubkey curve25519 as 24-byte nonce
//
// ciphertext: msg-curve25519-pub + chacha20poly1305(s2(message))
// context must be the same when decrypting
func EncryptToEd25519(
	t ed25519.PublicKey,
	context string,
	msgSrc []byte,
) ([]byte, error) {
	if len(t) != 32 {
		return nil, errors.Errorf("unexpected ed25519 public key len: %d", len(t))
	}

	// initialize pubKey from ed25519 public key
	var pubKey [32]byte
	copy(pubKey[:], t)

	// convert pubKey to curve15529 public key encPubKey
	var encPubKey [32]byte
	valid := extra25519.PublicKeyToCurve25519(&encPubKey, &pubKey)
	if !valid {
		return nil, errors.New("invalid ed25519 key")
	}

	// mix pub key into 32-byte seed: blake3(context + msgSrc + encPubKey)
	seedHasher := blake3.NewDeriveKey("bifrost/peer encrypt curve25519 " + context)
	seedHasher.Write(msgSrc)
	seedHasher.Write(encPubKey[:])
	msgSeed := seedHasher.Sum(nil)

	// generate the message priv key (ed25519) from seed
	msgEdPrivKey := ed25519.NewKeyFromSeed(msgSeed[:])

	// convert msg ed25519 key to [64]byte
	var msgEdPrivKeyArr [64]byte
	copy(msgEdPrivKeyArr[:], msgEdPrivKey)

	// convert message priv key to curve25519 msg private key
	var msgPrivKey [32]byte
	extra25519.PrivateKeyToCurve25519(&msgPrivKey, &msgEdPrivKeyArr)

	// convert msg ed25519 pubkey to [32]byte
	var msgEdPubKey [32]byte
	copy(msgEdPubKey[:], msgEdPrivKey[32:])

	// convert message pub key to curve25519 msg pub key
	var msgPubKey [32]byte
	valid = extra25519.PublicKeyToCurve25519(&msgPubKey, &msgEdPubKey)
	if !valid {
		return nil, errors.New("generated invalid ed25519 key")
	}

	// sharedSecret is 32 bytes with (msgPrivKey, encPubKey) or (encPrivKey, msgPubKey)
	sharedSecret, err := curve25519_ecdh.ComputeSharedSecret(&msgPrivKey, &encPubKey)
	if err != nil {
		return nil, err
	}

	// compress the message
	msg := s2.EncodeBetter(nil, msgSrc)

	// use the first 24 bytes of msg pubkey curve25519 as 24-byte nonce
	msgNonce := msgPubKey[:chacha20poly1305.NonceSizeX]

	// encrypt with chacha20poly1305 cipher
	// prepend the 32-byte curve25519 one-time message key
	cipher, err := chacha20poly1305.NewX(sharedSecret)
	if err != nil {
		return nil, err
	}
	return cipher.Seal(
		msgPubKey[:],
		msgNonce,
		msg,
		msgPubKey[:],
	), nil
}

// DecryptWithEd25519 decrypts with a ed25519 key using curve25519.
//
// convert privkey to curve25519 public + private
// mix pub key into seed: blake3(context + msgSrc + encPubKey)
// generate the message priv key (ed25519) from seed
// derive the curve25519 public key from the message priv key
// use the first 24 bytes of msg pubkey curve25519 as 24-byte nonce
//
// ciphertext: msg-curve25519-pub + chacha20poly1305(s2(message))
// context must be the same as when encrypting
func DecryptWithEd25519(
	t ed25519.PrivateKey,
	context string,
	ciphertext []byte,
) ([]byte, error) {
	if len(t) != 64 {
		return nil, errors.Errorf("unexpected ed25519 private key len: %d", len(t))
	}
	if len(ciphertext) < 34 {
		return nil, ErrShortMessage
	}

	// initialize privKey from ed25519 private key
	var privKey [64]byte
	copy(privKey[:], t)

	// convert privKey to curve15529 private key encPrivKey
	var encPrivKey [32]byte
	extra25519.PrivateKeyToCurve25519(&encPrivKey, &privKey)

	// msgPubKey is the 32-byte per-message curve25519 pub key
	var msgPubKey [32]byte
	copy(msgPubKey[:], ciphertext[:32])

	// sharedSecret is 32 bytes with (encPrivKey, msgPubKey)
	sharedSecret, err := curve25519_ecdh.ComputeSharedSecret(&encPrivKey, &msgPubKey)
	if err != nil {
		return nil, err
	}

	// use the first 24 bytes of msg pubkey curve25519 as 24-byte nonce
	msgNonce := msgPubKey[:chacha20poly1305.NonceSizeX]
	msgEnc := ciphertext[32:]

	// decrypt message with shared secret
	cipher, err := chacha20poly1305.NewX(sharedSecret)
	if err != nil {
		return nil, err
	}
	msgDec, err := cipher.Open(nil, msgNonce, msgEnc, msgPubKey[:])
	if err != nil {
		return nil, err
	}

	// decompress message
	msgSrc, err := s2.Decode(nil, msgDec)
	if err != nil {
		return nil, err
	}

	// initialize pubKey from ed25519 public key
	var pubKey [32]byte
	copy(pubKey[:], t[32:])

	// convert pubKey to curve15529 public key encPubKey
	var encPubKey [32]byte
	valid := extra25519.PublicKeyToCurve25519(&encPubKey, &pubKey)
	if !valid {
		return nil, errors.New("invalid ed25519 key")
	}

	// verify message: re-generate ed25519 private key
	// mix pub key into 32-byte seed: blake3(context + msgSrc + encPubKey)
	seedHasher := blake3.NewDeriveKey("bifrost/peer encrypt curve25519 " + context)
	seedHasher.Write(msgSrc)
	seedHasher.Write(encPubKey[:])
	msgSeed := seedHasher.Sum(nil)

	// generate the message priv key (ed25519) from seed
	msgEdPrivKey := ed25519.NewKeyFromSeed(msgSeed[:])

	// convert msg ed25519 key to [32]byte pub key
	var msgEdPubKeyArr [32]byte
	copy(msgEdPubKeyArr[:], msgEdPrivKey[32:])

	// convert message pub key to curve25519 msg pub key
	var expectedMsgPubKey [32]byte
	valid = extra25519.PublicKeyToCurve25519(&expectedMsgPubKey, &msgEdPubKeyArr)
	if !valid {
		return nil, errors.New("generated invalid curve25519 pubkey")
	}

	// check that they match
	if !bytes.Equal(expectedMsgPubKey[:], msgPubKey[:]) {
		return nil, errors.Errorf(
			"message pubkey %s does not match expected pubkey %s",
			b58.Encode(msgPubKey[:]),
			b58.Encode(expectedMsgPubKey[:]),
		)
	}

	// done
	return msgSrc, nil
}
