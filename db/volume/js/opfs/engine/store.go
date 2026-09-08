package engine

import (
	"context"

	"github.com/s4wave/spacewave/db/kvtx"
)

// NewTransaction opens a lazy generation-consistent transaction.
func (e *Engine) NewTransaction(ctx context.Context, write bool) (kvtx.Tx, error) {
	return e.newTransaction(ctx, write, false)
}

// newTransaction fixes the revision domain before observing committed records.
func (e *Engine) newTransaction(ctx context.Context, write, metadata bool) (kvtx.Tx, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	e.mtx.Lock()
	closed := e.closed
	e.mtx.Unlock()
	if closed {
		return nil, ErrClosed
	}
	return &transaction{engine: e, write: write, metadata: metadata, pending: make(map[string]*Record)}, nil
}

// Execute satisfies the durable volume store lifecycle; writes are synchronous.
func (e *Engine) Execute(ctx context.Context) error {
	return nil
}

// _ verifies the public transactional store contract.
var _ kvtx.Store = (*Engine)(nil)
