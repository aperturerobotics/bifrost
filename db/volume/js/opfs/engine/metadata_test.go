//go:build !js

package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/s4wave/spacewave/db/kvtx"
)

// TestMetadataTransactionIgnoresBlockAndGCChanges preserves the metadata view.
func TestMetadataTransactionIgnoresBlockAndGCChanges(t *testing.T) {
	ctx := t.Context()
	e, err := Open(ctx, newDiskBackend(t))
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	store := e.MetadataStore()
	writeMetadata := func(key string) {
		t.Helper()
		tx, err := store.NewTransaction(ctx, true)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Discard()
		if err := tx.Set(ctx, []byte(key), []byte("value")); err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
	}
	writeMetadata("saved")
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()
	if _, _, err := tx.Get(ctx, []byte("saved")); err != nil {
		t.Fatal(err)
	}

	// These logical writes do not change any public metadata key.
	blocks := &packStore{engine: e}
	if _, _, err := blocks.PutBlock(ctx, []byte("new block"), nil); err != nil {
		t.Fatal(err)
	}
	if err := e.Apply(ctx, nil, []*Record{{Key: []byte{0x02, 'g'}, Value: []byte("edge")}}); err != nil {
		t.Fatal(err)
	}
	if err := tx.ScanPrefix(ctx, nil, func(key, value []byte) error {
		if string(key) != "saved" || string(value) != "value" {
			t.Fatalf("metadata scan changed: %q=%q", key, value)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Set(ctx, []byte("after"), []byte("value")); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// A real metadata write must still invalidate the conservative read view.
	read, err := store.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Discard()
	if _, _, err := read.Get(ctx, []byte("saved")); err != nil {
		t.Fatal(err)
	}
	writeMetadata("other")
	if _, _, err := read.Get(context.Background(), []byte("saved")); !errors.Is(err, kvtx.ErrInvalidSnapshot) {
		t.Fatalf("changed metadata snapshot: %v", err)
	}
}
