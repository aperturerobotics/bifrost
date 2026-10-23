package mysql_devserver

import (
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
	sql_mysql "github.com/s4wave/spacewave/db/sql/mysql"
)

// memoryRoot holds a SQL root in a private in-memory block store.
type memoryRoot struct {
	// store holds the database's blocks.
	store block.StoreOps
	// ref identifies the committed SQL root, nil before the first commit.
	ref *block.BlockRef
}

// BuildTransaction opens a block transaction at the committed root.
func (r *memoryRoot) BuildTransaction(opts *block.PutOpts) (*block.Transaction, *block.Cursor) {
	return block.NewTransaction(r.store, nil, r.ref, opts)
}

// GetRef returns the committed SQL object reference.
func (r *memoryRoot) GetRef() *bucket.ObjectRef {
	return &bucket.ObjectRef{RootRef: r.ref.Clone()}
}

// SetRootRef advances the committed SQL root.
func (r *memoryRoot) SetRootRef(ref *block.BlockRef) {
	r.ref = ref
}

// _ is a type assertion.
var _ sql_mysql.RootCursor = (*memoryRoot)(nil)
