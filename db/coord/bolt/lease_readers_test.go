//go:build !js && !wasip1

package bolt

import (
	"testing"
	"time"

	"github.com/s4wave/spacewave/db/coord"
)

// TestLeaseRefreshRetainsReaders verifies that post-commit generation refresh
// does not remap storage beneath snapshots opened during the same lease.
func TestLeaseRefreshRetainsReaders(t *testing.T) {
	ctx := t.Context()
	db := openTestDB(t)
	c := NewCoordinator(db, nil)
	lease, err := c.WaitAcquireWriteLease(ctx, coord.Scope{VolumeID: "volume", ObjectStoreID: "objects"})
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Release(ctx)
	if _, err := lease.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	writeBoltValue(t, db, "published")
	reader, err := db.Begin(false)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Rollback()

	done := make(chan error, 1)
	go func() {
		snapshot, err := lease.Refresh(ctx)
		if err == nil && snapshot.Generation != 1 {
			t.Errorf("generation = %d, want 1", snapshot.Generation)
		}
		done <- err
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		_ = reader.Rollback()
		<-done
		t.Fatal("refresh waited for a reader opened within the held lease")
	}
}
