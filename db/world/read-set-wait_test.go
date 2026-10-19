package world

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// sharedChangeFeed delivers each published ChangeSet, the same pointer to
// every watcher its filter accepts, like the block engine's change watch. A
// watcher starting behind replays the published changes since its seqno.
type sharedChangeFeed struct {
	// mtx guards the fields below.
	mtx sync.Mutex
	// history lists every published change, oldest first.
	history []feedChange
	// subs receive each published change their filter accepts.
	subs map[chan feedChange]ChangeFilter
	// woken counts the changes handed to watchers.
	woken int
	// joined closes and is replaced when a watcher subscribes.
	joined chan struct{}
}

// feedChange is one published seqno and its changes.
type feedChange struct {
	seqno   uint64
	changes *ChangeSet
}

// newSharedChangeFeed constructs a feed at seqno 0.
func newSharedChangeFeed() *sharedChangeFeed {
	return &sharedChangeFeed{subs: make(map[chan feedChange]ChangeFilter), joined: make(chan struct{})}
}

// GetSeqno returns the newest published seqno.
func (f *sharedChangeFeed) GetSeqno(context.Context) (uint64, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	return uint64(len(f.history)), nil
}

// WaitSeqno is unused by waitChange with a change feed.
func (f *sharedChangeFeed) WaitSeqno(ctx context.Context, value uint64) (uint64, error) {
	<-ctx.Done()
	return 0, ctx.Err()
}

// WatchChanges replays the accepted changes after afterSeqno, then delivers
// each accepted change published after it subscribed.
func (f *sharedChangeFeed) WatchChanges(ctx context.Context, afterSeqno uint64, touches ChangeFilter, cb func(uint64, *ChangeSet) error) error {
	// Subscribe and queue the accepted history, waking anyone waiting for the
	// subscription.
	ch := make(chan feedChange, 1024)
	f.mtx.Lock()
	for _, c := range f.history[afterSeqno:] {
		if touches.Accepts(c.changes) {
			f.woken++
			ch <- c
		}
	}
	f.subs[ch] = touches
	close(f.joined)
	f.joined = make(chan struct{})
	f.mtx.Unlock()
	defer func() {
		f.mtx.Lock()
		delete(f.subs, ch)
		f.mtx.Unlock()
	}()

	// Deliver each accepted change until ctx ends.
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case c := <-ch:
			if err := cb(c.seqno, c.changes); err != nil {
				return err
			}
		}
	}
}

// waitSubscribers waits until n watchers subscribed.
func (f *sharedChangeFeed) waitSubscribers(t *testing.T, ctx context.Context, n int) {
	t.Helper()
	for {
		f.mtx.Lock()
		count, joined := len(f.subs), f.joined
		f.mtx.Unlock()
		if count >= n {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("%d of %d watchers subscribed", count, n)
		case <-joined:
		}
	}
}

// publish hands changes at the next seqno to every watcher whose filter
// accepts them and returns the seqno.
func (f *sharedChangeFeed) publish(changes *ChangeSet) uint64 {
	// Record the change and test it against each watcher's filter.
	f.mtx.Lock()
	defer f.mtx.Unlock()
	c := feedChange{seqno: uint64(len(f.history)) + 1, changes: changes}
	f.history = append(f.history, c)
	for ch, touches := range f.subs {
		if touches.Accepts(changes) {
			f.woken++
			ch <- c
		}
	}
	return c.seqno
}

// wokenCount returns the changes handed to watchers so far.
func (f *sharedChangeFeed) wokenCount() int {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	return f.woken
}

// waitResult is one waitChange's return.
type waitResult struct {
	seqno uint64
	err   error
}

// startWait runs waitChange on reads from seqno 0 in the background.
func startWait(ctx context.Context, f *sharedChangeFeed, reads *ReadSet, shared []*ReadSet) <-chan waitResult {
	result := make(chan waitResult, 1)
	go func() {
		s, err := reads.waitChange(ctx, f, 0, shared)
		result <- waitResult{s, err}
	}()
	return result
}

// expectResult checks that a wait returned seqno.
func expectResult(t *testing.T, ctx context.Context, result <-chan waitResult, seqno uint64) {
	t.Helper()
	select {
	case <-ctx.Done():
		t.Fatalf("wait did not return seqno %d", seqno)
	case r := <-result:
		if r.err != nil || r.seqno != seqno {
			t.Fatalf("wait returned %d, %v; want %d", r.seqno, r.err, seqno)
		}
	}
}

// TestWaitChangeFilter checks that changes a wait's reads do not touch never
// wake it, and that the first touching change ends it once.
func TestWaitChangeFilter(t *testing.T) {
	// Wait on a view that reads a single key.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	f := newSharedChangeFeed()
	reads := NewReadSet()
	reads.addKeys("mine")
	result := startWait(ctx, f, reads, nil)
	f.waitSubscribers(t, ctx, 1)

	// Unrelated changes reach no watcher.
	for i := range 100 {
		f.publish(&ChangeSet{Keys: []string{fmt.Sprintf("other/%d", i)}})
	}
	if n := f.wokenCount(); n != 0 {
		t.Fatalf("unrelated changes woke the wait %d times", n)
	}

	// A change to the read ends the wait with one wake.
	seqno := f.publish(&ChangeSet{Keys: []string{"mine"}})
	expectResult(t, ctx, result, seqno)
	if n := f.wokenCount(); n != 1 {
		t.Fatalf("touching change woke the wait %d times, want 1", n)
	}
}

// TestWaitChangeLateRead checks that a read recorded after a change it matches
// restarts the wait from its seqno, so the earlier change ends it.
func TestWaitChangeLateRead(t *testing.T) {
	// Wait on a view that has read nothing yet.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	f := newSharedChangeFeed()
	reads := NewReadSet()
	result := startWait(ctx, f, reads, nil)
	f.waitSubscribers(t, ctx, 1)

	// Change a key, then an unrelated one, before the read of the first.
	seqno := f.publish(&ChangeSet{Keys: []string{"late"}})
	f.publish(&ChangeSet{Keys: []string{"other"}})
	if n := f.wokenCount(); n != 0 {
		t.Fatalf("changes before the read woke the wait %d times", n)
	}

	// The read restarts the wait, which replays the change to its key.
	reads.addKeys("late")
	expectResult(t, ctx, result, seqno)
}

// TestWaitChangeSharedMemo checks that views sharing a finished set have the
// feed test each change against it once, through the set's memo of the
// delivered ChangeSet, and that a change to it ends every view.
func TestWaitChangeSharedMemo(t *testing.T) {
	// Share one finished set that reads a single key.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	f := newSharedChangeFeed()
	shared := NewReadSet()
	shared.addKeys("shared")

	// Start several views that read nothing beyond the shared set.
	const views = 8
	results := make([]<-chan waitResult, views)
	for i := range views {
		results[i] = startWait(ctx, f, NewReadSet(), []*ReadSet{shared})
	}
	f.waitSubscribers(t, ctx, views)

	// Each unrelated change leaves the shared set's memo on the delivered
	// ChangeSet and wakes no view.
	for i := range 3 {
		changes := &ChangeSet{Keys: []string{fmt.Sprintf("other/%d", i)}}
		f.publish(changes)
		shared.mtx.Lock()
		memo := shared.memoChanges
		shared.mtx.Unlock()
		if memo != changes {
			t.Fatalf("change %d: shared set last tested another ChangeSet than the delivered one", i)
		}
	}
	if n := f.wokenCount(); n != 0 {
		t.Fatalf("unrelated changes woke the views %d times", n)
	}

	// A change to the shared read ends every wait.
	seqno := f.publish(&ChangeSet{Keys: []string{"shared"}})
	for v := range views {
		expectResult(t, ctx, results[v], seqno)
	}
}

// TestWaitChangeUnknown checks that an unknown change ends every wait.
func TestWaitChangeUnknown(t *testing.T) {
	// Start waits with disjoint reads.
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	f := newSharedChangeFeed()
	const views = 4
	results := make([]<-chan waitResult, views)
	for i := range views {
		reads := NewReadSet()
		reads.addKeys(fmt.Sprintf("view/%d", i))
		results[i] = startWait(ctx, f, reads, nil)
	}
	f.waitSubscribers(t, ctx, views)

	// An unknown change could touch any of them.
	seqno := f.publish(NewUnknownChangeSet())
	for v := range views {
		expectResult(t, ctx, results[v], seqno)
	}
}
