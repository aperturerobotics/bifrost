//go:build darwin || linux || windows

package kvtx

import (
	"bytes"
	"context"
	"encoding/binary"
	"path/filepath"
	"slices"
	"strconv"
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
	"github.com/s4wave/spacewave/net/hash"
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
	// Migrate a version 1 store and check its graph.
	ctx := context.Background()
	keys := store_kvkey.NewDefaultKVKey()
	writeCayleyStore(t, store, keys)
	if err := upgradeFormat(ctx, store, keys); err != nil {
		t.Fatal(err)
	}
	checkConvertedRefGraph(t, store)
}

// TestUpgradeLegacyRefGraph opens a version 2 store and expects its edges in
// the current encoding with the version 2 records deleted.
func TestUpgradeLegacyRefGraph(t *testing.T) {
	// Migrate a version 2 store and check its graph.
	ctx := context.Background()
	store := store_kvtx_inmem.NewStore()
	keys := store_kvkey.NewDefaultKVKey()
	writeLegacyStore(t, store, keys)
	if err := upgradeFormat(ctx, store, keys); err != nil {
		t.Fatal(err)
	}
	checkLegacyRefGraph(t, store)
}

// TestMigrateLegacyRefGraphBatches moves more records than one batch holds,
// resuming each batch after the last, and expects every record moved.
func TestMigrateLegacyRefGraphBatches(t *testing.T) {
	// Write a version 2 root record per node across three batches.
	ctx := context.Background()
	store := store_kvtx_inmem.NewStore()
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()
	n := 2*convertBatchEdges + 3
	for i := range n {
		if err := tx.Set(ctx, legacyGraphKey('r', "node-"+strconv.Itoa(i)), nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// Migrate and count the records under each prefix.
	if err := MigrateLegacyRefGraph(ctx, store, legacyRefGraphPrefix, refGraphPrefix); err != nil {
		t.Fatal(err)
	}
	if got := countPrefix(t, store, legacyRefGraphPrefix); got != 0 {
		t.Fatalf("%d version 2 records left", got)
	}
	if got := countPrefix(t, store, refGraphPrefix); got != n {
		t.Fatalf("moved %d records, want %d", got, n)
	}
}

// countPrefix returns the number of keys under prefix.
func countPrefix(t *testing.T, store db_kvtx.Store, prefix []byte) int {
	// Scan the prefix in a snapshot.
	t.Helper()
	tx, err := store.NewTransaction(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()
	var n int
	err = tx.ScanPrefixKeys(context.Background(), prefix, func([]byte) error {
		n++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// copyCase is an earlier format store CopyStore converts.
type copyCase struct {
	// version is the format version write records.
	version uint64
	// write fills the source store.
	write func(*testing.T, db_kvtx.Store, *store_kvkey.KVKey)
	// check checks the converted graph of the destination.
	check func(*testing.T, db_kvtx.Store)
	// keys and edges are the final progress counts.
	keys, edges uint64
}

// copyCases are the earlier formats CopyStore converts, by name.
var copyCases = map[string]copyCase{
	"cayley": {version: 1, write: writeCayleyStore, check: checkConvertedRefGraph, keys: 1, edges: 3},
	"legacy": {version: 2, write: writeLegacyStore, check: checkLegacyRefGraph, keys: 8},
}

// TestCopyStore copies a store of each earlier format and expects its other
// keys verbatim, its live reference edges in the current encoding, no earlier
// graph records and the current format version, with the source unchanged.
func TestCopyStore(t *testing.T) {
	for name, c := range copyCases {
		t.Run(name+"/inmem", func(t *testing.T) {
			testCopyStore(t, c, store_kvtx_inmem.NewStore(), store_kvtx_inmem.NewStore())
		})
		t.Run(name+"/s4db", func(t *testing.T) {
			dir := t.TempDir()
			var dbs [2]*s4db.DB
			for i, name := range []string{"src.s4db", "dst.s4db"} {
				db, err := s4db.Open(filepath.Join(dir, name), s4db.Options{})
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = db.Close() })
				dbs[i] = db
			}
			testCopyStore(t, c, dbs[0], dbs[1])
		})
	}
}

// testCopyStore runs TestCopyStore on one pair of stores.
func testCopyStore(t *testing.T, c copyCase, src, dst db_kvtx.Store) {
	// Prepare an earlier format store with a key outside the graph.
	ctx := context.Background()
	keys := store_kvkey.NewDefaultKVKey()
	c.write(t, src, keys)
	tx, err := src.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Set(ctx, []byte("other"), []byte("value")); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// Copy it.
	var last CopyProgress
	if err := CopyStore(ctx, src, dst, keys, func(p CopyProgress) { last = p }); err != nil {
		t.Fatal(err)
	}
	if last.Keys != c.keys || last.Edges != c.edges {
		t.Fatalf("progress %+v, want %d keys and %d edges", last, c.keys, c.edges)
	}

	// The destination holds the converted graph, the other key and the
	// current version, and the source keeps its version.
	c.check(t, dst)
	rtx, err := dst.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer rtx.Discard()
	value, found, err := rtx.Get(ctx, []byte("other"))
	if err != nil || !found || string(value) != "value" {
		t.Fatalf("other key: found %v value %q err %v", found, value, err)
	}
	for store, want := range map[db_kvtx.Store]uint64{src: c.version, dst: store_kvkey.FormatVersion} {
		version, err := readFormatVersion(ctx, store, keys)
		if err != nil {
			t.Fatal(err)
		}
		if version != want {
			t.Fatalf("format version %d, want %d", version, want)
		}
	}

	// A second copy refuses the filled destination.
	if err := CopyStore(ctx, src, dst, keys, nil); err == nil {
		t.Fatal("copy into a non-empty store succeeded")
	}
}

// writeCayleyStore writes a version 1 store whose Cayley reference graph
// holds root->a and a->b, a removed a->c, and an edge of another predicate.
func writeCayleyStore(t *testing.T, store db_kvtx.Store, keys *store_kvkey.KVKey) {
	// Open a Cayley graph in the store.
	ctx := context.Background()
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
}

// checkConvertedRefGraph expects the edge records of the writeCayleyStore
// graph and no Cayley records.
func checkConvertedRefGraph(t *testing.T, store db_kvtx.Store) {
	// The live reference edges moved.
	ctx := context.Background()
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

// writeLegacyStore writes a version 2 store whose graph holds gcroot->bucket,
// bucket->a and a->b, with a and b block nodes, and registers gcroot as a
// root.
func writeLegacyStore(t *testing.T, store db_kvtx.Store, keys *store_kvkey.KVKey) {
	// Collect the root record and both records of each edge.
	records := [][]byte{legacyGraphKey('r', block_gc.NodeGCRoot)}
	for _, edge := range legacyEdges(t) {
		records = append(records,
			legacyGraphKey('f', edge.Subject, edge.Object),
			legacyGraphKey('i', edge.Object, edge.Subject),
		)
	}

	// Write them with the format version.
	ctx := context.Background()
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()
	for _, key := range records {
		if err := tx.Set(ctx, key, nil); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Set(ctx, keys.GetFormatVersionKey(), store_kvkey.MarshalFormatVersion(2)); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

// legacyGraphKey encodes a version 2 record key.
func legacyGraphKey(kind byte, nodes ...string) []byte {
	key := append(bytes.Clone(legacyRefGraphPrefix), kind)
	for _, node := range nodes {
		key = binary.BigEndian.AppendUint32(key, uint32(len(node))) //nolint:gosec // test IRIs are short.
		key = append(key, node...)
	}
	return key
}

// legacyEdges returns the edges writeLegacyStore writes.
func legacyEdges(t *testing.T) []block_gc.RefEdge {
	t.Helper()
	bucket := block_gc.BucketIRI("bucket")
	a, b := testBlockIRI(t, "a"), testBlockIRI(t, "b")
	return []block_gc.RefEdge{
		{Subject: block_gc.NodeGCRoot, Object: bucket},
		{Subject: bucket, Object: a},
		{Subject: a, Object: b},
	}
}

// testBlockIRI returns the IRI of a block holding data.
func testBlockIRI(t *testing.T, data string) string {
	t.Helper()
	h, err := hash.Sum(hash.HashType_HashType_SHA256, []byte(data))
	if err != nil {
		t.Fatal(err)
	}
	return block_gc.BlockIRI(block.NewBlockRef(h))
}

// checkLegacyRefGraph expects the writeLegacyStore edges in both directions,
// its root and no version 2 records.
func checkLegacyRefGraph(t *testing.T, store db_kvtx.Store) {
	// Each edge reads forward and back.
	ctx := context.Background()
	rg := openRefGraph(store)
	for _, edge := range legacyEdges(t) {
		out, err := rg.GetOutgoingRefs(ctx, edge.Subject)
		if err != nil {
			t.Fatal(err)
		}
		in, err := rg.GetIncomingRefs(ctx, edge.Object)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(out, []string{edge.Object}) || !slices.Equal(in, []string{edge.Subject}) {
			t.Fatalf("edge %v: outgoing %v, incoming %v", edge, out, in)
		}
	}

	// The root survives.
	roots, err := block_gc.NewCollectorRefGraph(kvtx_prefixer.NewPrefixer(store, refGraphPrefix)).GetRootNodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(roots, []string{block_gc.NodeGCRoot}) {
		t.Fatalf("roots %v", roots)
	}

	// No version 2 records remain.
	rtx, err := store.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer rtx.Discard()
	err = rtx.ScanPrefixKeys(ctx, legacyRefGraphPrefix, func(key []byte) error {
		t.Errorf("version 2 record %q remains", key)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
