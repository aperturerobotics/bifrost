package entitygraph_controller

import (
	"sync"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/entitygraph/entity"
)

// establishLinkHandler handles EstablishLink values
type establishLinkHandler struct {
	// c is the controller
	c *Controller
	// ref is the reference
	ref directive.Reference
	// mtx guards vals
	mtx sync.Mutex
	// vals are values
	vals map[directive.Value]entity.Entity
}

// newEstablishLinkHandler constructs a new establishLinkHandler
func newEstablishLinkHandler(c *Controller) *establishLinkHandler {
	return &establishLinkHandler{c: c, vals: make(map[directive.Value]entity.Entity)}
}

// HandleValueAdded is called when a value is added to the directive.
func (e *establishLinkHandler) HandleValueAdded(inst directive.Instance, val directive.Value) {
	vl, ok := val.(link.Link)
	if !ok {
		e.c.le.Warn("EstablishLink value was not a Link")
		return
	}

	entObj := NewLinkEntity(vl)
	e.mtx.Lock()
	_, exists := e.vals[val]
	if !exists {
		e.vals[val] = entObj
	}
	e.mtx.Unlock()

	if !exists {
		e.c.store.AddEntityObj(entObj)
	}
}

// HandleValueRemoved is called when a value is removed from the directive.
func (e *establishLinkHandler) HandleValueRemoved(inst directive.Instance, val directive.Value) {
	e.mtx.Lock()
	ent, exists := e.vals[val]
	if exists {
		delete(e.vals, val)
	}
	e.mtx.Unlock()
	if exists {
		e.c.store.RemoveEntityObj(ent)
	}
}

// HandleInstanceDisposed is called when a directive instance is disposed.
// This will occur if Close() is called on the directive instance.
func (e *establishLinkHandler) HandleInstanceDisposed(inst directive.Instance) {
	if e.ref == nil {
		return
	}

	e.c.mtx.Lock()
	for i, ref := range e.c.cleanupRefs {
		if ref == e.ref {
			a := e.c.cleanupRefs
			a[i] = a[len(a)-1]
			a[len(a)-1] = nil
			a = a[:len(a)-1]
			break
		}
	}
	e.c.mtx.Unlock()
}

var _ directive.ReferenceHandler = ((*establishLinkHandler)(nil))
