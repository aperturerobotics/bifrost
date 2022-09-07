package peer

import (
	"crypto/rand"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
)

// BuildMockKeys builds the set of mock keys that are expected to work.
func BuildMockKeys(t *testing.T) []crypto.PrivKey {
	keys := []crypto.PrivKey{}

	edPriv, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err.Error())
	}
	keys = append(keys, edPriv)

	rPriv, _, err := crypto.GenerateRSAKeyPair(2048, rand.Reader)
	if err != nil {
		t.Fatal(err.Error())
	}
	keys = append(keys, rPriv)

	return keys
}
