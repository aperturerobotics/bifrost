//go:build darwin || linux || windows

package s4db

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// TestCloseWaitsForCommit checks that Close waits for an open write
// transaction to commit durably, refuses new ones meanwhile, and leaves the
// commit in the file.
func TestCloseWaitsForCommit(t *testing.T) {
	// Open a write transaction and start closing.
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "drain.s4wave")
	db := openTest(t, path, Options{})
	tx, err := db.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Set(ctx, []byte("k"), []byte("v")); err != nil {
		t.Fatal(err)
	}
	closed := make(chan error, 1)
	go func() { closed <- db.Close() }()

	// Once closing, Close refuses a new writer and waits for the open one.
	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = db.bcast.Wait(wctx, func(_ func(), _ func() <-chan struct{}) (bool, error) {
		return db.stopping, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.NewTransaction(ctx, true); !errors.Is(err, ErrClosed) {
		t.Fatalf("NewTransaction while closing = %v, want ErrClosed", err)
	}
	select {
	case err := <-closed:
		t.Fatalf("Close returned %v before the commit", err)
	case <-time.After(50 * time.Millisecond):
	}

	// The commit finishes, then Close.
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-closed; err != nil {
		t.Fatal(err)
	}

	// The reopened file holds the commit.
	db = openTest(t, path, Options{})
	defer db.Close()
	rtx, err := db.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer rtx.Discard()
	if v, ok, err := rtx.Get(ctx, []byte("k")); err != nil || !ok || string(v) != "v" {
		t.Fatalf("get k = %q %v %v", v, ok, err)
	}
}

// TestCloseDuringCommits closes while writers commit in a loop. Under the
// race detector it checks that no commit flushes the file as it closes.
func TestCloseDuringCommits(t *testing.T) {
	// Start writers committing until the handle refuses them.
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "race.s4wave")
	db := openTest(t, path, Options{})
	var wg sync.WaitGroup
	committed := make([]int, 4)
	for w := range committed {
		wg.Go(func() {
			for i := 0; ; i++ {
				tx, err := db.NewTransaction(ctx, true)
				if err != nil {
					return
				}
				_ = tx.Set(ctx, fmt.Appendf(nil, "w%d/%06d", w, i), []byte("v"))
				if err := tx.Commit(ctx); err != nil {
					t.Errorf("commit: %v", err)
					return
				}
				committed[w] = i + 1
			}
		})
	}

	// Close while they run, then reopen.
	time.Sleep(50 * time.Millisecond)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	db = openTest(t, path, Options{})
	defer db.Close()

	// Every commit that returned is in the file.
	rtx, err := db.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer rtx.Discard()
	for w, n := range committed {
		if n == 0 {
			continue
		}
		key := fmt.Appendf(nil, "w%d/%06d", w, n-1)
		if _, ok, err := rtx.Get(ctx, key); err != nil || !ok {
			t.Fatalf("last commit %s of writer %d missing: %v", key, w, err)
		}
	}
}
