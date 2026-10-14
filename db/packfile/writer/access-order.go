package writer

import (
	"context"
	"slices"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/net/hash"
)

// ReadStoredBlock returns a block with its refs, or nil when the store lacks
// it.
type ReadStoredBlock func(ctx context.Context, ref *block.BlockRef) (*block.StoredBlock, error)

// WalkAccessOrder returns a BlockIterator over the blocks reachable from roots
// in the order a reader typically fetches them. Each block's direct refs are
// yielded together before any of their own refs, then the walk descends into
// each in turn. A directory's child nodes therefore sit side by side, followed
// by the contents of each child, so a read-ahead window over a pack carries
// the blocks a listing or a file read asks for next.
//
// Each block is yielded once. A block read returns nil for is skipped with
// its refs, so a walk over a partial store yields the part it holds.
func WalkAccessOrder(ctx context.Context, roots []*block.BlockRef, read ReadStoredBlock) BlockIterator {
	// pending holds the read blocks not yet yielded. stack holds the ref lists
	// still to expand, the next one on top, so it keeps refs and not data.
	type readBlock struct {
		h      *hash.Hash
		stored *block.StoredBlock
	}
	var pending []readBlock
	stack := [][]*block.BlockRef{roots}
	seen := make(map[string]struct{})

	// expand reads the unseen refs of one list, queues them for yielding and
	// pushes their own ref lists, the first child's on top.
	expand := func(refs []*block.BlockRef) error {
		// Collect the ref lists of the new blocks in order.
		var children [][]*block.BlockRef
		for _, ref := range refs {
			// Read each new ref once.
			key := ref.GetHash().MarshalString()
			if _, ok := seen[key]; ok || ref.GetEmpty() {
				continue
			}
			seen[key] = struct{}{}
			stored, err := read(ctx, ref)
			if err != nil {
				return err
			}
			if stored == nil {
				continue
			}

			// Queue it and remember its refs.
			pending = append(pending, readBlock{h: ref.GetHash(), stored: stored})
			if len(stored.Refs) != 0 {
				children = append(children, stored.Refs)
			}
		}

		// Push them so the first child expands next.
		slices.Reverse(children)
		stack = append(stack, children...)
		return nil
	}

	return func() (*hash.Hash, *block.StoredBlock, error) {
		// Expand ref lists until a block is ready or the walk ends.
		for len(pending) == 0 {
			if len(stack) == 0 {
				return nil, nil, nil
			}
			refs := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if err := expand(refs); err != nil {
				return nil, nil, err
			}
		}

		// Yield the oldest queued block.
		next := pending[0]
		pending[0] = readBlock{}
		pending = pending[1:]
		return next.h, next.stored, nil
	}
}
