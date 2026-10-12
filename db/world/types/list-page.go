package world_types

import (
	"context"
	"io"

	cayley_kv "github.com/aperturerobotics/cayley/graph/kv"
	"github.com/aperturerobotics/cayley/quad"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/world"
)

// ObjectTypePageLimit bounds typed pages, including the default traversal page.
const ObjectTypePageLimit uint32 = 1000

// ListObjectsWithTypeOnGraph returns distinct subjects in graph-index order.
// startAfter is the previous page's last key; empty begins the listing.
// limit must be non-zero and is capped at ObjectTypePageLimit. The World must
// remain unchanged to retain exactly-once traversal across pages.
func ListObjectsWithTypeOnGraph(ctx context.Context, h *cayley_kv.QuadStore, typeID, startAfter string, limit uint32) ([]string, bool, error) {
	// Validate the page before reading the graph.
	if typeID == "" {
		return nil, false, ErrTypeIDEmpty
	}
	if limit == 0 {
		return nil, false, errors.New("limit must be non-zero")
	}
	limit = min(limit, ObjectTypePageLimit)

	// Resolve the type and predicate to the index's durable node IDs.
	typeRef, err := h.ValueOf(ctx, BuildTypeQuadValue(typeID))
	if err != nil || typeRef == nil {
		return nil, false, err
	}
	predRef, err := h.ValueOf(ctx, TypePred)
	if err != nil || predRef == nil {
		return nil, false, err
	}

	// Resolve the cursor to its subject node before seeking the reverse index.
	var after cayley_kv.Int64Value
	if startAfter != "" {
		ref, err := h.ValueOf(ctx, world.KeyToGraphValue(startAfter))
		if err != nil {
			return nil, false, err
		}
		if ref == nil {
			return nil, false, errors.New("start_after has no graph node")
		}
		after = ref.(cayley_kv.Int64Value)
	}

	// Read at most one page plus its lookahead, deduplicated by subject posting.
	keys := make([]string, 0, limit)
	var more bool
	values := []uint64{uint64(typeRef.(cayley_kv.Int64Value)), uint64(predRef.(cayley_kv.Int64Value))}
	err = h.IterateIndexPrefixNextRefsAfter(ctx, cayley_kv.DefaultQuadIndexes[1], values, after, func(ref cayley_kv.Int64Value, hasLive func() (bool, error)) error {
		// Ignore a posting whose quads have all been deleted.
		live, err := hasLive()
		if err != nil || !live {
			return err
		}

		// Detect a following page without decoding its first subject.
		if uint32(len(keys)) == limit {
			more = true
			return io.EOF
		}

		// Decode the subject directly from its node reference.
		value, err := h.NameOf(ctx, ref)
		if err != nil {
			return err
		}
		key, ok := value.(quad.IRI)
		if !ok {
			return errors.New("type subject is not an object key")
		}
		keys = append(keys, string(key))
		return nil
	})
	return keys, more, err
}
