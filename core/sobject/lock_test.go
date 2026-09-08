package sobject

import (
	"context"
	"testing"
)

// TestSOStateLockRelease fences writes after releasing the provider's lock.
func TestSOStateLockRelease(t *testing.T) {
	// Forward state and lineage while the handle owns the provider lock.
	state := &SOState{}
	entry := &SOConfigChange{}
	writes, releases := 0, 0
	lock := NewSOStateLock(state, func(_ context.Context, got *SOState, changes ...*SOConfigChange) error {
		writes++
		if got != state || len(changes) != 1 || changes[0] != entry {
			t.Fatal("write did not forward the state and lineage together")
		}
		return nil
	}, func() { releases++ })
	if err := lock.WriteSOState(t.Context(), state, entry); err != nil {
		t.Fatal(err)
	}

	// Repeated cleanup cannot release twice or permit an unlocked write.
	lock.Release()
	lock.Release()
	if err := lock.WriteSOState(t.Context(), state, entry); err == nil {
		t.Fatal("accepted a write after releasing the provider lock")
	}
	if writes != 1 || releases != 1 {
		t.Fatalf("writes = %d, releases = %d", writes, releases)
	}
}
