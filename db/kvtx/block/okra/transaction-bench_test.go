package kvtx_block_okra

import (
	"strconv"
	"testing"

	"github.com/s4wave/spacewave/db/block"
)

// benchKeyCount is large enough for a multi-level page tree.
const benchKeyCount = 4096

// benchKey returns the key stored at index i.
func benchKey(i int) []byte {
	return []byte("key-" + strconv.Itoa(i))
}

// BenchmarkTransactionRead measures a read-only block transaction following
// the page references of a stored tree.
func BenchmarkTransactionRead(b *testing.B) {
	// Open a writable tree.
	ctx := b.Context()
	store := newOkraTestStore()
	btx, cursor := block.NewTransaction(store, nil, nil, nil)
	tx, err := NewTx(ctx, cursor, nil, true, nil)
	if err != nil {
		b.Fatal(err)
	}

	// Store a tree of many pages.
	for i := range benchKeyCount {
		if err := tx.Set(ctx, benchKey(i), benchKey(i)); err != nil {
			b.Fatal(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		b.Fatal(err)
	}
	ref, _, err := btx.Write(ctx, true)
	if err != nil {
		b.Fatal(err)
	}

	// Read every key through a fresh transaction.
	b.ReportAllocs()
	for b.Loop() {
		_, cursor := block.NewTransaction(store, nil, ref, nil)
		tx, err := NewTx(ctx, cursor, nil, false, nil)
		if err != nil {
			b.Fatal(err)
		}
		for i := range benchKeyCount {
			if _, _, err := tx.Get(ctx, benchKey(i)); err != nil {
				b.Fatal(err)
			}
		}
		tx.Discard()
	}
}

// BenchmarkTransactionWriteTree measures writing a transaction that dirties
// every page of a multi-level tree.
func BenchmarkTransactionWriteTree(b *testing.B) {
	ctx := b.Context()
	store := newOkraTestStore()
	b.ReportAllocs()
	for b.Loop() {
		btx, cursor := block.NewTransaction(store, nil, nil, nil)
		tx, err := NewTx(ctx, cursor, nil, true, nil)
		if err != nil {
			b.Fatal(err)
		}
		for i := range benchKeyCount {
			if err := tx.Set(ctx, benchKey(i), benchKey(i)); err != nil {
				b.Fatal(err)
			}
		}
		if err := tx.Commit(ctx); err != nil {
			b.Fatal(err)
		}
		if _, _, err := btx.Write(ctx, true); err != nil {
			b.Fatal(err)
		}
	}
}
