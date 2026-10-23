package plugin_host_scheduler

import (
	"context"
	"testing"
	"time"

	bldr_manifest_world "github.com/s4wave/spacewave/bldr/manifest/world"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/testbed"
	"github.com/sirupsen/logrus"
)

// TestBucketMountWakeOnMount checks that a candidate's unmounted bucket wakes
// the selection loop once its config is applied, and not before.
func TestBucketMountWakeOnMount(t *testing.T) {
	// Open a testbed whose bus resolves bucket lookups.
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	le := logrus.NewEntry(logrus.New())
	tb, err := testbed.NewTestbed(ctx, le)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(tb.Release)

	// Watch a bucket no volume serves yet.
	const bucketID = "release-bucket"
	woke := make(chan struct{}, 1)
	mounts := newBucketMountWake(ctx, tb.Bus, func() {
		select {
		case woke <- struct{}{}:
		default:
		}
	})
	candidates := []*bldr_manifest_world.StartupManifestCandidateEligibility{
		{ObjectKey: "ref/release", UnmountedBucketID: bucketID},
		{ObjectKey: "ref/local"},
	}
	if err := mounts.watch(ctx, candidates); err != nil {
		t.Fatal(err)
	}

	// Expect no wake while the bucket stays unmounted.
	select {
	case <-woke:
		t.Fatal("woke before the bucket mounted")
	case <-time.After(100 * time.Millisecond):
	}

	// Mount the bucket and expect the wake.
	apply := bucket.NewApplyBucketConfigToVolume(&bucket.Config{Id: bucketID, Rev: 1}, tb.Volume.GetID())
	if _, err := bucket.ExApplyBucketConfig(ctx, tb.Bus, apply); err != nil {
		t.Fatal(err)
	}
	select {
	case <-woke:
	case <-ctx.Done():
		t.Fatal("no wake after the bucket mounted")
	}
}
