package link_holdopen_controller

import (
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// establishLinkHandler handles EstablishLink values
type establishLinkHandler struct {
	// c is the controller
	c *Controller
	// le is the logger
	le *logrus.Entry
	// ref is the reference
	ref directive.Reference
	// peerID is the peer id in the directive
	peerID peer.ID
	// di is the directive instance
	di directive.Instance

	// valCount is the number of added values
	valCount int
	// rigidRef is the non-weak reference
	rigidRef directive.Reference
}

// newEstablishLinkHandler constructs a new establishLinkHandler
func newEstablishLinkHandler(
	c *Controller,
	le *logrus.Entry,
	di directive.Instance,
	peerID peer.ID,
) *establishLinkHandler {
	return &establishLinkHandler{
		c:      c,
		di:     di,
		le:     le.WithField("peer-id", peerID.Pretty()),
		peerID: peerID,
	}
}

// HandleValueAdded is called when a value is added to the directive.
func (e *establishLinkHandler) HandleValueAdded(inst directive.Instance, val directive.AttachedValue) {
	vl, ok := val.GetValue().(link.Link)
	if !ok || vl == nil {
		return
	}
	e.valCount++
	if e.rigidRef == nil {
		e.le.
			WithField("link-uuid", vl.GetUUID()).
			WithField("local-peer", vl.GetLocalPeer().Pretty()).
			Debug("starting peer hold-open tracking")
		go func() {
			e.rigidRef = e.di.AddReference(nil, false)
		}()
	}
}

// HandleValueRemoved is called when a value is removed from the directive.
func (e *establishLinkHandler) HandleValueRemoved(inst directive.Instance, val directive.AttachedValue) {
	if e.valCount > 0 {
		e.valCount--
	}
	if e.valCount == 0 && e.rigidRef != nil {
		e.rigidRef.Release()
		e.rigidRef = nil
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
	if e.rigidRef != nil {
		e.rigidRef.Release()
		e.rigidRef = nil
	}

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
