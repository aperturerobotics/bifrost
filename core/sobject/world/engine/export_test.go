package sobject_world_engine

import (
	"context"
	"time"

	"github.com/s4wave/spacewave/core/sobject"
)

// SetSlowCommitThreshold sets the duration after which a World write logs its
// stage times and returns a function that restores the previous one.
func SetSlowCommitThreshold(threshold time.Duration) func() {
	previous := slowCommitThreshold
	slowCommitThreshold = threshold
	return func() { slowCommitThreshold = previous }
}

// ExecuteWorld serves the World of so on c until ctx ends or an error ends it,
// as Execute does for the SharedObject it mounts.
func ExecuteWorld(ctx context.Context, c *Controller, so sobject.SharedObject) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	return c.executeSharedObject(ctx, cancel, so)
}

// WaitInstalled calls check after each World the engine of c installs, and
// once first, until it returns true or an error.
func WaitInstalled(ctx context.Context, c *Controller, check func() (bool, error)) error {
	for {
		// Read the wait channel before checking, so no install is missed.
		var wait <-chan struct{}
		c.writeBcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			wait = getWaitCh()
		})
		done, err := check()
		if err != nil || done {
			return err
		}

		// Wait for the next install.
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-wait:
		}
	}
}
