package entitygraph_controller

import (
	"context"

	"github.com/aperturerobotics/entitygraph/entity"
	"github.com/aperturerobotics/entitygraph/store"

	"github.com/aperturerobotics/controllerbus/directive"
)

// collectResolver is a CollectEntityGraph resolver
type collectResolver struct {
	ctx         context.Context
	c           *Controller
	valCreateCh chan entity.Entity
	valRemoveCh chan entity.Entity
}

// newCollectResolver builds a new collectResolver
func newCollectResolver(ctx context.Context, c *Controller) *collectResolver {
	return &collectResolver{
		c:           c,
		ctx:         ctx,
		valCreateCh: make(chan entity.Entity, 5),
		valRemoveCh: make(chan entity.Entity, 5),
	}
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (r *collectResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	r.ctx = ctx
	emittedValues := make(map[entity.Entity]uint32)
	r.c.mtx.Lock()
	for _, val := range r.c.values {
		valID, ok := handler.AddValue(val)
		if ok {
			emittedValues[val] = valID
		}
	}
	r.c.handlers = append(r.c.handlers, r)
	r.c.mtx.Unlock()

	defer func() {
		for _, id := range emittedValues {
			handler.RemoveValue(id)
		}

		r.c.mtx.Lock()
		for i, h := range r.c.handlers {
			if h == r {
				a := r.c.handlers
				a[i] = a[len(a)-1]
				a[len(a)-1] = nil
				a = a[:len(a)-1]
				break
			}
		}
		r.c.mtx.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case ent := <-r.valCreateCh:
			if valID, ok := handler.AddValue(ent); ok {
				emittedValues[ent] = valID
			}
		case ent := <-r.valRemoveCh:
			valID, ok := emittedValues[ent]
			if ok {
				handler.RemoveValue(valID)
				delete(emittedValues, ent)
			}
		}
	}
}

// HandleEntityAdded handles a new entity being added to the store.
func (r *collectResolver) HandleEntityAdded(ent entity.Entity) {
	select {
	case <-r.ctx.Done():
		return
	case r.valCreateCh <- ent:
	}
}

// HandleEntityRemoved handles a entity being removed from the store.
func (r *collectResolver) HandleEntityRemoved(ent entity.Entity) {
	select {
	case <-r.ctx.Done():
		return
	case r.valRemoveCh <- ent:
	}
}

// _ is a type assertion
var _ directive.Resolver = ((*collectResolver)(nil))

// _ is a type assertion
var _ store.Handler = ((*collectResolver)(nil))
