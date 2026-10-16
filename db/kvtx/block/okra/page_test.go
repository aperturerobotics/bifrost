package kvtx_block_okra

import (
	"errors"
	"testing"

	"github.com/s4wave/spacewave/db/block"
)

// TestCursorRejectsInvalidStoredPage checks that a cursor validates a decoded
// page before returning it.
func TestCursorRejectsInvalidStoredPage(t *testing.T) {
	// Store a page with no entries, which Validate rejects.
	ctx := t.Context()
	store := newOkraTestStore()
	ref, _, err := block.PutBlock(ctx, store, &Page{Level: 1})
	if err != nil {
		t.Fatal(err)
	}

	// Require the cursor decode to report the page metadata error.
	_, cursor := block.NewTransaction(store, nil, ref, nil)
	_, err = block.UnmarshalBlock[Page](ctx, cursor)
	if !errors.Is(err, ErrUnexpectedPageMetadata) {
		t.Fatalf("UnmarshalBlock error = %v, want ErrUnexpectedPageMetadata", err)
	}
}
