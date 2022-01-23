package peer

import (
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"

	"github.com/klauspost/compress/s2"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
	"golang.org/x/crypto/chacha20poly1305"
)

// rsaOaepLen is the constant length of oaep for rsa.
const rsaOaepLen = 256

// EncryptToRSA encrypts a message to a RSA public key.
//
// marshal public key to pkix
// derive 32byte message key with blake3(context + msgSrc + pubPkix)
// derive 32byte message nonce with blake3(context + msgKey + pubPkix)
// compress message with s2 (snappy2)
// encrypt message with chacha20-poly1305
//
// ciphertext: oaep(message-key) + chacha20poly1305(s2(msgSrc))
//
// context must be the same at decrypt time
func EncryptToRSA(
	t *rsa.PublicKey,
	context string,
	msgSrc []byte,
) ([]byte, error) {
	pubPkix, err := x509.MarshalPKIXPublicKey(t)
	if err != nil {
		return nil, err
	}

	// derive message key with blake3(context + msgSrc + pubPkix)
	seedHasher := blake3.NewDeriveKey("bifrost/peer encrypt rsa " + context)
	seedHasher.Write(msgSrc)
	seedHasher.Write(pubPkix)
	msgKey := seedHasher.Sum(nil)

	// derive message nonce with blake3(context + msgKey + pubPkix)
	nonceHasher := blake3.NewDeriveKey("bifrost/peer nonce rsa " + context)
	nonceHasher.Write(msgKey)
	nonceHasher.Write(pubPkix)
	msgNonce := nonceHasher.Sum(nil)[:chacha20poly1305.NonceSizeX]

	// seed a random generator for oaep with blake3(message key + pubPkix)
	oaepRnd := blake3.New()
	oaepRnd.Write(msgKey)
	oaepRnd.Write(pubPkix)
	oaepSeed := oaepRnd.Digest()

	// encrypt message key with oaep(message-key)
	encHash := sha256.New()
	msgOaep, err := rsa.EncryptOAEP(encHash, oaepSeed, t, msgKey, pubPkix)
	if err != nil {
		return nil, err
	}
	if len(msgOaep) != rsaOaepLen {
		return nil, errors.Errorf(
			"generated message oaep len %d != expected %d",
			len(msgOaep),
			rsaOaepLen,
		)
	}

	// encrypt message with chacha20poly1305
	cipher, err := chacha20poly1305.NewX(msgKey)
	if err != nil {
		return nil, err
	}

	// prepend the oaep packet
	msg := s2.EncodeBetter(nil, msgSrc)
	return cipher.Seal(msgOaep, msgNonce, msg, pubPkix), nil
}

// DecryptWithRSA decrypts a message with a RSA private key.
//
// context must be the same as at encrypt time
func DecryptWithRSA(
	t *rsa.PrivateKey,
	context string,
	ciphertext []byte,
) ([]byte, error) {
	if len(ciphertext) < rsaOaepLen+2 {
		return nil, ErrShortMessage
	}

	tPub := &t.PublicKey
	pubPkix, err := x509.MarshalPKIXPublicKey(tPub)
	if err != nil {
		return nil, err
	}

	// decrypt the prepended oaep msgKey
	encHash := sha256.New()
	msgOaep := ciphertext[:rsaOaepLen]
	msgEnc := ciphertext[rsaOaepLen:]
	msgKey, err := rsa.DecryptOAEP(encHash, crand.Reader, t, msgOaep, pubPkix)
	if err != nil {
		return nil, err
	}

	// derive message nonce with blake3(context + msgKey + pubPkix)
	nonceHasher := blake3.NewDeriveKey("bifrost/peer nonce rsa " + context)
	nonceHasher.Write(msgKey)
	nonceHasher.Write(pubPkix)
	msgNonce := nonceHasher.Sum(nil)[:chacha20poly1305.NonceSizeX]

	// decrypt message with chacha20poly1305
	cipher, err := chacha20poly1305.NewX(msgKey)
	if err != nil {
		return nil, err
	}
	msgDec, err := cipher.Open(nil, msgNonce, msgEnc, pubPkix)
	if err != nil {
		return nil, err
	}

	// decompress message & return
	// re-use pubPkix buffer
	return s2.Decode(pubPkix[:cap(pubPkix)], msgDec)
}
