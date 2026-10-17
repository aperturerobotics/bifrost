package kvtx_block_iavl

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
)

// TestNodeSharedOnlyWithReaders checks that a read-only transaction takes a
// cached node without a clone, while a writer gets a private copy it may change
// in place.
func TestNodeSharedOnlyWithReaders(t *testing.T) {
	// Store a leaf node and attach an operation cache to the context.
	ctx := context.Background()
	store := block_mock.NewMockStore(0)
	ref, _, err := block.PutBlock(ctx, store, &Node{Size: 1, Key: []byte("k")})
	if err != nil {
		t.Fatal(err.Error())
	}
	opCtx, counter := block.WithReadCounter(ctx)
	opCtx = block.WithReadOperationStore(opCtx, store)

	// Decode the node through two read-only transactions.
	readNode := func() *Node {
		// Mark the transaction read-only before decoding through it.
		tx, cursor := block.NewTransaction(store, nil, ref, nil)
		tx.SetReadOnly()
		n, err := loadNode(opCtx, cursor)
		if err != nil {
			t.Fatal(err.Error())
		}
		return n
	}
	first, second := readNode(), readNode()
	if first != second {
		t.Fatal("read-only transactions received different node instances")
	}

	// Decode the node through a write transaction and change it in place.
	_, cursor := block.NewTransaction(store, nil, ref, nil)
	written, err := loadNode(opCtx, cursor)
	if err != nil {
		t.Fatal(err.Error())
	}
	if written == first {
		t.Fatal("write transaction received the shared node instance")
	}
	written.Key = []byte("changed")
	if got := string(readNode().GetKey()); got != "k" {
		t.Fatalf("shared node key = %q after a writer changed its copy", got)
	}

	// Verify one decode served every read and only the writer cloned.
	snapshot := counter.Snapshot()
	if snapshot.DecodedBlockUnmarshalCount != 1 ||
		snapshot.DecodedBlockCacheHitCount != 3 ||
		snapshot.DecodedBlockCloneCount != 1 {
		t.Fatalf("unexpected decoded cache counters: %+v", snapshot)
	}
}
