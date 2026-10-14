package journal

import (
	"slices"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_gc "github.com/s4wave/spacewave/db/block/gc"
	"github.com/s4wave/spacewave/net/hash"
)

// TestUnmarshalLegacyBlockIRI decodes an entry that spells a block in base58
// and expects the block's binary IRI, so a journal written before the change
// still names the node the graph holds.
func TestUnmarshalLegacyBlockIRI(t *testing.T) {
	// Name a block node and a bucket.
	h, err := hash.Sum(hash.HashType_HashType_SHA256, []byte("journal"))
	if err != nil {
		t.Fatal(err)
	}
	node := block_gc.BlockIRI(block.NewBlockRef(h))
	bucket := block_gc.BucketIRI("bucket")

	// Round trip an entry that spells the block in base58.
	entry := Marshal(
		[]block_gc.RefEdge{{Subject: bucket, Object: block_gc.FormatIRI(node)}},
		[]block_gc.RefEdge{{Subject: block_gc.FormatIRI(node), Object: bucket}},
	)
	adds, removes, err := Unmarshal(entry)
	if err != nil {
		t.Fatal(err)
	}

	// Expect the binary block IRI on both sides.
	wantAdds := []block_gc.RefEdge{{Subject: bucket, Object: node}}
	wantRemoves := []block_gc.RefEdge{{Subject: node, Object: bucket}}
	if !slices.Equal(adds, wantAdds) || !slices.Equal(removes, wantRemoves) {
		t.Fatalf("decoded adds %q removes %q, want %q and %q", adds, removes, wantAdds, wantRemoves)
	}
}
