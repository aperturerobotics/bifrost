package engine

import (
	"context"

	"github.com/s4wave/spacewave/db/kvtx"
	kvtx_prefixer "github.com/s4wave/spacewave/db/kvtx/prefixer"
)

// metadataPrefix separates public metadata from block, graph, and journal keys.
const metadataPrefix = 0x01

// metadataStore selects the public metadata revision for transaction validation.
type metadataStore struct {
	// engine owns the shared index, publication, and store lifetime.
	engine *Engine
}

// MetadataStore exposes metadata keys with their own conservative revision.
// Block publication and GC changes cannot invalidate an unchanged metadata view.
func (e *Engine) MetadataStore() kvtx.Store {
	return kvtx_prefixer.NewPrefixer(&metadataStore{engine: e}, []byte{metadataPrefix})
}

// NewTransaction opens a view whose conflicts are confined to metadata writes.
func (s *metadataStore) NewTransaction(ctx context.Context, write bool) (kvtx.Tx, error) {
	return s.engine.newTransaction(ctx, write, true)
}

// Execute satisfies the store lifecycle; the owning engine publishes all writes.
func (s *metadataStore) Execute(context.Context) error {
	return nil
}

// _ verifies the public metadata store contract.
var _ kvtx.Store = (*metadataStore)(nil)
