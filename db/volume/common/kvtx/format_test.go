//go:build darwin || linux || windows

package kvtx

import (
	"bytes"
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/aperturerobotics/cayley/graph"
	"github.com/aperturerobotics/cayley/quad"
	b58 "github.com/mr-tron/base58/base58"
	"github.com/s4wave/spacewave/db/block"
	block_gc "github.com/s4wave/spacewave/db/block/gc"
	db_kvtx "github.com/s4wave/spacewave/db/kvtx"
	kvtx_cayley "github.com/s4wave/spacewave/db/kvtx/cayley"
	kvtx_prefixer "github.com/s4wave/spacewave/db/kvtx/prefixer"
	"github.com/s4wave/spacewave/db/s4db"
	store_kvkey "github.com/s4wave/spacewave/db/store/kvkey"
	store_kvtx_inmem "github.com/s4wave/spacewave/db/store/kvtx/inmem"
)

// TestUpgradeBase58BlockKeys opens a store whose blocks are keyed in base58,
// as version 0 wrote them, with one block already moved by an interrupted
// migration, and expects every block under its binary key.
func TestUpgradeBase58BlockKeys(t *testing.T) {
	t.Run("inmem", func(t *testing.T) {
		testUpgradeBase58BlockKeys(t, store_kvtx_inmem.NewStore())
	})
	t.Run("s4db", func(t *testing.T) {
		db, err := s4db.Open(filepath.Join(t.TempDir(), "volume.s4db"), s4db.Options{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		testUpgradeBase58BlockKeys(t, db)
	})
}

// testUpgradeBase58BlockKeys runs TestUpgradeBase58BlockKeys on one store.
func testUpgradeBase58BlockKeys(t *testing.T, store db_kvtx.Store) {
	// Prepare the blocks of the unversioned store.
	ctx := context.Background()
	keys := store_kvkey.NewDefaultKVKey()
	blocks := [][]byte{[]byte("alpha"), []byte("beta"), []byte("gamma")}
	refs := make([]*block.BlockRef, len(blocks))

	// Write version 0 blocks, leaving the last one moved.
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	for i, data := range blocks {
		refs[i], err = block.BuildBlockRef(data, nil)
		if err != nil {
			t.Fatal(err)
		}
		rm, err := refs[i].MarshalKey()
		if err != nil {
			t.Fatal(err)
		}
		key := keys.GetBlockKey(rm)
		if i != len(blocks)-1 {
			key = append(keys.GetBlockFullPrefix(), b58.Encode(rm)...)
		}
		if err := tx.Set(ctx, key, data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// Open the Volume, which migrates the store.
	vol, err := NewVolume(ctx, "hydra/test-volume", keys, store, nil, false, false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = vol.Close() })

	// Every block reads back.
	for i, ref := range refs {
		data, found, err := vol.GetBlock(ctx, ref)
		if err != nil || !found || !bytes.Equal(data, blocks[i]) {
			t.Fatalf("block %d: found %v data %q err %v", i, found, data, err)
		}
	}

	// Only binary block keys remain.
	rtx, err := store.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer rtx.Discard()
	var count int
	err = rtx.ScanPrefixKeys(ctx, keys.GetBlockFullPrefix(), func(key []byte) error {
		count++
		if key[len(keys.GetBlockFullPrefix())] != binaryBlockRefTag {
			t.Errorf("block key %q is not binary", key)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != len(blocks) {
		t.Fatalf("store holds %d block keys, want %d", count, len(blocks))
	}

	// The store records the current version.
	version, err := readFormatVersion(ctx, store, keys)
	if err != nil {
		t.Fatal(err)
	}
	if version != store_kvkey.FormatVersion {
		t.Fatalf("format version %d, want %d", version, store_kvkey.FormatVersion)
	}
}

// TestUpgradeCayleyRefGraph opens a version 1 store whose reference graph is
// a Cayley quad store and expects the live reference edges as edge records
// with the Cayley records deleted.
func TestUpgradeCayleyRefGraph(t *testing.T) {
	t.Run("inmem", func(t *testing.T) {
		testUpgradeCayleyRefGraph(t, store_kvtx_inmem.NewStore())
	})
	t.Run("s4db", func(t *testing.T) {
		db, err := s4db.Open(filepath.Join(t.TempDir(), "volume.s4db"), s4db.Options{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		testUpgradeCayleyRefGraph(t, db)
	})
}

// testUpgradeCayleyRefGraph runs TestUpgradeCayleyRefGraph on one store.
func testUpgradeCayleyRefGraph(t *testing.T, store db_kvtx.Store) {
	// Open a Cayley graph in the store.
	ctx := context.Background()
	keys := store_kvkey.NewDefaultKVKey()
	h, err := kvtx_cayley.NewGraph(ctx, kvtx_prefixer.NewPrefixer(store, cayleyGraphPrefix), graph.Options{
		"ignore_duplicate": true,
		"ignore_missing":   true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Add the edges, including one of another predicate, and remove one.
	ref := func(subject, object string) quad.Quad {
		return quad.MakeIRI(subject, block_gc.PredGCRef, object, "")
	}
	for _, q := range []quad.Quad{
		ref(block_gc.NodeGCRoot, "a"),
		ref("a", "b"),
		ref("a", "c"),
		quad.MakeIRI("a", "other", "d", ""),
	} {
		if err := h.AddQuad(ctx, q); err != nil {
			t.Fatal(err)
		}
	}
	if err := h.RemoveQuad(ctx, ref("a", "c")); err != nil {
		t.Fatal(err)
	}
	if err := h.Close(); err != nil {
		t.Fatal(err)
	}

	// Record format version 1.
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Set(ctx, keys.GetFormatVersionKey(), store_kvkey.MarshalFormatVersion(1)); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// Migrate the store.
	if err := upgradeFormat(ctx, store, keys); err != nil {
		t.Fatal(err)
	}

	// The live reference edges moved.
	rg := openRefGraph(store)
	for subject, want := range map[string][]string{
		block_gc.NodeGCRoot: {"a"},
		"a":                 {"b"},
	} {
		got, err := rg.GetOutgoingRefs(ctx, subject)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("outgoing refs of %q: %v, want %v", subject, got, want)
		}
	}

	// No Cayley records remain.
	rtx, err := store.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer rtx.Discard()
	err = rtx.ScanPrefixKeys(ctx, cayleyGraphPrefix, func(key []byte) error {
		t.Errorf("Cayley record %q remains", key)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
