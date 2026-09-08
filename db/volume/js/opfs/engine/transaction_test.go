//go:build !js

package engine

import (
	"errors"
	"testing"
	"time"

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

// TestTransactionRetainsSnapshot proves reads remain stable and reclamation resumes.
func TestTransactionRetainsSnapshot(t *testing.T) {
	ctx := t.Context()
	disk := newDiskBackend(t)
	queued := make(chan struct{}, 1)
	backend := &queuedReclaimBackend{Backend: disk, queued: queued}
	engine, err := Open(ctx, backend)
	if err != nil {
		t.Fatal(err)
	}
	defer engine.Close()
	if err := engine.Apply(ctx, nil, []*Record{{Key: []byte("key"), Value: []byte("before")}}); err != nil {
		t.Fatal(err)
	}

	for _, commit := range []bool{false, true} {
		tx, err := engine.NewTransaction(ctx, false)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Discard()
		value, found, err := tx.Get(ctx, []byte("key"))
		if err != nil || !found {
			t.Fatalf("initial read: %q, %t, %v", value, found, err)
		}
		reads := disk.reads
		for range 100 {
			if _, _, err := tx.Get(ctx, []byte("key")); err != nil {
				t.Fatal(err)
			}
		}
		if disk.reads != reads {
			t.Fatalf("repeated snapshot reads performed %d extra file reads", disk.reads-reads)
		}

		reclaimed := make(chan error, 1)
		go func() {
			_, err := engine.Reclaim(ctx)
			reclaimed <- err
		}()
		select {
		case <-queued:
		case <-time.After(5 * time.Second):
			t.Fatal("reclamation did not queue")
		}
		if err := engine.Apply(ctx, nil, []*Record{{Key: []byte("key"), Value: []byte("after")}}); err != nil {
			t.Fatal(err)
		}
		if got, found, err := tx.Get(ctx, []byte("key")); err != nil || !found || string(got) != string(value) {
			t.Fatalf("snapshot changed during publication: %q, %t, %v", got, found, err)
		}
		select {
		case err := <-reclaimed:
			t.Fatalf("reclamation bypassed live snapshot: %v", err)
		default:
		}
		if commit {
			if err := tx.Commit(ctx); err != nil {
				t.Fatal(err)
			}
		} else {
			tx.Discard()
		}
		select {
		case err := <-reclaimed:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("reclamation did not resume after snapshot release")
		}
	}
}
