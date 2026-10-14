package s4db

import (
	"unsafe"
)

// entryOverhead is the memory an overlay entry holds besides its key and
// inline value: the entry and its item in the overlay tree.
const entryOverhead = int64(unsafe.Sizeof(oentry{}) + unsafe.Sizeof(overlayItem{}))

// oentry is one change in the overlay.
type oentry struct {
	// key is the key.
	key []byte
	// del marks a deleted key.
	del bool
	// val is the stored value.
	val value
	// seq is the commit that made the change.
	seq uint64
}

// size estimates the memory an overlay entry holds. Small keys cost more
// in their entry than in their bytes, so the checkpoint limits count this
// and not the encoded size.
func (e *oentry) size() int64 {
	return entryOverhead + int64(len(e.key)+len(e.val.inline))
}
