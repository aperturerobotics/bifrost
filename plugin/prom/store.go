package prom_plugin

import (
	"context"
	"math"
	"sync"

	"github.com/pkg/errors"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/storage"
	"github.com/s4wave/spacewave/db/block"
	block_kvtx "github.com/s4wave/spacewave/db/kvtx/block"
	"github.com/s4wave/spacewave/db/world"
)

// Store is the single writer of one target's samples.
//
// The target is a World object whose root is a block-backed key-value tree.
// The Store keeps the open window and the target's series table in memory;
// Append adds to the window and CommitStores writes the window as one segment.
// A crash or Rollback loses only the uncommitted window. Exactly one Store
// writes a target, so a Store never reads the tree again after it opens.
type Store struct {
	// key is the World object key of the target.
	key string
	// meta holds the labels that every series of the target carries.
	meta labels.Labels

	// mtx guards the fields below.
	mtx sync.Mutex
	// table maps series label sets to series numbers, including series that
	// only the uncommitted windows introduced.
	table *SeriesTable
	// open is the window Append adds to.
	open *Window
	// frozen is the encoded window a commit is writing, or retrying after a
	// failed commit; nil when none is pending.
	frozen *commitBatch
	// inFlight is set while a commit of the frozen batch is running.
	inFlight bool
	// committedSeries is the number of series the committed pages hold.
	committedSeries uint64
	// committedMax is the latest sample time committed, valid when hasCommitted.
	committedMax int64
	// hasCommitted is set once the target holds a segment.
	hasCommitted bool
}

// CreateStore creates the World object key as an empty target carrying the
// labels target on every series, and returns its writer.
func CreateStore(ctx context.Context, eng world.Engine, key string, target labels.Labels) (*Store, error) {
	// Encode the target's labels once; the transaction may replay.
	metaData, err := (&TargetMeta{Labels: NewLabelSet(target)}).MarshalVT()
	if err != nil {
		return nil, err
	}

	// Create the object with an empty tree holding the target meta.
	err = world.ExecTransaction(ctx, eng, true, func(ctx context.Context, wtx world.WorldState) error {
		obj, _, err := world.CreateWorldObject(ctx, wtx, key, func(bcs *block.Cursor) error {
			// Start an empty tree for the target.
			bcs.SetBlock(block_kvtx.NewKeyValueStore(0), true)
			kv, err := block_kvtx.BuildKvTransaction(ctx, bcs, true)
			if err != nil {
				return err
			}
			defer kv.Discard()

			// Store the target's labels.
			if err := kv.Set(ctx, metaKey(), metaData); err != nil {
				return err
			}
			return kv.Commit(ctx)
		})
		world.ReleaseObjectState(obj)
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "create target %q", key)
	}
	return newStore(key, target.Copy(), NewSeriesTable(), 0, false, 0), nil
}

// OpenStore opens the existing target key for writing, reading its labels,
// series table and latest committed time.
func OpenStore(ctx context.Context, eng world.Engine, key string) (*Store, error) {
	// Read the committed state of the target.
	var st *Store
	err := world.ExecTransaction(ctx, eng, false, func(ctx context.Context, wtx world.WorldState) error {
		// Access the target's tree.
		obj, err := world.MustGetObject(ctx, wtx, key)
		if err != nil {
			return err
		}
		defer world.ReleaseObjectState(obj)
		_, _, err = world.AccessObjectState(ctx, obj, false, func(bcs *block.Cursor) error {
			// Open the tree for reading.
			kv, err := block_kvtx.BuildKvTransaction(ctx, bcs, false)
			if err != nil {
				return err
			}
			defer kv.Discard()

			// Read the labels, the series table and the latest committed time.
			r := NewStoreReader(kv)
			meta, err := r.Meta(ctx)
			if err != nil {
				return err
			}
			table, err := r.Table(ctx)
			if err != nil {
				return err
			}
			maxTime, has, err := r.LastSegmentMaxTime(ctx)
			if err != nil {
				return err
			}

			// Resume the writer after the committed state.
			st = newStore(key, meta, table, table.Len(), has, maxTime)
			return nil
		})
		return err
	})
	if err != nil {
		return nil, errors.Wrapf(err, "open target %q", key)
	}
	return st, nil
}

// newStore constructs a store over already committed state.
func newStore(key string, meta labels.Labels, table *SeriesTable, committedSeries uint64, hasCommitted bool, committedMax int64) *Store {
	return &Store{
		key:             key,
		meta:            meta,
		table:           table,
		open:            NewWindow(table.Len()),
		committedSeries: committedSeries,
		committedMax:    committedMax,
		hasCommitted:    hasCommitted,
	}
}

// Key returns the World object key of the target.
func (s *Store) Key() string {
	return s.key
}

// Append adds a sample of the series lset to the open window.
//
// lset must not carry a name that the target's labels set. The timestamp must
// be later than every sample the target has committed or is committing, and
// later than the series' earlier samples in the open window; otherwise Append
// returns storage.ErrOutOfOrderSample, or
// storage.ErrDuplicateSampleForTimestamp for the same time with another value.
func (s *Store) Append(lset labels.Labels, t int64, v float64) error {
	// Reject labels that the target already provides.
	var reserved error
	lset.Range(func(l labels.Label) {
		if s.meta.Has(l.Name) {
			reserved = errors.Wrapf(ErrReservedLabel, "%q", l.Name)
		}
	})
	if reserved != nil {
		return reserved
	}

	// Keep each window after the one before it, so segments never overlap.
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if floor, ok := s.floorLocked(); ok && t <= floor {
		return storage.ErrOutOfOrderSample
	}

	// Number the series if the window is the first to see it, then append.
	num, ok := s.table.Lookup(lset)
	if !ok {
		num = s.table.Add(lset.Copy())
	}
	return s.open.Append(num, t, v)
}

// AppendStale adds a stale marker for the series lset at time t, which tells
// PromQL the series ended, as Prometheus does for a failed or vanished scrape.
func (s *Store) AppendStale(lset labels.Labels, t int64) error {
	return s.Append(lset, t, math.Float64frombits(value.StaleNaN))
}

// Rollback discards every uncommitted sample and the series only they
// introduced. It returns ErrCommitInProgress while a commit is running.
func (s *Store) Rollback() error {
	// Refuse while a commit is running.
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.inFlight {
		return ErrCommitInProgress
	}

	// Forget the frozen batch, the open window and the series they introduced.
	s.frozen = nil
	s.table.Truncate(s.committedSeries)
	s.open = NewWindow(s.committedSeries)
	return nil
}

// floorLocked returns the latest sample time that is committed or frozen.
// Callers hold mtx.
func (s *Store) floorLocked() (int64, bool) {
	floor, ok := s.committedMax, s.hasCommitted
	if s.frozen != nil && (!ok || s.frozen.maxTime > floor) {
		floor, ok = s.frozen.maxTime, true
	}
	return floor, ok
}

// freeze returns the batch a commit should write: the batch left by a failed
// commit, or else the open window encoded as a batch. It returns nil when
// there is nothing to write. A returned batch is in flight until finish.
func (s *Store) freeze() (*commitBatch, error) {
	// Refuse while a commit is running.
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.inFlight {
		return nil, ErrCommitInProgress
	}

	// Encode the open window unless a failed commit left a batch to retry.
	if s.frozen == nil {
		batch, err := newCommitBatch(s.open, s.table)
		if err != nil || batch == nil {
			return nil, err
		}
		s.frozen = batch
		s.open = NewWindow(s.table.Len())
	}

	// Mark the commit running and hand out the batch.
	s.inFlight = true
	return s.frozen, nil
}

// finish ends the in-flight commit of batch. A written batch becomes
// committed state; a failed one stays frozen for the next commit.
func (s *Store) finish(batch *commitBatch, written bool) {
	// End the commit; a failed one leaves the frozen batch for the next try.
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.inFlight = false
	if !written {
		return
	}

	// Move the committed state past the written batch.
	s.frozen = nil
	s.committedSeries = batch.endSeries
	s.committedMax, s.hasCommitted = batch.maxTime, true
}

// write stores batch in the target's tree within wtx. Writing a batch again
// stores the same keys with the same values, so a replayed or retried commit
// is harmless.
func (s *Store) write(ctx context.Context, wtx world.WorldState, batch *commitBatch) error {
	// Open the target; it must exist because a Store is created or opened on it.
	obj, err := world.MustGetObject(ctx, wtx, s.key)
	if err != nil {
		return errors.Wrapf(err, "target %q", s.key)
	}
	defer world.ReleaseObjectState(obj)

	// Set the batch's keys in the target's tree and move the object root.
	_, _, err = world.AccessObjectState(ctx, obj, true, func(bcs *block.Cursor) error {
		// Open the tree for writing.
		kv, err := block_kvtx.BuildKvTransaction(ctx, bcs, true)
		if err != nil {
			return err
		}
		defer kv.Discard()

		// Set every entry, then commit the tree to move the object root.
		for _, e := range batch.entries {
			if err := kv.Set(ctx, e.key, e.value); err != nil {
				return err
			}
		}
		return kv.Commit(ctx)
	})
	return err
}
