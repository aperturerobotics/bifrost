package s4wave_flowgraph

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/s4wave/spacewave/db/world"
)

// BindingWatch follows the derived binding of one Space object, as
// WatchFlowgraph follows the bindings of a graph: it wakes when the object
// appears, is retyped or is deleted, and when a plugin registers or releases
// its type. Only one goroutine calls its methods at a time.
type BindingWatch struct {
	// b is the bus the object type registry answers on.
	b bus.Bus
	// engine is the World engine holding the object.
	engine world.Engine
	// engineID scopes the registry lookup to the granting World engine.
	engineID string
	// objectKey is the watched object.
	objectKey string
	// watches holds the registry watch of the object's current type.
	watches *typeWatches
	// sent is the binding the last Next returned, or nil before the first.
	sent *FlowgraphBinding
}

// NewBindingWatch constructs a watch of the binding of objectKey. Call Close to
// stop its registry watch.
func NewBindingWatch(b bus.Bus, engine world.Engine, engineID, objectKey string) *BindingWatch {
	return &BindingWatch{
		b:         b,
		engine:    engine,
		engineID:  engineID,
		objectKey: objectKey,
		watches:   newTypeWatches(b, engineID),
	}
}

// Next returns the binding as it is now on the first call, and afterwards waits
// until the binding differs from the one the last call returned. It returns the
// context's error when ctx ends first, and a later call resumes the wait.
func (w *BindingWatch) Next(ctx context.Context) (*FlowgraphBinding, error) {
	for {
		// Read the binding and the object's root from one World state, so a
		// change after the read ends the wait below.
		binding, seen, err := w.read(ctx)
		if err != nil {
			return nil, err
		}

		// Return a binding the caller has not seen.
		if w.sent == nil || !binding.EqualVT(w.sent) {
			w.sent = binding
			return binding.CloneVT(), nil
		}

		// Watch the registry for the object's type, and read again after a new
		// watch since it saw no change before it began.
		snapshot := &FlowgraphSnapshot{Bindings: map[string]*FlowgraphBinding{w.objectKey: binding}}
		added, err := w.watches.sync(ctx, snapshot)
		if err != nil {
			return nil, err
		}
		if added {
			continue
		}
		if err := waitGraphChange(ctx, world.NewEngineWorldState(w.engine, false), w.watches.wake, seen); err != nil {
			return nil, err
		}
	}
}

// read derives the binding and reads the object's root in one read transaction.
func (w *BindingWatch) read(ctx context.Context) (*FlowgraphBinding, []*world.ObjectRootRef, error) {
	// Open the read transaction.
	tx, err := w.engine.NewTransaction(ctx, false)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Discard()

	// Derive the binding and read the object's root from it.
	binding, err := readBinding(ctx, tx, w.b, w.engineID, w.objectKey)
	if err != nil {
		return nil, nil, err
	}
	seen, err := world.GetObjectRootRefsBatch(ctx, tx, []string{w.objectKey})
	if err != nil {
		return nil, nil, err
	}
	return binding, seen, nil
}

// Close stops the registry watch.
func (w *BindingWatch) Close() {
	w.watches.close()
}
