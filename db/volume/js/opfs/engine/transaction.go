package engine

import (
	"bytes"
	"context"
	"sort"

	"github.com/s4wave/spacewave/db/kvtx"
)

// transaction buffers bounded mutations and pins the first observed generation.
// Each file operation releases reclamation protection before application callbacks.
type transaction struct {
	// engine owns durable data and publication.
	engine *Engine
	// write permits pending mutations.
	write bool
	// metadata confines conflict validation to the public metadata namespace.
	metadata bool
	// discarded closes the transaction and all derived iterators.
	discarded bool
	// pinned records whether generation has been established.
	pinned bool
	// generation is the complete view observed by this transaction.
	generation uint64
	// pending provides read-your-writes within explicit memory bounds.
	pending map[string]*Record
	// pendingBytes charges retained mutation keys, values, and record overhead.
	pendingBytes int
}

// revision selects the logical state visible through this transaction's store.
func (t *transaction) revision(root *Root) uint64 {
	if t.metadata {
		return root.MetadataRevision
	}
	return root.Revision
}

// pin accepts the first observed generation and rejects all later mismatches.
func (t *transaction) pin(generation uint64) error {
	if !t.pinned {
		t.generation = generation
		t.pinned = true
	} else if t.generation != generation {
		return kvtx.ErrInvalidSnapshot
	}
	return nil
}

// check validates the transaction lifetime and caller cancellation.
func (t *transaction) check(ctx context.Context) error {
	if t.discarded {
		return kvtx.ErrDiscarded
	}
	return ctx.Err()
}

// Get returns a copied committed value or the transaction's pending value.
func (t *transaction) Get(ctx context.Context, key []byte) ([]byte, bool, error) {
	if err := t.check(ctx); err != nil {
		return nil, false, err
	}
	snapshot, err := t.engine.snapshot(ctx)
	if err != nil {
		return nil, false, err
	}
	defer snapshot.release()
	if err := t.pin(t.revision(snapshot.root)); err != nil {
		return nil, false, err
	}
	if pending := t.pending[string(key)]; pending != nil {
		return bytes.Clone(pending.Value), !pending.Deleted, nil
	}
	value, found, err := snapshot.get(ctx, key)
	return bytes.Clone(value), found, err
}

// Exists checks the same generation-consistent record view as Get.
func (t *transaction) Exists(ctx context.Context, key []byte) (bool, error) {
	_, found, err := t.Get(ctx, key)
	return found, err
}

// Set records a bounded caller-owned mutation for commit.
func (t *transaction) Set(ctx context.Context, key, value []byte) error {
	return t.mutate(ctx, key, value, false)
}

// Delete hides a key and remains a no-op when that key is absent.
func (t *transaction) Delete(ctx context.Context, key []byte) error {
	return t.mutate(ctx, key, nil, true)
}

// mutate retains bounded copied input; blind writes observe no committed state.
func (t *transaction) mutate(ctx context.Context, key, value []byte, deleted bool) error {
	if err := t.check(ctx); err != nil {
		return err
	}
	if !t.write {
		return kvtx.ErrNotWrite
	}
	if len(key) == 0 {
		return kvtx.ErrEmptyKey
	}
	if len(key) > maxKeyBytes || len(value) > MaxValueBytes {
		return ErrLimit
	}
	previous := t.pending[string(key)]
	count, size := len(t.pending)+1, t.pendingBytes+len(key)+len(value)+32
	if previous != nil {
		count--
		size -= len(previous.Key) + len(previous.Value) + 32
	}
	if count > maxBatchRecords || size > maxBatchBytes {
		return ErrLimit
	}
	t.pending[string(key)] = &Record{Key: bytes.Clone(key), Value: bytes.Clone(value), Deleted: deleted}
	t.pendingBytes = size
	return nil
}

// Size counts visible keys with a bounded cursor.
func (t *transaction) Size(ctx context.Context) (uint64, error) {
	var count uint64
	err := t.ScanPrefixKeys(ctx, nil, func([]byte) error {
		count++
		return nil
	})
	return count, err
}

// ScanPrefix visits ordered visible records without retaining the whole answer.
func (t *transaction) ScanPrefix(ctx context.Context, prefix []byte, cb func(key, value []byte) error) error {
	it := t.Iterate(ctx, prefix, true, false)
	defer it.Close()
	for it.Next() {
		value, err := it.Value()
		if err != nil {
			return err
		}
		if err := cb(it.Key(), value); err != nil {
			return err
		}
	}
	return it.Err()
}

// ScanPrefixKeys visits keys through the same bounded transaction cursor.
func (t *transaction) ScanPrefixKeys(ctx context.Context, prefix []byte, cb func(key []byte) error) error {
	return t.ScanPrefix(ctx, prefix, func(key, _ []byte) error { return cb(key) })
}

// Iterate merges bounded pending changes with a lazy committed partition cursor.
func (t *transaction) Iterate(ctx context.Context, prefix []byte, _ bool, reverse bool) kvtx.Iterator {
	pending := make([]*Record, 0, len(t.pending))
	for _, record := range t.pending {
		if bytes.HasPrefix(record.Key, prefix) {
			pending = append(pending, record)
		}
	}
	sort.Slice(pending, func(i, j int) bool { return bytes.Compare(pending[i].Key, pending[j].Key) < 0 })
	return &iterator{ctx: ctx, tx: t, prefix: bytes.Clone(prefix), reverse: reverse, pending: pending}
}

// Commit validates the complete observed generation and durably publishes writes.
func (t *transaction) Commit(ctx context.Context) error {
	if err := t.check(ctx); err != nil {
		return err
	}
	if !t.write || len(t.pending) == 0 {
		t.Discard()
		return nil
	}
	records := make([]*Record, 0, len(t.pending))
	for _, record := range t.pending {
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool { return bytes.Compare(records[i].Key, records[j].Key) < 0 })
	// Blind writes serialize against the current root without a stale read dependency.
	var base *uint64
	if t.pinned {
		base = &t.generation
	}
	err := t.engine.apply(ctx, base, records, t.metadata)
	t.Discard()
	return err
}

// Discard releases retained mutations and invalidates derived iterators.
func (t *transaction) Discard() {
	t.discarded = true
	t.pending = nil
	t.pendingBytes = 0
}

// _ verifies the transaction interface.
var _ kvtx.Tx = (*transaction)(nil)
