//go:build !js && !wasip1

package world_block_engine

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/s4wave/spacewave/db/bucket"
	store_kvtx_bolt "github.com/s4wave/spacewave/db/store/kvtx/bolt"
)

// TestLoadHeadStateWithActiveReader requires head reads to coexist with retained snapshots.
func TestLoadHeadStateWithActiveReader(t *testing.T) {
	// Seed the persisted head through the same Bolt store used by native controllers.
	store, err := store_kvtx_bolt.Open(filepath.Join(t.TempDir(), "head.db"), 0o600, nil, []byte("head"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.GetDB().Close() })
	controller := &Controller{conf: &Config{}}
	want := &bucket.ObjectRef{BucketId: "retained"}
	if err := controller.writeHeadState(t.Context(), store, nil, want); err != nil {
		t.Fatal(err)
	}

	// Hold an existing read snapshot while another caller loads the current head.
	reader, err := store.NewTransaction(t.Context(), false)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		state, found, err := controller.loadHeadState(t.Context(), store)
		if err == nil && (!found || !state.GetHeadRef().EqualsRef(want)) {
			t.Error("head read did not return the persisted reference")
		}
		done <- err
	}()

	// Release and join on failure as well, so the regression never strands a reader.
	select {
	case err := <-done:
		reader.Discard()
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		reader.Discard()
		if err := <-done; err != nil {
			t.Error(err)
		}
		t.Fatal("head read waited for an unrelated read snapshot to close")
	}
}
