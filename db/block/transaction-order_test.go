package block

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
)

// TestWriteAtRootEncodesChildrenFirst verifies that a single encode worker
// reaches every child before the parent that waits on it.
func TestWriteAtRootEncodesChildrenFirst(t *testing.T) {
	// Serialize encoding so a parent scheduled first would hold the only worker.
	oldEncodeConcurrency := maxEncodeConcurrency
	maxEncodeConcurrency = 1
	defer func() { maxEncodeConcurrency = oldEncodeConcurrency }()

	// Chain several dirty blocks below the root.
	tx, cursor := NewTransaction(NopStoreOps{}, nil, nil, nil)
	cursor.SetBlock(&transactionWorkerBlock{}, true)
	for range 4 {
		cursor = cursor.FollowRef(1, nil)
		cursor.SetBlock(&transactionWorkerBlock{}, true)
	}

	// The write fails by timeout when a parent waits on an unscheduled child.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if _, _, err := tx.Write(ctx, true); err != nil {
		t.Fatal(err)
	}
}

// TestWriteAtRootRejectsReferenceCycle verifies that a block reachable from
// itself fails the write instead of waiting on its own encode.
func TestWriteAtRootRejectsReferenceCycle(t *testing.T) {
	// Point a child back at the root.
	tx, root := NewTransaction(NopStoreOps{}, nil, nil, nil)
	root.SetBlock(&transactionWorkerBlock{}, true)
	child := root.FollowRef(1, nil)
	child.SetBlock(&transactionWorkerBlock{}, true)
	child.SetRef(1, root)

	// The write reports the cycle.
	if _, _, err := tx.Write(t.Context(), true); !errors.Is(err, ErrReferenceCycle) {
		t.Fatalf("write error=%v want=%v", err, ErrReferenceCycle)
	}
}
