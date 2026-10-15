package worker_controller

import (
	"context"

	"github.com/aperturerobotics/util/keyed"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/world"
	world_control "github.com/s4wave/spacewave/db/world/control"
	"github.com/sirupsen/logrus"
)

// keypairTracker tracks a Keypair linked to the Worker.
//
// Linking an object to the Keypair increments its revision, which wakes the
// Worker to scan for objects to track.
type keypairTracker struct {
	// c is the controller
	c *Controller
	// objKey is the object key
	objKey string
	// objLoop is the object tracking loop
	objLoop *world_control.WatchLoop
}

// newKeypairTracker constructs a new worker keypair tracker routine.
func (c *Controller) newKeypairTracker(key string) (keyed.Routine, *keypairTracker) {
	tr := &keypairTracker{
		c:      c,
		objKey: key,
	}
	tr.objLoop = world_control.NewWatchLoop(
		c.le.WithField("object-loop", "keypair-tracker"),
		key,
		tr.processState,
	)
	return tr.execute, tr
}

// execute executes the keypair tracker.
func (t *keypairTracker) execute(ctx context.Context) error {
	objKey, le := t.objKey, t.c.le

	le.Debugf("starting keypair tracker: %s", objKey)
	return world_control.ExecuteBusWatchLoop(
		ctx,
		t.c.bus,
		t.c.conf.GetEngineId(),
		true, t.objLoop,
	)
}

// processState wakes the Worker on each observed Keypair revision.
func (t *keypairTracker) processState(
	ctx context.Context,
	le *logrus.Entry,
	ws world.WorldState,
	obj world.ObjectState, // may be nil if not found
	rootRef *bucket.ObjectRef, rev uint64,
) (waitForChanges bool, err error) {
	// Wake on the first observation too: an object linked between the
	// Worker's scan and this tracker's first read appears only here.
	t.c.Wake()
	return true, nil
}

// _ is a type assertion
var _ world_control.WatchLoopHandler = (*keypairTracker)(nil).processState
