package curve25519_ecdh

import (
	curve25519 "golang.org/x/crypto/curve25519"
)

// DeriveKey derives a shared key from a public/private curve25519.
//
// use PublicKeyToCurve25519 to convert a Ed25519 key.
func ComputeSharedSecret(privKey, pubKey *[32]byte) ([]byte, error) {
	var secret [32]byte
	curve25519.ScalarMult(&secret, privKey, pubKey)
	return secret[:], nil
}
