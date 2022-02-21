package curve25519_ecdh

import (
	curve25519 "golang.org/x/crypto/curve25519"
)

// DeriveKey derives a shared key from a public/private curve25519.
//
// use PublicKeyToCurve25519 to convert a Ed25519 key.
func ComputeSharedSecret(privKey, pubKey []byte) ([]byte, error) {
	return curve25519.X25519(privKey, pubKey)
}
