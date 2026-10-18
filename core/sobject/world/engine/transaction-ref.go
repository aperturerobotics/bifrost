package sobject_world_engine

import (
	"github.com/aperturerobotics/util/refcount"
	world_block "github.com/s4wave/spacewave/db/world/block"
)

// transactionRef retains the engine until its transaction commits or discards.
type transactionRef struct {
	// engine owns the reference count and its lock.
	engine *blkEngine
	// released prevents repeated cleanup from releasing another transaction's hold.
	// Guarded by engine.bcast.
	released bool
}

// acquireTransaction retains the engine unless its release has started.
func (w *blkEngine) acquireTransaction() (*transactionRef, error) {
	// Admit the transaction under the same lock Release uses to retire the engine.
	locked := w.bcast.Lock()
	defer locked.Unlock()
	if w.releasing {
		return nil, world_block.ErrEngineClosed
	}
	w.transactions++
	return &transactionRef{engine: w}, nil
}

// Release drops the transaction's hold once, after its underlying cleanup.
func (r *transactionRef) Release() {
	locked := r.engine.bcast.Lock()
	if !r.released {
		r.released = true
		r.engine.transactions--
		locked.Broadcast()
	}
	locked.Unlock()
}

// _ is a type assertion.
var _ refcount.RefLike = (*transactionRef)(nil)
