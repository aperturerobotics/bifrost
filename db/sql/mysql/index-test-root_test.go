package mysql

import (
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
)

// indexTestRoot stores a SQL root over a private in-memory block store.
type indexTestRoot struct {
	// store holds the test's blocks.
	store block.StoreOps
	// ref identifies the committed SQL root.
	ref *block.BlockRef
}

// BuildTransaction opens a block transaction at the committed root.
func (r *indexTestRoot) BuildTransaction(opts *block.PutOpts) (*block.Transaction, *block.Cursor) {
	return block.NewTransaction(r.store, nil, r.ref, opts)
}

// GetRef returns the committed SQL object reference.
func (r *indexTestRoot) GetRef() *bucket.ObjectRef {
	return &bucket.ObjectRef{RootRef: r.ref.Clone()}
}

// SetRootRef advances the committed SQL root.
func (r *indexTestRoot) SetRootRef(ref *block.BlockRef) {
	r.ref = ref
}

// _ is a type assertion.
var _ RootCursor = (*indexTestRoot)(nil)
