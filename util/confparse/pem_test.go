package confparse

import (
	"crypto/rand"
	"testing"

	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/libp2p/go-libp2p-core/crypto"
)

// TestParseKeysPEM tests parsing public and private key pems.
func TestParseKeysPEM(t *testing.T) {
	keyPriv, keyPub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err.Error())
	}

	privPEM, err := keypem.MarshalPrivKeyPem(keyPriv)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("private key pem: %s", string(privPEM))

	pubPEM, err := keypem.MarshalPubKeyPem(keyPub)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("public key pem: %s", string(pubPEM))

	// parse
	privOut, err := ParsePrivateKey(string(privPEM))
	if err != nil {
		t.Fatal(err.Error())
	}
	if !privOut.Equals(keyPriv) {
		t.Fail()
	}

	// parse
	pubOut, err := ParsePublicKey(string(pubPEM))
	if err != nil {
		t.Fatal(err.Error())
	}
	if !pubOut.Equals(keyPub) {
		t.Fail()
	}
}
