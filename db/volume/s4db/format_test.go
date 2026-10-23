//go:build !js && !wasip1

package volume_s4db_test

import (
	"context"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/directive"
	backoff "github.com/aperturerobotics/util/backoff/cbackoff"
	"github.com/s4wave/spacewave/db/core"
	"github.com/s4wave/spacewave/db/s4db"
	store_kvkey "github.com/s4wave/spacewave/db/store/kvkey"
	"github.com/s4wave/spacewave/db/volume"
	volume_s4db "github.com/s4wave/spacewave/db/volume/s4db"
	"github.com/sirupsen/logrus"
)

// restartCounter is a retry backoff that counts the loader's restarts.
type restartCounter struct {
	// restarts is the number of times the loader asked for a retry delay.
	restarts atomic.Int32
}

// NextBackOff counts a restart and delays it past the test.
func (c *restartCounter) NextBackOff() time.Duration {
	c.restarts.Add(1)
	return time.Hour
}

// Reset does nothing: the counter keeps its total.
func (c *restartCounter) Reset() {}

// TestOpenNewerFormatFails checks that a Volume file written at a newer format
// than this binary supports fails a waiting open with the format error, and
// that the loader does not restart the controller.
func TestOpenNewerFormatFails(t *testing.T) {
	// Write a file at the next format version.
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "volume.s4wave")
	writeFormatVersion(t, path, store_kvkey.FormatVersion+1)

	// Register the Volume factory on a test bus.
	b, sr, err := core.NewCoreBus(ctx, logrus.NewEntry(logrus.New()))
	if err != nil {
		t.Fatal(err)
	}
	factory := volume_s4db.NewFactory(b)
	sr.AddFactory(factory)

	// Run the Volume controller, counting the loader's restarts.
	restarts := &restartCounter{}
	execDir := loader.NewExecControllerWithOpts(
		factory,
		&volume_s4db.Config{Path: path},
		func() backoff.BackOff { return restarts },
		directive.ValueOptions{},
	)
	_, execRef, err := b.AddDirective(ctx, execDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer execRef.Release()

	// Wait for the open to fail.
	lookupCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, _, ref, err := volume.ExLookupVolume(lookupCtx, b, "", "", false)
	if ref != nil {
		ref.Release()
	}
	if !volume.IsPermanent(err) || !strings.Contains(err.Error(), "is newer than supported version") {
		t.Fatalf("ExLookupVolume = %v, want the newer format error", err)
	}

	// The loader keeps the exited controller instead of restarting it.
	if n := restarts.restarts.Load(); n != 0 {
		t.Fatalf("the loader restarted the controller %d times", n)
	}
}

// writeFormatVersion creates the s4db file at path recording version as its
// store format.
func writeFormatVersion(t *testing.T, path string, version uint64) {
	// Open the file, closed on return.
	db, err := s4db.Open(path, s4db.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Record the version.
	ctx := t.Context()
	tx, err := db.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()
	key := store_kvkey.NewDefaultKVKey().GetFormatVersionKey()
	if err := tx.Set(ctx, key, store_kvkey.MarshalFormatVersion(version)); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
