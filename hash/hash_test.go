package hash

import (
	"testing"
)

// TestVerifyData tests verifying some data.
func TestVerifyData(t *testing.T) {
	data := []byte("hello world")
	h, err := Sum(HashType_HashType_SHA256, data)
	if err != nil {
		t.Fatal(err.Error())
	}
	if _, err := h.VerifyData(data); err != nil {
		t.Fatal(err.Error())
	}
}
