package peer

import (
	"bytes"
	"crypto/aes"
	"crypto/ed25519"

	"github.com/aperturerobotics/bifrost/util/crypto/extra25519"
	curve25519_ecdh "github.com/aperturerobotics/bifrost/util/crypto/extra25519/ecdh"
	"github.com/aperturerobotics/bifrost/util/scrub"
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
// use blake3(msgPubkeyCurve25519)[:24] as the message nonce
// generate msgPubKey aes256 key: blake3(context + encPubKey + msgNonce[:4])
//
// ciphertext: msgNonce[:4] + aes256(msgPubKey) + chacha20poly1305(s2(message))
// context and destination public key must be the same when decrypting
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
	scrub.Scrub(pubKey[:])
	defer scrub.Scrub(encPubKey[:])
	if !valid {
		return nil, errors.New("invalid ed25519 key")
	}

	// mix pub key and msg src into 32-byte seed: blake3(context + msgSrc + encPubKey)
	seedHasher := blake3.NewDeriveKey("bifrost/peer encrypt curve25519 " + context)
	_, err := seedHasher.Write(msgSrc)
	if err != nil {
		return nil, err
	}
	_, err = seedHasher.Write(encPubKey[:])
	if err != nil {
		return nil, err
	}
	msgSeed := seedHasher.Sum(nil)

	// generate the message priv key (ed25519) from seed
	msgEdPrivKey := ed25519.NewKeyFromSeed(msgSeed[:])
	scrub.Scrub(msgSeed[:])

	// convert msg ed25519 key to [64]byte
	var msgEdPrivKeyArr [64]byte
	copy(msgEdPrivKeyArr[:], msgEdPrivKey)
	// convert message priv key to curve25519 msg private key
	var msgPrivKey [32]byte
	extra25519.PrivateKeyToCurve25519(&msgPrivKey, &msgEdPrivKeyArr)
	scrub.Scrub(msgEdPrivKeyArr[:])

	// convert msg ed25519 pubkey to [32]byte
	var msgEdPubKey [32]byte
	copy(msgEdPubKey[:], msgEdPrivKey[32:])
	scrub.Scrub(msgEdPrivKey)
	// convert message pub key to curve25519 msg pub key
	var msgPubKey [32]byte
	valid = extra25519.PublicKeyToCurve25519(&msgPubKey, &msgEdPubKey)
	scrub.Scrub(msgEdPubKey[:])
	defer scrub.Scrub(msgPubKey[:])
	if !valid {
		return nil, errors.New("generated invalid ed25519 key")
	}

	// compress the message
	msg := s2.EncodeBetter(nil, msgSrc)

	// blake3 hash the msg pubkey
	msgPubKeyHash := blake3.Sum256(msgPubKey[:])
	defer scrub.Scrub(msgPubKeyHash[:])

	// use the first 24 bytes of msg pubkey hash as 24-byte nonce
	msgNonce := msgPubKeyHash[:chacha20poly1305.NonceSizeX]

	// xor the remaining 8 bytes of the hash with the nonce
	xorHash := msgPubKeyHash[chacha20poly1305.NonceSizeX:]
	for i := range msgNonce {
		msgNonce[i] ^= xorHash[(i+2)%len(xorHash)]
	}

	// build the seed for the message prefix: blake3(context+" prefix", encPubKey+msgNonce[:4])
	seedHasher = blake3.NewDeriveKey("bifrost/peer encrypt curve25519 prefix " + context)
	_, err = seedHasher.Write(encPubKey[:])
	if err != nil {
		return nil, err
	}
	_, err = seedHasher.Write(msgNonce[:4])
	if err != nil {
		return nil, err
	}
	prefixSeed := seedHasher.Sum(nil)

	// encrypt the message prefix with aes256
	// allocate enough space for the ciphertext as well
	prefix := make([]byte, 4+32, len(msgPubKey)+chacha20poly1305.Overhead+len(msg)+8)
	copy(prefix[:4], msgNonce[:4])
	copy(prefix[4:], msgPubKey[:])
	prefixCipher, err := aes.NewCipher(prefixSeed[:32])
	if err != nil {
		scrub.Scrub(prefixSeed)
		return nil, err
	}
	prefixCipher.Encrypt(prefix[4:], prefix[4:])
	scrub.Scrub(prefixSeed)

	// sharedSecret is 32 bytes with (msgPrivKey, encPubKey) or (encPrivKey, msgPubKey)
	sharedSecret, err := curve25519_ecdh.ComputeSharedSecret(msgPrivKey[:], encPubKey[:])
	if err != nil {
		return nil, err
	}
	defer scrub.Scrub(sharedSecret)

	// encrypt with chacha20poly1305 cipher
	// prepend the 32-byte curve25519 one-time message key
	// re-use the prefix buf
	cipher, err := chacha20poly1305.NewX(sharedSecret)
	if err != nil {
		return nil, err
	}
	return cipher.Seal(
		prefix,
		msgNonce,
		msg,
		msgPubKey[:],
	), nil
}

// DecryptWithEd25519 decrypts with a ed25519 key using curve25519.
//
// generate msgPubKey aes256 key: blake3(context + encPubKey + ciphertext[:4])
// decrypt msgPubKey from ciphertext[4:][:32]
// convert privKey to curve25519 public + private
// derive the shared secret with (privKey, msgPubKey)
// use blake3(msgPubKey)[:24] as the message nonce
//
// ciphertext: msgNonce[:4] + aes256(msgPubKey) + chacha20poly1305(s2(message))
// context and destination public key must be the same as when encrypting
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

	// initialize pubKey from ed25519 public key
	var pubKey [32]byte
	copy(pubKey[:], t[32:])
	defer scrub.Scrub(pubKey[:])

	// convert pubKey to curve15529 public key encPubKey
	var encPubKey [32]byte
	defer scrub.Scrub(encPubKey[:])
	valid := extra25519.PublicKeyToCurve25519(&encPubKey, &pubKey)
	if !valid {
		return nil, errors.New("invalid ed25519 key")
	}

	// build the seed for the message prefix: blake3(context+" prefix", encPubKey+prefix[:4])
	seedHasher := blake3.NewDeriveKey("bifrost/peer encrypt curve25519 prefix " + context)
	_, err := seedHasher.Write(encPubKey[:])
	if err != nil {
		return nil, err
	}
	_, err = seedHasher.Write(ciphertext[:4])
	if err != nil {
		return nil, err
	}
	prefixSeed := seedHasher.Sum(nil)

	// msgPubKey is the 32-byte per-message curve25519 pub key
	var msgPubKey [32]byte
	copy(msgPubKey[:], ciphertext[4:])
	defer scrub.Scrub(msgPubKey[:])

	// decrypt the message prefix (msg pub key) with aes256
	prefixCipher, err := aes.NewCipher(prefixSeed[:32])
	if err == nil {
		prefixCipher.Decrypt(msgPubKey[:], msgPubKey[:])
	}
	scrub.Scrub(prefixSeed)
	if err != nil {
		return nil, err
	}

	// initialize privKey from ed25519 private key
	var privKey [64]byte
	copy(privKey[:], t)
	var encPrivKey [32]byte
	// convert privKey to curve15529 private key encPrivKey
	extra25519.PrivateKeyToCurve25519(&encPrivKey, &privKey)
	scrub.Scrub(privKey[:])

	// sharedSecret is 32 bytes with (encPrivKey, msgPubKey)
	sharedSecret, err := curve25519_ecdh.ComputeSharedSecret(encPrivKey[:], msgPubKey[:])
	scrub.Scrub(encPrivKey[:])
	if err != nil {
		scrub.Scrub(sharedSecret)
		return nil, err
	}

	// blake3 hash the msg pubkey
	msgPubKeyHash := blake3.Sum256(msgPubKey[:])
	defer scrub.Scrub(msgPubKeyHash[:])

	// use the first 24 bytes of msg pubkey hash as 24-byte nonce
	msgNonce := msgPubKeyHash[:chacha20poly1305.NonceSizeX]

	// xor the remaining 8 bytes of the hash with the nonce
	xorHash := msgPubKeyHash[chacha20poly1305.NonceSizeX:]
	for i := range msgNonce {
		msgNonce[i] ^= xorHash[(i+2)%len(xorHash)]
	}

	// decrypt message with shared secret
	msgEnc := ciphertext[32+4:]
	cipher, err := chacha20poly1305.NewX(sharedSecret)
	if err != nil {
		scrub.Scrub(sharedSecret)
		return nil, err
	}
	msgDec, err := cipher.Open(nil, msgNonce, msgEnc, msgPubKey[:])
	scrub.Scrub(sharedSecret)
	if err != nil {
		return nil, err
	}

	// decompress message, re-use shared secret buffer
	msgSrc, err := s2.Decode(sharedSecret, msgDec)
	if err != nil {
		return nil, err
	}

	// verify message: re-generate ed25519 private key
	// mix pub key into 32-byte seed: blake3(context + msgSrc + encPubKey)
	seedHasher = blake3.NewDeriveKey("bifrost/peer encrypt curve25519 " + context)
	_, err = seedHasher.Write(msgSrc)
	if err != nil {
		return nil, err
	}
	_, err = seedHasher.Write(encPubKey[:])
	if err != nil {
		return nil, err
	}
	msgSeed := seedHasher.Sum(nil)

	// generate the message priv key (ed25519) from seed
	msgEdPrivKey := ed25519.NewKeyFromSeed(msgSeed[:])
	scrub.Scrub(msgSeed)

	// convert msg ed25519 key to [32]byte pub key
	var msgEdPubKeyArr [32]byte
	copy(msgEdPubKeyArr[:], msgEdPrivKey[32:])
	scrub.Scrub(msgEdPrivKey)

	// convert message pub key to curve25519 msg pub key
	var expectedMsgPubKey [32]byte
	valid = extra25519.PublicKeyToCurve25519(&expectedMsgPubKey, &msgEdPubKeyArr)
	scrub.Scrub(msgEdPubKeyArr[:])
	defer scrub.Scrub(expectedMsgPubKey[:])
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
