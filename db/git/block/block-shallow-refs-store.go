package git_block

import (
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/s4wave/spacewave/db/block"
)

// NewShallowRefsStore constructs a shallow refs store from a list of hashes.
func NewShallowRefsStore(hashes []plumbing.Hash) (*ShallowRefsStore, error) {
	hashSet, err := NewHashSet(hashes)
	if err != nil {
		return nil, err
	}
	return &ShallowRefsStore{ShallowRefs: hashSet}, nil
}

// NewShallowRefsStoreBlock builds a new repo references block.
func NewShallowRefsStoreBlock() block.Block {
	return &ShallowRefsStore{}
}

// _ is a type assertion
var _ block.Block = (*ShallowRefsStore)(nil)
