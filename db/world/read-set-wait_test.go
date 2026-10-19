package world

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// sharedChangeFeed delivers each published ChangeSet to every watcher, the
// same pointer to all of them, like the block engine's change watch.
type sharedChangeFeed struct {
	// mtx guards the fields below.
	mtx sync.Mutex
	// seqno is the newest published seqno.
	seqno uint64
	// subs receive each published change.
	subs map[chan feedChange]struct{}
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
	return &sharedChangeFeed{subs: make(map[chan feedChange]struct{}), joined: make(chan struct{})}
}

// GetSeqno returns the newest published seqno.
func (f *sharedChangeFeed) GetSeqno(context.Context) (uint64, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()
	return f.seqno, nil
}

// WaitSeqno is unused by waitChange with a change feed.
func (f *sharedChangeFeed) WaitSeqno(ctx context.Context, value uint64) (uint64, error) {
	<-ctx.Done()
	return 0, ctx.Err()
}

// WatchChanges calls cb with each change published after it subscribed.
func (f *sharedChangeFeed) WatchChanges(ctx context.Context, _ uint64, cb func(uint64, *ChangeSet) error) error {
	// Subscribe, waking anyone waiting for the subscription.
	ch := make(chan feedChange, 8192)
	f.mtx.Lock()
	f.subs[ch] = struct{}{}
	close(f.joined)
	f.joined = make(chan struct{})
	f.mtx.Unlock()
	defer func() {
		f.mtx.Lock()
		delete(f.subs, ch)
		f.mtx.Unlock()
	}()

	// Deliver each published change until ctx ends.
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

// publish sends changes to every watcher at the next seqno and returns it.
func (f *sharedChangeFeed) publish(changes *ChangeSet) uint64 {
	// Advance the seqno and hand the same ChangeSet to every watcher.
	f.mtx.Lock()
	defer f.mtx.Unlock()
	f.seqno++
	for ch := range f.subs {
		ch <- feedChange{seqno: f.seqno, changes: changes}
	}
	return f.seqno
}

// waitResult is one waitChange's return.
type waitResult struct {
	seqno uint64
	err   error
}

// startWait runs waitChange on reads in the background, reporting each
// untouched seqno.
func startWait(ctx context.Context, f *sharedChangeFeed, reads *ReadSet, shared []*ReadSet) (<-chan waitResult, <-chan uint64) {
	result := make(chan waitResult, 1)
	untouched := make(chan uint64, 8192)
	go func() {
		s, err := reads.waitChange(ctx, f, 0, shared, func(s uint64) error {
			untouched <- s
			return nil
		})
		result <- waitResult{s, err}
	}()
	return result, untouched
}

// awaitUntouched waits for the wait to report seqno untouched.
func awaitUntouched(t *testing.T, ctx context.Context, untouched <-chan uint64, result <-chan waitResult, seqno uint64) {
	t.Helper()
	for {
		select {
		case <-ctx.Done():
			t.Fatalf("seqno %d was not reported untouched", seqno)
		case r := <-result:
			t.Fatalf("wait returned %d, %v before seqno %d", r.seqno, r.err, seqno)
		case s := <-untouched:
			if s >= seqno {
				return
			}
		}
	}
}

// TestWaitChangeSharedMemo checks that views sharing a finished set test each
// delivered change against it with the delivered ChangeSet, so the set's memo
// answers every view after the first, from the second change on too.
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
	untouched := make([]<-chan uint64, views)
	for i := range views {
		results[i], untouched[i] = startWait(ctx, f, NewReadSet(), []*ReadSet{shared})
	}
	f.waitSubscribers(t, ctx, views)

	// Each unrelated change is tested against the shared set as delivered.
	for i := range 3 {
		changes := &ChangeSet{Keys: []string{fmt.Sprintf("other/%d", i)}}
		seqno := f.publish(changes)
		for v := range views {
			awaitUntouched(t, ctx, untouched[v], results[v], seqno)
		}
		shared.mtx.Lock()
		memo := shared.memoChanges
		shared.mtx.Unlock()
		if memo != changes {
			t.Fatalf("change %d: shared set last tested another ChangeSet than the delivered one", i)
		}
	}

	// A change to the shared read ends every wait.
	seqno := f.publish(&ChangeSet{Keys: []string{"shared"}})
	for v := range views {
		if r := <-results[v]; r.err != nil || r.seqno != seqno {
			t.Fatalf("view %d returned %d, %v; want %d", v, r.seqno, r.err, seqno)
		}
	}
}

// TestWaitChangeQuietPastLimit checks that a wait outlasts more unrelated
// changes than it keeps, and still ends on a change to its reads.
func TestWaitChangeQuietPastLimit(t *testing.T) {
	// Wait on a view that reads a single key.
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	f := newSharedChangeFeed()
	reads := NewReadSet()
	reads.addKeys("mine")
	result, untouched := startWait(ctx, f, reads, nil)
	f.waitSubscribers(t, ctx, 1)

	// Publish more unrelated entries than the wait keeps.
	var seqno uint64
	for i := range readSetChangeLimit/64 + 2 {
		keys := make([]string, 64)
		for k := range keys {
			keys[k] = fmt.Sprintf("other/%d/%d", i, k)
		}
		seqno = f.publish(&ChangeSet{Keys: keys})
	}
	awaitUntouched(t, ctx, untouched, result, seqno)

	// A change to the read still ends the wait.
	seqno = f.publish(&ChangeSet{Keys: []string{"mine"}})
	if r := <-result; r.err != nil || r.seqno != seqno {
		t.Fatalf("wait returned %d, %v; want %d", r.seqno, r.err, seqno)
	}
}

// TestWaitChangeLateRead checks that a read recorded after a change it matches
// ends the wait, and that once the kept changes overflow any new read does.
func TestWaitChangeLateRead(t *testing.T) {
	// Bound both waits.
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()

	// A change before the read of its key ends the wait at the read.
	f := newSharedChangeFeed()
	reads := NewReadSet()
	result, untouched := startWait(ctx, f, reads, nil)
	f.waitSubscribers(t, ctx, 1)
	seqno := f.publish(&ChangeSet{Keys: []string{"late"}})
	awaitUntouched(t, ctx, untouched, result, seqno)
	reads.addKeys("late")
	if r := <-result; r.err != nil || r.seqno != seqno {
		t.Fatalf("wait returned %d, %v; want %d", r.seqno, r.err, seqno)
	}

	// Start a second wait for the overflow case.
	f = newSharedChangeFeed()
	reads = NewReadSet()
	result, untouched = startWait(ctx, f, reads, nil)
	f.waitSubscribers(t, ctx, 1)

	// Past the kept changes, an unrelated new read ends the wait too, since
	// the dropped changes might have touched it.
	keys := make([]string, readSetChangeLimit+1)
	for k := range keys {
		keys[k] = fmt.Sprintf("other/%d", k)
	}
	seqno = f.publish(&ChangeSet{Keys: keys})
	awaitUntouched(t, ctx, untouched, result, seqno)
	reads.addKeys("unrelated")
	if r := <-result; r.err != nil || r.seqno != seqno {
		t.Fatalf("wait returned %d, %v; want %d", r.seqno, r.err, seqno)
	}
}
