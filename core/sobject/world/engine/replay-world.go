package sobject_world_engine

import (
	"context"

	"github.com/s4wave/spacewave/core/sobject"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
	trace "github.com/s4wave/spacewave/db/traceutil"
	world_block "github.com/s4wave/spacewave/db/world/block"
)

// replayWorld serves the World a replay pass advances through one block
// engine. The engine defers block durability, so a pass makes the blocks of
// all its operations durable in one fence at close instead of one per
// operation. The owner serializes calls.
type replayWorld struct {
	// c builds the engine.
	c *Controller
	// so holds the World blocks.
	so sobject.SharedObject
	// ws serves head, or is nil before the first bind.
	ws *blkEngine
	// head is the World ws serves.
	head *bucket.ObjectRef
	// base is the World ws was built on.
	base *block.BlockRef
	// roots are the roots ws committed, proven complete at close when base is.
	roots []*block.BlockRef
}

// newReplayWorld constructs a replay World over the blocks of so.
func newReplayWorld(c *Controller, so sobject.SharedObject) *replayWorld {
	return &replayWorld{c: c, so: so}
}

// bind returns the engine serving head. It rebuilds the engine when the pass
// moved to a World the engine did not commit.
func (w *replayWorld) bind(ctx context.Context, head *bucket.ObjectRef) (*blkEngine, error) {
	// Reuse the engine that committed head.
	if w.ws != nil && w.head.EqualVT(head) {
		return w.ws, nil
	}

	// Fence the previous World before binding the next.
	if err := w.close(ctx); err != nil {
		return nil, err
	}
	ws, err := w.c.buildBlkEngine(ctx, w.c.le, w.so, head, head.GetTransformConf(), world_block.WithDeferredDurability())
	if err != nil {
		return nil, err
	}
	w.ws, w.head, w.base = ws, head.CloneVT(), head.GetRootRef().CloneVT()
	return ws, nil
}

// advance records that the engine committed head.
func (w *replayWorld) advance(head *bucket.ObjectRef) {
	w.head = head.CloneVT()
	w.roots = append(w.roots, head.GetRootRef())
}

// close makes the committed blocks durable, proves their roots complete when
// the base World is, and releases the engine.
func (w *replayWorld) close(ctx context.Context) error {
	// Detach the engine.
	if w.ws == nil {
		return nil
	}
	ws, base, roots := w.ws, w.base, w.roots
	w.ws, w.head, w.base, w.roots = nil, nil, nil, nil
	defer ws.Release()

	// A completion proof lands only after the fence covers its writes.
	ctx, task := trace.NewTask(ctx, "alpha/so-engine/replay/fence")
	defer task.End()
	if _, err := ws.bengine.Sync(ctx); err != nil {
		return err
	}
	return block.MarkDerivedRootsComplete(ctx, w.so.GetBlockStore(), base, roots...)
}
