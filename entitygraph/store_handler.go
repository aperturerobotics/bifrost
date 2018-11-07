package bifrost_entitygraph

import (
	"github.com/aperturerobotics/entitygraph/entity"
	"github.com/aperturerobotics/entitygraph/store"
)

// storeHandler is the entity graph store handler
type storeHandler struct {
	c *Controller
}

// newStoreHandler constructs a new storeHandler
func newStoreHandler(c *Controller) *storeHandler {
	return &storeHandler{c: c}
}

// HandleEntityAdded handles a new entity being added to the store.
func (s *storeHandler) HandleEntityAdded(ent entity.Entity) {
	s.c.mtx.Lock()
	defer s.c.mtx.Unlock()

	handlers := s.c.handlers
	s.c.values[store.NewEntityMapKey(ent)] = ent
	for _, h := range handlers {
		h.HandleEntityAdded(ent)
	}
}

// HandleEntityRemoved handles a entity being removed from the store.
func (s *storeHandler) HandleEntityRemoved(ent entity.Entity) {
	s.c.mtx.Lock()
	defer s.c.mtx.Unlock()

	handlers := s.c.handlers
	delete(s.c.values, store.NewEntityMapKey(ent))
	for _, h := range handlers {
		h.HandleEntityRemoved(ent)
	}
}

// _ is a type assertion
var _ store.Handler = ((*storeHandler)(nil))
