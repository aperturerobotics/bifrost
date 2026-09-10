package bucket_lookup

import (
	"context"
	"slices"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
)

// WalkObjectGraph visits a block graph depth first with memory bounded by depth.
// visit may prune a previously completed subtree or supply a dynamic decoder.
// complete runs only after visit and every descendant succeed. A caller can
// durably record that boundary to avoid traversing immutable subtrees again.
// Callbacks are sequential and must not retain mutable entries after returning.
func WalkObjectGraph(
	ctx context.Context,
	root *WalkObjectBlocksEntry,
	visit WalkObjectBlocksCb,
	complete func(*WalkObjectBlocksEntry) error,
	store bucket.BucketOps,
	xfrm block.Transformer,
) error {
	if root == nil {
		return nil
	}
	var walk func(*WalkObjectBlocksEntry) error
	walk = func(entry *WalkObjectBlocksEntry) error {
		// Load one immutable node while retaining only its ancestors.
		if err := ctx.Err(); err != nil {
			return err
		}
		if !entry.Found && entry.Err == nil && !entry.IsSubBlock && !entry.Ref.GetEmpty() {
			entry.Data, entry.Found, entry.Err = store.GetBlock(ctx, entry.Ref)
		}
		if entry.Err == nil && entry.Found && !entry.IsSubBlock {
			entry.Err = entry.decodeBlock(false, xfrm)
		}
		if visit != nil {
			proceed, err := visit(entry)
			if err != nil || !proceed {
				return err
			}
		} else if entry.Err != nil {
			return entry.Err
		}

		// Resolve children only after the callback has supplied dynamic state.
		var children []*WalkObjectBlocksEntry
		if sub, ok := entry.Blk.(block.BlockWithSubBlocks); ok {
			for id, child := range sub.GetSubBlocks() {
				if child != nil && !child.IsNil() {
					children = append(children, &WalkObjectBlocksEntry{RefID: id, Ref: entry.Ref, Blk: child, IsSubBlock: true, Found: true})
				}
			}
		}
		if refs, ok := entry.Blk.(block.BlockWithRefs); ok {
			childrenRefs, err := refs.GetBlockRefs()
			if err != nil {
				return err
			}
			for id, ref := range childrenRefs {
				if !ref.GetEmpty() {
					children = append(children, &WalkObjectBlocksEntry{RefID: id, Ref: ref, Ctor: refs.GetBlockRefCtor(id)})
				}
			}
		}
		slices.SortStableFunc(children, func(a, b *WalkObjectBlocksEntry) int {
			if a.RefID < b.RefID {
				return -1
			}
			if a.RefID > b.RefID {
				return 1
			}
			return 0
		})

		// Complete a parent only after all referenced descendants completed.
		for _, child := range children {
			child.Depth = entry.Depth + 1
			if err := walk(child); err != nil {
				return err
			}
		}
		if complete != nil {
			return complete(entry)
		}
		return nil
	}
	return walk(root)
}
