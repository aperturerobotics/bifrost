package bifrost_entitygraph

import (
	"sync"

	"github.com/aperturerobotics/bifrost/node"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/entitygraph/entity"
)

// getNodeHandler handles the GetNode directive results
type getNodeHandler struct {
	c    *Controller
	mtx  sync.Mutex
	vals map[directive.Value]entity.Entity
}

// newGetNodeHandler constructs a getNodeHandler
func newGetNodeHandler(c *Controller) *getNodeHandler {
	return &getNodeHandler{c: c, vals: make(map[directive.Value]entity.Entity)}
}

// HandleValueAdded is called when a value is added to the directive.
func (h *getNodeHandler) HandleValueAdded(
	inst directive.Instance,
	val directive.Value,
) {
	nod, ok := val.(node.Node)
	if !ok {
		h.c.le.Warn("GetNode value was not a Node")
		return
	}

	nodObj := NewNodeEntity(nod.GetPeerID())
	h.mtx.Lock()
	_, exists := h.vals[val]
	if !exists {
		h.vals[val] = nodObj
	}
	h.mtx.Unlock()

	if !exists {
		h.c.store.AddEntityObj(nodObj)
	}
}

// HandleValueRemoved is called when a value is removed from the directive.
func (h *getNodeHandler) HandleValueRemoved(
	inst directive.Instance,
	val directive.Value,
) {
	h.mtx.Lock()
	ent, exists := h.vals[val]
	if exists {
		delete(h.vals, val)
	}
	h.mtx.Unlock()

	if exists {
		h.c.store.RemoveEntityObj(ent)
	}
}

// HandleInstanceDisposed is called when a directive instance is disposed.
// This will occur if Close() is called on the directive instance.
func (h *getNodeHandler) HandleInstanceDisposed(inst directive.Instance) {
	// noop
}

// _ is a type assertion
var _ directive.ReferenceHandler = ((*getNodeHandler)(nil))
