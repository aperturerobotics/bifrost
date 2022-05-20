package randstring

import (
	"math/rand"
	"testing"
)

func TestRandString(t *testing.T) {
	rnd := rand.New(rand.NewSource(518234812384))
	str := RandString(rnd, 8)
	t.Log(str)
	if str != "NyUHvCmd" {
		t.Fail()
	}
}
