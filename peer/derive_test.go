package peer

import (
	"testing"

	b58 "github.com/mr-tron/base58/base58"
)

// TestDerive tests deriving a context-specific key from crypto keys.
func TestDerive(t *testing.T) {
	keys := BuildMockKeys(t)
	for ki, key := range keys {
		var secret [32]byte
		err := DeriveKey("bifrost/peer/derive_test", key, secret[:])
		if err != nil {
			t.Fatal(err.Error())
		}
		t.Logf("keys[%d]: derived key: %s", ki, b58.Encode(secret[:]))
	}
}
