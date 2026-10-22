package world_block

import (
	"context"
	"runtime"
	"sync"
	"testing"

	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/testbed"
	"github.com/sirupsen/logrus"
)

// TestReadSnapshotAccessDuringRelease reads a read snapshot's storage from
// several goroutines while Close or Discard releases the head it holds. Under
// the race detector it checks that no read clones the held cursor as it is
// released, and that every read either succeeds or reports the Engine closed.
func TestReadSnapshotAccessDuringRelease(t *testing.T) {
	// Skip the race when the process has one processor.
	if runtime.GOMAXPROCS(0) < 2 {
		t.Skip("requires concurrent execution")
	}

	// Open a testbed.
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	le := logrus.NewEntry(log)
	tb, err := testbed.NewTestbed(ctx, le, testbed.WithVerbose(false))
	if err != nil {
		t.Fatal(err)
	}
	defer tb.Release()

	// Build an empty World cursor.
	base, err := tb.BuildEmptyCursor(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer base.Release()

	// Release the snapshot by Close or by Discard while readers run.
	for _, viaClose := range []bool{true, false} {
		for range 16 {
			// Open an engine and a read snapshot on it.
			eng, err := NewEngine(ctx, le, base, nil, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			rtx, err := eng.NewBlockEngineTransaction(ctx, false)
			if err != nil {
				t.Fatal(err)
			}

			// Read through the snapshot until it reports the Engine closed or
			// the snapshot discarded.
			entered := make(chan struct{})
			var enteredOnce sync.Once
			stop := make(chan struct{})
			var wg sync.WaitGroup
			for range 8 {
				wg.Go(func() {
					for {
						err := rtx.AccessWorldState(ctx, rtx.readRoot, func(*bucket_lookup.Cursor) error {
							enteredOnce.Do(func() { close(entered) })
							return nil
						})
						if err == ErrEngineClosed {
							return
						}
						if err != nil {
							t.Errorf("access failed: %v", err)
							return
						}
						select {
						case <-stop:
							return
						default:
						}
					}
				})
			}

			// Release the held head while the reads run.
			<-entered
			if viaClose {
				if err := eng.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				rtx.Discard()
			}
			close(stop)
			wg.Wait()
			rtx.Discard()
			if err := eng.Close(); err != nil {
				t.Fatal(err)
			}
		}
	}
}
