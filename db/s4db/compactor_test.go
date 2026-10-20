//go:build darwin || linux || windows

package s4db

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"
)

// fillSlack writes two generations of keys, deletes the first and releases
// its space, so the second lies past the goal with free space below it. It
// returns the keys left.
func fillSlack(t *testing.T, ctx context.Context, db *DB) model {
	// Write or delete one generation of 512 keys in batches of 16.
	t.Helper()
	m := make(model)
	put := func(gen int, del bool) {
		for i := 0; i < 512; i += 16 {
			tx, err := db.NewTransaction(ctx, true)
			if err != nil {
				t.Fatal(err)
			}
			for j := i; j < i+16; j++ {
				key := fmt.Sprintf("g%d/%04d", gen, j)
				if del {
					err = tx.Delete(ctx, []byte(key))
					delete(m, key)
				} else {
					val := bytes.Repeat([]byte{byte(gen), byte(j)}, 2500)
					err = tx.Set(ctx, []byte(key), val)
					m[key] = val
				}
				if err != nil {
					tx.Discard()
					t.Fatal(err)
				}
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}
	put(0, false)
	put(1, false)
	put(0, true)

	// Checkpoint twice, so the deleted values' pages are released.
	w := db.w
	for range 2 {
		if err := w.lock(ctx); err != nil {
			t.Fatal(err)
		}
		err := w.checkpointDurable(ctx)
		if err == nil {
			err = w.releaseNow(ctx)
		}
		w.unlock(err)
		if err != nil {
			t.Fatal(err)
		}
	}
	return m
}

// TestCompactorKeepsChangedKey checks that a value changed or deleted
// between a compaction step's read and its commit is not overwritten, and
// that the others move.
func TestCompactorKeepsChangedKey(t *testing.T) {
	// Fill a file with slack, without the background compactor.
	ctx := context.Background()
	opts := Options{RelocateBudget: -1, SingleProcess: true}
	path := filepath.Join(t.TempDir(), "changed.s4wave")
	db := openTest(t, path, opts)
	defer db.Close()
	m := fillSlack(t, ctx, db)

	// Plan one step by hand.
	c := &compactor{db: db}
	ok, busy, err := c.plan(ctx)
	if err != nil || busy || !ok {
		t.Fatalf("plan: ok %v busy %v err %v", ok, busy, err)
	}

	// Scan, choose and read on one snapshot.
	st, stripe := db.acquire()
	refs, _, err := c.scan(ctx, st)
	var moves []move
	if err == nil {
		moves, busy, err = c.choose(ctx, refs)
	}
	if err == nil {
		err = c.read(moves)
	}
	db.release(st, stripe)
	if err != nil || busy {
		t.Fatalf("scan: busy %v err %v", busy, err)
	}

	// Pick the value moves.
	var vals []move
	for _, mv := range moves {
		if mv.val.isRef {
			vals = append(vals, mv)
		}
	}
	if len(vals) < 3 {
		t.Fatalf("%d values to move, want at least 3", len(vals))
	}

	// Change one key and delete another before the commit.
	tx, err := db.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	changed, deleted := string(vals[0].key), string(vals[1].key)
	m[changed] = []byte("changed")
	delete(m, deleted)
	if err := tx.Set(ctx, []byte(changed), m[changed]); err != nil {
		t.Fatal(err)
	}
	if err := tx.Delete(ctx, []byte(deleted)); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// Commit the step: the changed keys keep their new state, the others
	// move lower.
	held, moved, err := c.commit(ctx, moves)
	if err != nil || !held {
		t.Fatalf("commit: held %v err %v", held, err)
	}
	if moved == 0 {
		t.Fatal("no value moved")
	}
	cur, found, err := db.lookup(db.cur.Load(), vals[2].key)
	if err != nil || !found {
		t.Fatalf("lookup: found %v err %v", found, err)
	}
	if !cur.isRef || cur.ref.off >= vals[2].val.ref.off {
		t.Fatalf("value at %d, was at %d", cur.ref.off, vals[2].val.ref.off)
	}
	m.check(t, db)
	checkSpace(t, db)
}

// TestCompactorClose closes handles while the background compactor works
// through a file with slack, and checks that each reopen holds every key.
func TestCompactorClose(t *testing.T) {
	// Fill a file with slack, with the background compactor.
	ctx := context.Background()
	opts := Options{SingleProcess: true}
	path := filepath.Join(t.TempDir(), "close.s4wave")
	db := openTest(t, path, opts)
	m := fillSlack(t, ctx, db)

	// Close and reopen as the compactor works, checking every key.
	for range 8 {
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		db = openTest(t, path, opts)
		m.check(t, db)
		checkSpace(t, db)
	}

	// Close the last handle.
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestCompactorScanCurrentValue checks that a scan lists the value a key
// holds now, not the one in the checkpoint's index, and skips a key deleted
// since: the index's values may lie in space a commit freed and reused.
func TestCompactorScanCurrentValue(t *testing.T) {
	// Fill a file with slack, without the background compactor.
	ctx := context.Background()
	opts := Options{RelocateBudget: -1, SingleProcess: true}
	db := openTest(t, filepath.Join(t.TempDir(), "current.s4wave"), opts)
	defer db.Close()
	fillSlack(t, ctx, db)

	// Change one key and delete another after the checkpoint.
	changed, deleted := []byte("g1/0005"), []byte("g1/0006")
	tx, err := db.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Set(ctx, changed, bytes.Repeat([]byte{9}, 6000)); err != nil {
		t.Fatal(err)
	}
	if err := tx.Delete(ctx, deleted); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	// Scan every key on a snapshot whose overlay holds both changes.
	c := &compactor{db: db}
	if ok, _, err := c.plan(ctx); err != nil || !ok {
		t.Fatalf("plan: ok=%v err=%v", ok, err)
	}
	st, stripe := db.acquire()
	defer db.release(st, stripe)
	if !st.filter.has(changed) || !st.filter.has(deleted) {
		t.Fatal("a checkpoint took the changes into the index")
	}
	refs, _, err := c.scan(ctx, st)
	if err != nil {
		t.Fatal(err)
	}

	// The changed key lists its current value and the deleted key none.
	want, _, err := db.lookup(st, changed)
	if err != nil {
		t.Fatal(err)
	}
	var sawChanged bool
	for _, m := range refs {
		switch string(m.key) {
		case string(deleted):
			t.Fatal("scan listed a deleted key")
		case string(changed):
			sawChanged = true
			if !sameValue(m.val, want) {
				t.Fatalf("scan listed %+v, want the current value %+v", m.val.ref, want.ref)
			}
		}
	}
	if !sawChanged {
		t.Fatal("scan did not list the changed key")
	}
}
