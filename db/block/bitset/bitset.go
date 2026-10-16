package bitset

import "github.com/s4wave/spacewave/db/block"

// NewBitset constructs a new bitset from existing words.
func NewBitset(words []uint64, length uint32) *BitSet {
	if len(words) == 0 {
		return nil
	}
	return &BitSet{
		Len: length,
		Set: words,
	}
}

// Clone clones the bitset block.
func (b *BitSet) Clone() *BitSet {
	// Preserve an absent bitset when cloning.
	if b == nil {
		return nil
	}

	// Copy the bitset words so the clone can change independently.
	set := b.GetSet()
	bs := make([]uint64, len(set))
	copy(bs, set)

	return &BitSet{
		Set: bs,
		Len: b.GetLen(),
	}
}

// _ is a type assertion
var _ block.Block = (*BitSet)(nil)
