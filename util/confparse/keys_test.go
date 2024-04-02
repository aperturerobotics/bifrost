package confparse

import (
	"crypto/rand"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// TestParseKeys tests parsing public and private key pems.
func TestParseKeys(t *testing.T) {
	keyPriv, keyPub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err.Error())
	}

	privStr, err := MarshalPrivateKey(keyPriv)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("private key: %s", privStr)

	pubStr, err := MarshalPublicKey(keyPub)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("public key: %s", pubStr)

	// parse
	privOut, err := ParsePrivateKey(privStr)
	if err != nil {
		t.Fatal(err.Error())
	}
	if !privOut.Equals(keyPriv) {
		t.Fail()
	}

	// parse
	pubOut, err := ParsePublicKey(pubStr)
	if err != nil {
		t.Fatal(err.Error())
	}
	if !pubOut.Equals(keyPub) {
		t.Fail()
	}
}
