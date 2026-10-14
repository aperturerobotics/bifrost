package block_gc

import (
	"context"
	"slices"

	"github.com/s4wave/spacewave/db/block"
)

// GetStoredBlock reads a block and its outgoing block edges. The edges come
// from the ref graph, the journal edges it has not applied, and the edges this
// store has buffered but not flushed. Blocks are immutable, so an edge once
// recorded stays valid for the life of the block. The refs are known only when
// those sources hold the block's edge to NodeRefsKnown, which a tracked put
// writes with its other edges, or a re-put buffered the block's refs.
func (g *GCStoreOps) GetStoredBlock(ctx context.Context, ref *block.BlockRef) (*block.StoredBlock, error) {
	// Read the immutable block data before resolving its outgoing references.
	data, found, err := g.store.GetBlock(ctx, ref)
	if err != nil || !found {
		return nil, err
	}
	stored, err := g.getBlockEdges(ctx, ref, data)
	if err != nil || stored.RefsKnown {
		return stored, err
	}

	// A flush takes the buffered edges before it delivers them, so a block it
	// is tracking can read as untracked. Read again between flushes.
	release, err := g.flushMu.Lock(ctx)
	if err != nil {
		return nil, err
	}
	defer release()
	return g.getBlockEdges(ctx, ref, data)
}

// getBlockEdges resolves the outgoing edges of a stored block.
func (g *GCStoreOps) getBlockEdges(ctx context.Context, ref *block.BlockRef, data []byte) (*block.StoredBlock, error) {
	// Collect durable, journaled, and buffered references for the stored block.
	refs, known, err := g.getOutgoingBlockRefs(ctx, BlockIRI(ref))
	if err != nil {
		return nil, err
	}

	// A re-put carries the block's own refs before FlushPending records them.
	if !known {
		g.mu.Lock()
		for _, reput := range g.pendingReputs {
			if reput.ref.EqualsRef(ref) {
				refs, known = block.CloneBlockRefs(reput.targets), true
				break
			}
		}
		g.mu.Unlock()
	}
	return &block.StoredBlock{Data: data, Refs: refs, RefsKnown: known}, nil
}

// getOutgoingBlockRefs collects the block targets of node's outgoing edges,
// and reports whether one of the edges marks its refs known.
func (g *GCStoreOps) getOutgoingBlockRefs(ctx context.Context, node string) ([]*block.BlockRef, bool, error) {
	// Read the durable outgoing edges and include pending journal entries.
	targets, err := g.refGraph.GetOutgoingRefs(ctx, node)
	if err != nil {
		return nil, false, err
	}
	if g.wal != nil {
		journaled, err := g.wal.GetPendingOutgoingRefs(ctx, node)
		if err != nil {
			return nil, false, err
		}
		targets = append(targets, journaled...)
	}

	// Include edges buffered for the next flush, or left by a failed flush.
	g.mu.Lock()
	for _, edge := range g.pendingRefs {
		if edge.source == node {
			targets = append(targets, edge.target)
		}
	}
	for _, edge := range g.pendingAdds {
		if edge.Subject == node {
			targets = append(targets, edge.Object)
		}
	}
	g.mu.Unlock()

	// Keep one ref per block target; other nodes are not block edges.
	slices.Sort(targets)
	targets = slices.Compact(targets)
	refs := make([]*block.BlockRef, 0, len(targets))
	for _, target := range targets {
		if ref, ok := ParseBlockIRI(target); ok {
			refs = append(refs, ref)
		}
	}
	return refs, slices.Contains(targets, NodeRefsKnown), nil
}
