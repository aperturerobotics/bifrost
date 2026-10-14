package resource_world

import (
	"context"
	"maps"
	"slices"

	"github.com/aperturerobotics/util/broadcast"
	"github.com/s4wave/spacewave/db/bucket"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/net/peer"
)

// TrackedWorldState wraps a WorldState and records the reads made through it.
// One watch compares the recorded reads with the live World until one of them
// changes, so a view built from the reads knows when to rebuild.
type TrackedWorldState struct {
	// ws is the read transaction the view reads.
	ws world.WorldState
	// watchWs is the live World the watch compares with.
	watchWs world.WorldState
	// initialSeqno is the seqno of ws.
	initialSeqno uint64
	// cancel stops the watch.
	cancel context.CancelFunc
	// changeResultCh receives the watch result: nil when a read changed.
	changeResultCh chan error

	// bcast guards the fields below and fires when a read is recorded.
	bcast broadcast.Broadcast
	// revs maps each read object key to its revision in ws, 0 when missing.
	revs map[string]uint64
	// pending are the keys read since the watch last checked.
	pending []string
	// coarse is set by reads that object revisions cannot describe: graph
	// reads and object iteration. Any commit after initialSeqno changes them.
	coarse bool
}

// NewTrackedWorldState creates a TrackedWorldState reading ws at initialSeqno
// and starts watching watchWs for changes to its reads.
func NewTrackedWorldState(ws world.WorldState, watchWs world.WorldState, initialSeqno uint64, ctx context.Context) *TrackedWorldState {
	// Construct the tracked state with no reads.
	ctx, cancel := context.WithCancel(ctx)
	t := &TrackedWorldState{
		ws:             ws,
		watchWs:        watchWs,
		initialSeqno:   initialSeqno,
		cancel:         cancel,
		changeResultCh: make(chan error, 1),
		revs:           make(map[string]uint64),
	}

	// Publish the watch result unless the watch was stopped.
	go func() {
		err := t.watch(ctx)
		if ctx.Err() == nil {
			t.changeResultCh <- err
		}
	}()
	return t
}

// WaitForChanges waits until any tracked resource changes.
// Returns nil when changes detected, error on failure or context cancel.
func (t *TrackedWorldState) WaitForChanges(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-t.changeResultCh:
		return err
	}
}

// trackObjectAccess records an object read at its revision, 0 when missing.
func (t *TrackedWorldState) trackObjectAccess(key string, rev uint64) {
	t.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		if t.trackRevLocked(key, rev) {
			broadcast()
		}
	})
}

// trackObjectBodyAccesses records the reads from one batched body read with a
// single wake of the watch.
func (t *TrackedWorldState) trackObjectBodyAccesses(bodies []*world.ObjectBody) {
	t.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		var added bool
		for _, body := range bodies {
			added = t.trackRevLocked(body.ObjectKey, body.Rev) || added
		}
		if added {
			broadcast()
		}
	})
}

// trackRevLocked records one object read and reports whether it is new.
// The caller holds bcast.
func (t *TrackedWorldState) trackRevLocked(key string, rev uint64) bool {
	// Skip a key already recorded at this revision.
	if old, ok := t.revs[key]; ok && old == rev {
		return false
	}

	// Record the revision and queue the key for the next check.
	t.revs[key] = rev
	t.pending = append(t.pending, key)
	return true
}

// trackCoarse records a read that object revisions cannot describe.
func (t *TrackedWorldState) trackCoarse() {
	t.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		if !t.coarse {
			t.coarse = true
			broadcast()
		}
	})
}

// Close stops the change detection routine.
func (t *TrackedWorldState) Close() {
	t.cancel()
}

// watch returns nil once a recorded read differs from the live World.
func (t *TrackedWorldState) watch(ctx context.Context) error {
	// The recorded revisions describe initialSeqno, so a check there is free.
	checkedSeqno := t.initialSeqno
	for {
		// Sample the seqno before reading the revisions it describes, so a
		// commit the check misses lands above it and ends the wait below.
		seqno, err := t.watchWs.GetSeqno(ctx)
		if err != nil {
			return err
		}

		// Take the reads to check with the wait channel. A commit since the
		// last check needs every read checked; otherwise only the new ones.
		var (
			keys   []string
			revs   []uint64
			coarse bool
			idle   bool
			readCh <-chan struct{}
		)
		t.bcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			// Choose the keys and take the wait channel.
			readCh = getWaitCh()
			coarse = t.coarse
			idle = len(t.revs) == 0 && !coarse
			keys = t.pending
			if seqno != checkedSeqno {
				keys = slices.Collect(maps.Keys(t.revs))
			}
			t.pending = nil

			// Copy their recorded revisions.
			revs = make([]uint64, len(keys))
			for i, key := range keys {
				revs[i] = t.revs[key]
			}
		})

		// Report a change to a coarse read or to any read object.
		if coarse && seqno > t.initialSeqno {
			return nil
		}
		if seqno != t.initialSeqno {
			changed, err := t.revsChanged(ctx, keys, revs)
			if err != nil || changed {
				return err
			}
		}
		checkedSeqno = seqno

		// Wait for a new read, or with reads recorded for a later commit.
		if idle {
			select {
			case <-ctx.Done():
				return context.Cause(ctx)
			case <-readCh:
			}
			continue
		}
		if err := t.waitReadOrCommit(ctx, readCh, seqno); err != nil {
			return err
		}
	}
}

// revsChanged checks whether any of keys has a live revision other than revs.
// A missing object has revision 0.
func (t *TrackedWorldState) revsChanged(ctx context.Context, keys []string, revs []uint64) (bool, error) {
	// Read the live revisions in one batch.
	if len(keys) == 0 {
		return false, nil
	}
	refs, err := world.GetObjectRootRefsBatch(ctx, t.watchWs, keys)
	if err != nil {
		return false, err
	}

	// Compare them with the recorded revisions.
	for i, ref := range refs {
		var rev uint64
		if ref.Exists {
			rev = ref.Rev
		}
		if rev != revs[i] {
			return true, nil
		}
	}
	return false, nil
}

// waitReadOrCommit waits for readCh or a commit after seqno.
func (t *TrackedWorldState) waitReadOrCommit(ctx context.Context, readCh <-chan struct{}, seqno uint64) error {
	// Stop the seqno wait when a read is recorded.
	waitCtx, waitCancel := context.WithCancel(ctx)
	defer waitCancel()
	go func() {
		select {
		case <-readCh:
			waitCancel()
		case <-waitCtx.Done():
		}
	}()

	// A canceled wait is a recorded read unless ctx itself ended.
	_, err := t.watchWs.WaitSeqno(waitCtx, seqno+1)
	if err != nil && waitCtx.Err() != nil && ctx.Err() == nil {
		return nil
	}
	return err
}

// WorldState interface implementation - delegates to wrapped ws

// GetReadOnly returns the wrapped state's transaction mode.
func (t *TrackedWorldState) GetReadOnly() bool {
	return t.ws.GetReadOnly()
}

// Sync fences the block writes made through the tracked world state durable.
func (t *TrackedWorldState) Sync(ctx context.Context) (bool, error) {
	return t.ws.Sync(ctx)
}

// GetSeqno returns the wrapped state's sequence number.
func (t *TrackedWorldState) GetSeqno(ctx context.Context) (uint64, error) {
	return t.ws.GetSeqno(ctx)
}

// GetObjectBodiesBatchPage forwards budgeted body paging to the wrapped state.
func (t *TrackedWorldState) GetObjectBodiesBatchPage(ctx context.Context, keys []string, byteBudget int) ([]*world.ObjectBody, uint32, error) {
	// Page bodies through the wrapped state's batcher when available.
	var (
		bodies   []*world.ObjectBody
		consumed uint32
		err      error
	)
	if pager, ok := t.ws.(world.ObjectBodyPageBatcher); ok {
		bodies, consumed, err = pager.GetObjectBodiesBatchPage(ctx, keys, byteBudget)
	} else {
		bodies, consumed, err = world.GetObjectBodiesBatchPage(ctx, t.ws, keys, byteBudget)
	}
	if err != nil {
		return nil, 0, err
	}
	t.trackObjectBodyAccesses(bodies)
	return bodies, consumed, nil
}

// GetObjectBodiesBatchPageWithSeqno forwards body paging and the wrapped transaction seqno.
func (t *TrackedWorldState) GetObjectBodiesBatchPageWithSeqno(ctx context.Context, keys []string, byteBudget int) ([]*world.ObjectBody, uint32, uint64, error) {
	// Page bodies and read the wrapped transaction seqno.
	var (
		bodies   []*world.ObjectBody
		consumed uint32
		seqno    uint64
		err      error
	)
	if pager, ok := t.ws.(world.ObjectBodyPageSeqnoBatcher); ok {
		bodies, consumed, seqno, err = pager.GetObjectBodiesBatchPageWithSeqno(ctx, keys, byteBudget)
	} else {
		seqno, err = t.ws.GetSeqno(ctx)
		if err == nil {
			bodies, consumed, err = world.GetObjectBodiesBatchPage(ctx, t.ws, keys, byteBudget)
		}
	}
	if err != nil {
		return nil, 0, 0, err
	}
	t.trackObjectBodyAccesses(bodies)
	return bodies, consumed, seqno, nil
}

// WaitSeqno waits for the wrapped state to reach the requested sequence number.
func (t *TrackedWorldState) WaitSeqno(ctx context.Context, seqno uint64) (uint64, error) {
	return t.ws.WaitSeqno(ctx, seqno)
}

// BuildStorageCursor builds a cursor through the wrapped World storage.
func (t *TrackedWorldState) BuildStorageCursor(ctx context.Context) (*bucket_lookup.Cursor, error) {
	return t.ws.BuildStorageCursor(ctx)
}

// StageWorldState opens a staging scope through the wrapped state.
func (t *TrackedWorldState) StageWorldState(ctx context.Context) (world.WorldStage, error) {
	return t.ws.StageWorldState(ctx)
}

// AccessWorldState lends a storage cursor from the wrapped state to cb.
func (t *TrackedWorldState) AccessWorldState(ctx context.Context, ref *bucket.ObjectRef, cb func(*bucket_lookup.Cursor) error) error {
	return t.ws.AccessWorldState(ctx, ref, cb)
}

// CreateObject creates an object and records its revision when readable.
func (t *TrackedWorldState) CreateObject(ctx context.Context, key string, rootRef *bucket.ObjectRef) (world.ObjectState, error) {
	obj, err := t.ws.CreateObject(ctx, key, rootRef)
	if err == nil {
		// Record the created object's revision. A failed read is skipped rather
		// than recorded as rev=0, which the change detector reads as missing.
		if _, rev, revErr := obj.GetRootRef(ctx); revErr == nil {
			t.trackObjectAccess(key, rev)
		}
	}
	return obj, err
}

// GetObject looks up an object and records its revision or absence.
func (t *TrackedWorldState) GetObject(ctx context.Context, key string) (world.ObjectState, bool, error) {
	obj, found, err := t.ws.GetObject(ctx, key)
	if err == nil {
		// Record the requested object's revision, including a missing object.
		// A failed read is skipped rather than recorded as rev=0.
		if found {
			if _, rev, revErr := obj.GetRootRef(ctx); revErr == nil {
				t.trackObjectAccess(key, rev)
			}
		} else {
			t.trackObjectAccess(key, 0)
		}
	}
	return obj, found, err
}

// IterateObjects iterates objects and records a coarse read.
func (t *TrackedWorldState) IterateObjects(ctx context.Context, prefix string, reversed bool) world.ObjectIterator {
	t.trackCoarse()
	return t.ws.IterateObjects(ctx, prefix, reversed)
}

// RenameObject renames an object and records the affected objects and graph.
func (t *TrackedWorldState) RenameObject(ctx context.Context, oldKey, newKey string, descendants bool) (world.ObjectState, error) {
	obj, err := t.ws.RenameObject(ctx, oldKey, newKey, descendants)
	if err == nil {
		t.trackObjectAccess(oldKey, 0)
		if _, rev, revErr := obj.GetRootRef(ctx); revErr == nil {
			t.trackObjectAccess(newKey, rev)
		}
		t.trackCoarse()
	}
	return obj, err
}

// DeleteObject deletes an object through the wrapped state.
func (t *TrackedWorldState) DeleteObject(ctx context.Context, key string) (bool, error) {
	return t.ws.DeleteObject(ctx, key)
}

// AccessCayleyGraph records graph access and lends the wrapped graph handle to cb.
func (t *TrackedWorldState) AccessCayleyGraph(ctx context.Context, write bool, cb func(ctx context.Context, h world.CayleyHandle) error) error {
	t.trackCoarse()
	return t.ws.AccessCayleyGraph(ctx, write, cb)
}

// SetGraphQuad records graph access and sets a quad through the wrapped state.
func (t *TrackedWorldState) SetGraphQuad(ctx context.Context, q world.GraphQuad) error {
	t.trackCoarse()
	return t.ws.SetGraphQuad(ctx, q)
}

// DeleteGraphQuad records graph access and deletes a quad through the wrapped state.
func (t *TrackedWorldState) DeleteGraphQuad(ctx context.Context, q world.GraphQuad) error {
	t.trackCoarse()
	return t.ws.DeleteGraphQuad(ctx, q)
}

// LookupGraphQuads looks up quads and records a successful graph read.
func (t *TrackedWorldState) LookupGraphQuads(ctx context.Context, filter world.GraphQuad, limit uint32) ([]world.GraphQuad, error) {
	quads, err := t.ws.LookupGraphQuads(ctx, filter, limit)
	if err == nil {
		t.trackCoarse()
	}
	return quads, err
}

// LookupGraphQuadsBatch looks up quad batches and records a successful graph read.
func (t *TrackedWorldState) LookupGraphQuadsBatch(ctx context.Context, filters []world.GraphQuad, limitPerFilter uint32) ([][]world.GraphQuad, error) {
	quads, err := t.ws.LookupGraphQuadsBatch(ctx, filters, limitPerFilter)
	if err == nil {
		t.trackCoarse()
	}
	return quads, err
}

// QueryGraphPath queries a graph path and records a successful graph read.
func (t *TrackedWorldState) QueryGraphPath(ctx context.Context, query *world.GraphPathQuery) (*world.GraphPathQueryResult, error) {
	result, err := t.ws.QueryGraphPath(ctx, query)
	if err == nil {
		t.trackCoarse()
	}
	return result, err
}

// ListObjectsWithType records a successful type-index read and forwards its page.
func (t *TrackedWorldState) ListObjectsWithType(ctx context.Context, typeID, startAfter string, limit uint32) ([]string, bool, error) {
	// Retain the page result before marking the graph dependency.
	keys, more, err := t.ws.ListObjectsWithType(ctx, typeID, startAfter, limit)
	if err == nil {
		t.trackCoarse()
	}
	return keys, more, err
}

// DeleteGraphObject records graph access and deletes the object's quads.
func (t *TrackedWorldState) DeleteGraphObject(ctx context.Context, objKey string) error {
	t.trackCoarse()
	return t.ws.DeleteGraphObject(ctx, objKey)
}

// ApplyWorldOp applies a World operation through the wrapped state.
func (t *TrackedWorldState) ApplyWorldOp(ctx context.Context, op world.Operation, opSender peer.ID) (uint64, bool, error) {
	return t.ws.ApplyWorldOp(ctx, op, opSender)
}

// HasObject checks whether an object exists and records its revision.
func (t *TrackedWorldState) HasObject(ctx context.Context, key string) (bool, error) {
	// Read the root ref, which carries the revision HasObject does not.
	refs, err := world.GetObjectRootRefsBatch(ctx, t.ws, []string{key})
	if err != nil {
		return false, err
	}

	// Record the revision, 0 when missing.
	var rev uint64
	if refs[0].Exists {
		rev = refs[0].Rev
	}
	t.trackObjectAccess(key, rev)
	return refs[0].Exists, nil
}
