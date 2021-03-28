package hash

import (
	"testing"

	"github.com/pkg/errors"
)

// TestVerifyData tests verifying some data with each hash type.
func TestVerifyData(t *testing.T) {
	data := []byte("hello world")
	for _, ht := range SupportedHashTypes {
		h, err := Sum(ht, data)
		werr := func(e error) error {
			return errors.Wrapf(e, "hash_type[%v]", ht)
		}
		if err != nil {
			t.Fatal(werr(err))
		}
		if _, err := h.VerifyData(data); err != nil {
			t.Fatal(werr(err))
		}
		t.Logf("OK: %s", ht.String())
	}
}
