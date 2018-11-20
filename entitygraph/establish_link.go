package bifrost_entitygraph

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
	vals map[directive.Value]establishLinkHandlerVal
}

// establishLinkHandlerVal is the value tuple
type establishLinkHandlerVal struct {
	linkObj, remoteNodObj           entity.Entity
	remoteTptObj, remoteTptAssocObj entity.Entity
}

// newEstablishLinkHandler constructs a new establishLinkHandler
func newEstablishLinkHandler(c *Controller) *establishLinkHandler {
	return &establishLinkHandler{c: c, vals: make(map[directive.Value]establishLinkHandlerVal)}
}

// HandleValueAdded is called when a value is added to the directive.
func (e *establishLinkHandler) HandleValueAdded(inst directive.Instance, val directive.Value) {
	vl, ok := val.(link.Link)
	if !ok {
		e.c.le.Warn("EstablishLink value was not a Link")
		return
	}

	entObj := NewLinkEntity(vl)
	nodObj := NewPeerEntity(vl.GetRemotePeer())
	remoteTptObj, remoteTptAssocObj := NewTransportEntity(vl.GetRemoteTransportUUID(), vl.GetRemotePeer())
	e.mtx.Lock()
	_, exists := e.vals[val]
	if !exists {
		e.vals[val] = establishLinkHandlerVal{
			remoteNodObj:      nodObj,
			linkObj:           entObj,
			remoteTptAssocObj: remoteTptAssocObj,
			remoteTptObj:      remoteTptObj,
		}
	}
	e.mtx.Unlock()

	if !exists {
		e.c.store.AddEntityObj(entObj)
		e.c.store.AddEntityObj(nodObj)
		e.c.store.AddEntityObj(remoteTptAssocObj)
		e.c.store.AddEntityObj(remoteTptObj)
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
		e.c.store.RemoveEntityObj(ent.linkObj)
		e.c.store.RemoveEntityObj(ent.remoteNodObj)
		e.c.store.RemoveEntityObj(ent.remoteTptAssocObj)
		e.c.store.RemoveEntityObj(ent.remoteTptObj)
	}
}

// HandleInstanceDisposed is called when a directive instance is disposed.
// This will occur if Close() is called on the directive instance.
func (e *establishLinkHandler) HandleInstanceDisposed(inst directive.Instance) {
	eref := e.ref
	if eref == nil {
		return
	}
	e.ref = nil

	e.mtx.Lock()
	for k, val := range e.vals {
		e.c.store.RemoveEntityObj(val.linkObj)
		e.c.store.RemoveEntityObj(val.remoteNodObj)
		delete(e.vals, k)
	}
	e.mtx.Unlock()

	e.c.mtx.Lock()
	for i, ref := range e.c.cleanupRefs {
		if ref == eref {
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
