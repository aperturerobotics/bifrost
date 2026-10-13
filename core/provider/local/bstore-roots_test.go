package provider_local

import (
	"errors"
	"testing"

	"github.com/s4wave/spacewave/db/block"
)

// TestPinRootUnknownRefs checks that fetching an untracked root cannot write it
// into the destination as a leaf.
func TestPinRootUnknownRefs(t *testing.T) {
	// Serve an untracked block from a store that retains only its bytes.
	ctx := t.Context()
	source := newBatchForwardTestStore()
	ref, _, err := source.PutBlock(ctx, []byte("untracked root"), nil)
	if err != nil {
		t.Fatal(err)
	}
	destination := newBatchForwardTestStore()
	store := &BlockStore{store: destination, readStore: source}

	// Refuse the pin before writing a block with unknown references.
	release, err := store.PinRoot(ctx, ref)
	if release != nil {
		release()
	}
	if !errors.Is(err, block.ErrRefsUnknown) {
		t.Fatalf("pin = %v, want ErrRefsUnknown", err)
	}
	found, err := destination.GetBlockExists(ctx, ref)
	if err != nil || found {
		t.Fatalf("destination = %v/%v, want absent", found, err)
	}
}
