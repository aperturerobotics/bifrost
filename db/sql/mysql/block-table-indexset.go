package mysql

import (
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/block/sbset"
)

// tableIndexSet exposes a table's secondary indexes as sub-blocks.
type tableIndexSet struct {
	// root contains the index definitions and trees.
	root *TableRoot
}

// newTableIndexSet constructs the index sub-block container.
func newTableIndexSet(root *TableRoot) *sbset.SubBlockSet {
	return sbset.NewSubBlockSet(&tableIndexSet{root: root}, nil)
}

// Get returns the index at i, or nil outside the set.
func (s *tableIndexSet) Get(i int) block.SubBlock {
	if i < 0 || i >= s.Len() {
		return nil
	}
	return s.root.Indexes[i]
}

// Len returns the number of indexes.
func (s *tableIndexSet) Len() int {
	return len(s.root.Indexes)
}

// Set replaces an index sub-block.
func (s *tableIndexSet) Set(i int, value block.SubBlock) {
	if i >= 0 && i < s.Len() {
		if index, ok := value.(*TableIndex); ok {
			s.root.Indexes[i] = index
		}
	}
}

// Truncate removes trailing index sub-blocks.
func (s *tableIndexSet) Truncate(length int) {
	if length >= 0 && length < s.Len() {
		s.root.Indexes = s.root.Indexes[:length]
	}
}

// _ is a type assertion.
var _ sbset.SubBlockContainer = (*tableIndexSet)(nil)
