package world

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"sync"
)

// errReadsTouched ends a wait's change subscription at the first change that
// touches its reads.
var errReadsTouched = errors.New("reads touched")

// ReadSet records what a view read from the World, so a ChangeSet can be
// checked against it. Reads it cannot describe precisely mark it conservative,
// and any change touches it. A ReadSet can track several WorldStates at once.
type ReadSet struct {
	// mtx guards the fields below.
	mtx sync.Mutex
	// keys are the object keys read, including keys that were not found.
	keys map[string]struct{}
	// prefixes are the object key prefixes iterated.
	prefixes []string
	// patterns are the graph quad lookup filters.
	patterns []graphPattern
	// endpoints are the object keys that patterns name as subject or object.
	endpoints map[string]struct{}
	// graph is set when the graph was read beyond quad lookups.
	graph bool
	// all is set when the World was read beyond objects and the graph.
	all bool
	// memoChanges and memoTouched cache Touched for the last ChangeSet it
	// checked, so views sharing this set test each change window once. Any
	// recorded read clears memoChanges.
	memoChanges *ChangeSet
	memoTouched bool
	// added closes when a read is recorded and is then replaced.
	added chan struct{}
}

// NewReadSet constructs an empty ReadSet.
func NewReadSet() *ReadSet {
	return &ReadSet{
		keys:      make(map[string]struct{}),
		endpoints: make(map[string]struct{}),
		added:     make(chan struct{}),
	}
}

// Track returns a WorldState that records the reads made through it.
// The seqno of ws identifies the state the reads observed.
func (r *ReadSet) Track(ws WorldState) WorldState {
	return &readSetWorldState{WorldState: ws, reads: r}
}

// Touched checks whether a ChangeSet may change what was read.
func (r *ReadSet) Touched(changes *ChangeSet) bool {
	// Hold the reads while comparing.
	r.mtx.Lock()
	defer r.mtx.Unlock()

	// A ChangeSet is immutable once delivered, so its answer holds until a read
	// is recorded.
	if changes == r.memoChanges {
		return r.memoTouched
	}
	touched := r.touchedLocked(changes)
	r.memoChanges, r.memoTouched = changes, touched
	return touched
}

// touchedLocked checks changes against the reads. The caller holds mtx.
func (r *ReadSet) touchedLocked(changes *ChangeSet) bool {
	// Unknown and conservative reads match every non-empty change.
	if changes.Unknown {
		return true
	}
	if len(changes.Keys) == 0 && len(changes.Quads) == 0 {
		return false
	}
	if r.all {
		return true
	}

	// A changed key matches a read key, a read prefix, or a pattern endpoint.
	if slices.ContainsFunc(changes.Keys, r.touchesKey) {
		return true
	}

	// A changed quad matches a graph read.
	for _, q := range changes.Quads {
		if r.graph {
			return true
		}
		for _, p := range r.patterns {
			if p.matches(q) {
				return true
			}
		}
	}
	return false
}

// WaitChange waits for the first change of w after afterSeqno that touches the
// reads, and returns its seqno. Record afterSeqno before the reads. Reads may
// continue while WaitChange runs: each one restarts the wait from afterSeqno,
// so a change that lands before the read of what it changed still ends it.
func (r *ReadSet) WaitChange(ctx context.Context, w WorldWaitSeqno, afterSeqno uint64) (uint64, error) {
	return r.waitChange(ctx, w, afterSeqno, nil)
}

// waitChange is WaitChange that also ends when a change touches any of the
// shared finished sets. The change feed tests each change against the sets,
// so a change that touches none of them never wakes the wait, and a shared
// set answers every view waiting on it from its memo.
func (r *ReadSet) waitChange(ctx context.Context, w WorldWaitSeqno, afterSeqno uint64, shared []*ReadSet) (uint64, error) {
	// touches checks one ChangeSet against the reads and the shared sets.
	touches := func(cs *ChangeSet) bool {
		return r.Touched(cs) || slices.ContainsFunc(shared, func(set *ReadSet) bool { return set.Touched(cs) })
	}
	for {
		// Capture the read signal before subscribing, so a read recorded
		// after the subscription starts restarts it.
		r.mtx.Lock()
		added := r.added
		r.mtx.Unlock()

		// Wait in the change feed for the first touching change since
		// afterSeqno.
		watchCtx, cancel := context.WithCancel(ctx)
		done := make(chan error, 1)
		var seqno uint64
		go func() {
			done <- WatchTouchingChanges(watchCtx, w, afterSeqno, touches, func(s uint64, _ *ChangeSet) error {
				seqno = s
				return errReadsTouched
			})
		}()

		// End at the touching change, or restart from afterSeqno after a new
		// read so the feed tests the earlier changes against it too.
		var err error
		select {
		case err = <-done:
		case <-added:
			cancel()
			if err = <-done; errors.Is(err, context.Canceled) && ctx.Err() == nil {
				continue
			}
		}
		cancel()
		if errors.Is(err, errReadsTouched) {
			return seqno, nil
		}
		return 0, err
	}
}

// touchesKey checks a changed object key. The caller holds mtx.
func (r *ReadSet) touchesKey(key string) bool {
	// Match the key itself and the iterated prefixes.
	if _, ok := r.keys[key]; ok {
		return true
	}
	for _, prefix := range r.prefixes {
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}

	// Match a graph read naming the key as a subject or object.
	if r.graph {
		return true
	}
	_, ok := r.endpoints[key]
	return ok
}

// addKeys records object key reads.
func (r *ReadSet) addKeys(keys ...string) {
	r.mtx.Lock()
	for _, key := range keys {
		r.keys[key] = struct{}{}
	}
	r.signalLocked()
	r.mtx.Unlock()
}

// addPrefix records an object key prefix read.
func (r *ReadSet) addPrefix(prefix string) {
	r.mtx.Lock()
	r.prefixes = append(r.prefixes, prefix)
	r.signalLocked()
	r.mtx.Unlock()
}

// addPatterns records graph quad lookups.
func (r *ReadSet) addPatterns(filters ...GraphQuad) {
	r.mtx.Lock()
	for _, q := range filters {
		// Keep the filter for changed quads.
		p := newGraphPattern(q)
		r.patterns = append(r.patterns, p)

		// Index the object keys it names for changed objects.
		for _, value := range []string{p.subject, p.obj} {
			if key, err := GraphValueToKey(value); err == nil && key != "" {
				r.endpoints[key] = struct{}{}
			}
		}
	}
	r.signalLocked()
	r.mtx.Unlock()
}

// markGraph records a graph read that quad patterns cannot describe.
func (r *ReadSet) markGraph() {
	r.mtx.Lock()
	r.graph = true
	r.signalLocked()
	r.mtx.Unlock()
}

// markAll records a read that object keys and graph reads cannot describe.
func (r *ReadSet) markAll() {
	r.mtx.Lock()
	r.all = true
	r.signalLocked()
	r.mtx.Unlock()
}

// addReadSet records every read other recorded.
func (r *ReadSet) addReadSet(other *ReadSet) {
	// Copy other's reads without holding both locks.
	other.mtx.Lock()
	keys := maps.Clone(other.keys)
	prefixes := slices.Clone(other.prefixes)
	patterns := slices.Clone(other.patterns)
	endpoints := maps.Clone(other.endpoints)
	graph, all := other.graph, other.all
	other.mtx.Unlock()

	// Join the object and graph reads to r's.
	r.mtx.Lock()
	maps.Copy(r.keys, keys)
	r.prefixes = append(r.prefixes, prefixes...)
	r.patterns = append(r.patterns, patterns...)
	maps.Copy(r.endpoints, endpoints)

	// Keep the conservative marks and wake the waits.
	r.graph = r.graph || graph
	r.all = r.all || all
	r.signalLocked()
	r.mtx.Unlock()
}

// signalLocked wakes the waits checking the reads and drops the cached Touched
// answer. The caller holds mtx.
func (r *ReadSet) signalLocked() {
	r.memoChanges = nil
	close(r.added)
	r.added = make(chan struct{})
}
