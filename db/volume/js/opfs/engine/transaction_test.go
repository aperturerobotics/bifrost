//go:build !js

package engine

import (
	"errors"
	"testing"

	"github.com/s4wave/spacewave/db/kvtx"
)

// TestTransactionBlindWritesSerialize proves unobserved writes use commit order.
func TestTransactionBlindWritesSerialize(t *testing.T) {
	ctx := t.Context()
	engine, err := Open(ctx, newDiskBackend(t))
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	first, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct {
		tx    kvtx.Tx
		key   string
		value string
	}{
		{tx: first, key: "shared", value: "first"},
		{tx: first, key: "first-only", value: "one"},
		{tx: second, key: "shared", value: "second"},
		{tx: second, key: "second-only", value: "two"},
	} {
		if err := mutation.tx.Set(ctx, []byte(mutation.key), []byte(mutation.value)); err != nil {
			t.Fatal(err)
		}
	}
	if err := first.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := second.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	read, err := engine.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Discard()
	for key, want := range map[string]string{
		"shared":      "second",
		"first-only":  "one",
		"second-only": "two",
	} {
		value, found, err := read.Get(ctx, []byte(key))
		if err != nil || !found || string(value) != want {
			t.Errorf("get %q = %q, %t, %v; want %q", key, value, found, err, want)
		}
	}
}

// TestTransactionReadDependencyRejectsLogicalWrite proves revision validation.
func TestTransactionReadDependencyRejectsLogicalWrite(t *testing.T) {
	ctx := t.Context()
	engine, err := Open(ctx, newDiskBackend(t))
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	seed, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.Set(ctx, []byte("source"), []byte("initial")); err != nil {
		t.Fatal(err)
	}
	if err := seed.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	dependent, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	value, found, err := dependent.Get(ctx, []byte("source"))
	if err != nil || !found || string(value) != "initial" {
		t.Fatalf("dependent read = %q, %t, %v", value, found, err)
	}
	if err := dependent.Set(ctx, []byte("derived"), []byte("result")); err != nil {
		t.Fatal(err)
	}

	competing, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := competing.Set(ctx, []byte("other"), []byte("logical write")); err != nil {
		t.Fatal(err)
	}
	if err := competing.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := dependent.Commit(ctx); !errors.Is(err, kvtx.ErrInvalidSnapshot) {
		t.Fatalf("dependent commit error = %v, want %v", err, kvtx.ErrInvalidSnapshot)
	}

	read, err := engine.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Discard()
	if _, found, err := read.Get(ctx, []byte("derived")); err != nil || found {
		t.Fatalf("rejected write exists = %t, error = %v", found, err)
	}
	if value, found, err := read.Get(ctx, []byte("other")); err != nil || !found || string(value) != "logical write" {
		t.Fatalf("competing write = %q, %t, %v", value, found, err)
	}
}

// TestTransactionSurvivesPhysicalReclaim proves metadata commits use revision.
func TestTransactionSurvivesPhysicalReclaim(t *testing.T) {
	ctx := t.Context()
	engine, err := Open(ctx, newDiskBackend(t))
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()

	seed, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := seed.Set(ctx, []byte("metadata"), []byte("before reclaim")); err != nil {
		t.Fatal(err)
	}
	if err := seed.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	before, err := engine.loadRoot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if before.ReclaimNext > before.RetireThrough {
		t.Fatal("seed publication did not construct retired files")
	}

	tx, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	value, found, err := tx.Get(ctx, []byte("metadata"))
	if err != nil || !found || string(value) != "before reclaim" {
		t.Fatalf("metadata read = %q, %t, %v", value, found, err)
	}
	if err := tx.Set(ctx, []byte("metadata"), []byte("after reclaim")); err != nil {
		t.Fatal(err)
	}

	for {
		progress, err := engine.Reclaim(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if !progress {
			break
		}
	}
	after, err := engine.loadRoot(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if after.Generation <= before.Generation {
		t.Fatalf("generation = %d, want greater than %d", after.Generation, before.Generation)
	}
	if after.Revision != before.Revision {
		t.Fatalf("revision changed during reclaim: %d to %d", before.Revision, after.Revision)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit after reclaim: %v", err)
	}

	read, err := engine.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer read.Discard()
	value, found, err = read.Get(ctx, []byte("metadata"))
	if err != nil || !found || string(value) != "after reclaim" {
		t.Fatalf("metadata after commit = %q, %t, %v", value, found, err)
	}
}
