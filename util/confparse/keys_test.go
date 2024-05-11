package confparse

import (
	"testing"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/libp2p/go-libp2p/core/crypto"
)

// TestParseKeys tests parsing public and private key pems.
func TestParseKeys(t *testing.T) {
	testKeyTypes(t, testParseKeys)
}

func testParseKeys(t *testing.T, keyPriv crypto.PrivKey, keyPub crypto.PubKey) {
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

	id, err := peer.IDFromPublicKey(pubOut)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("peer id: %v", id.String())
}

func testKeyTypes(t *testing.T, testFunc func(*testing.T, crypto.PrivKey, crypto.PubKey)) {
	t.Run("RSA", func(t *testing.T) {
		priv, pub, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
		if err != nil {
			t.Fatal(err)
		}
		testFunc(t, priv, pub)
	})
	t.Run("Ed25519", func(t *testing.T) {
		priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
		if err != nil {
			t.Fatal(err)
		}
		testFunc(t, priv, pub)
	})
	t.Run("EdDilithium2", func(t *testing.T) {
		priv, pub, err := crypto.GenerateKeyPair(crypto.EdDilithium2, 0)
		if err != nil {
			t.Fatal(err)
		}
		testFunc(t, priv, pub)
	})
	t.Run("EdDilithium3", func(t *testing.T) {
		priv, pub, err := crypto.GenerateKeyPair(crypto.EdDilithium3, 0)
		if err != nil {
			t.Fatal(err)
		}
		testFunc(t, priv, pub)
	})
}
