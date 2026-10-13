package sobject_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"testing"

	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/net/peer"
)

// TestSOCheckpointHash pins the checkpoint identity: a domain-framed digest of
// the signed body that excludes the signatures.
func TestSOCheckpointHash(t *testing.T) {
	// Hash a checkpoint with one signature.
	checkpoint := &sobject.SOCheckpoint{
		Inner:      []byte{0x01, 0x02},
		Signatures: []*peer.Signature{{SigData: []byte("old")}},
	}
	preimage := append([]byte("spacewave/sharedobject/checkpoint/v1"), make([]byte, 8)...)
	binary.BigEndian.PutUint64(preimage[len(preimage)-8:], uint64(len(checkpoint.Inner)))
	preimage = append(preimage, checkpoint.Inner...)
	want := sha256.Sum256(preimage)
	got := checkpoint.Hash()
	if !bytes.Equal(got, want[:]) {
		t.Fatalf("hash = %x, want %x", got, want)
	}

	// Adding a signature leaves the hash unchanged.
	checkpoint.Signatures = append(checkpoint.Signatures, &peer.Signature{SigData: []byte("new")})
	if !bytes.Equal(checkpoint.Hash(), got) {
		t.Fatal("checkpoint hash must exclude signatures")
	}
	checkpoint.Inner[0]++
	if bytes.Equal(checkpoint.Hash(), got) {
		t.Fatal("checkpoint hash must bind the body")
	}
}

// TestSOOperationHash pins the operation identity: a domain-framed digest of
// the signed body that excludes the signature, for bodies of any size.
func TestSOOperationHash(t *testing.T) {
	for _, size := range []int{0, 1, 63, 64, 65, 1 << 16} {
		// Hash the preimage as one buffer.
		inner := bytes.Repeat([]byte{0x5a}, size)
		preimage := append([]byte("spacewave/sharedobject/operation/v1"), make([]byte, 8)...)
		binary.BigEndian.PutUint64(preimage[len(preimage)-8:], uint64(size))
		preimage = append(preimage, inner...)
		want := sha256.Sum256(preimage)

		// Both entry points hash the body without a signature.
		op := &sobject.SOOperation{Inner: inner, Signature: &peer.Signature{SigData: []byte("sig")}}
		if got := op.Hash(); !bytes.Equal(got, want[:]) {
			t.Fatalf("size %d: hash = %x, want %x", size, got, want)
		}
		if got := sobject.HashSOOperationInner(inner); !bytes.Equal(got, want[:]) {
			t.Fatalf("size %d: body hash = %x, want %x", size, got, want)
		}
	}
}
