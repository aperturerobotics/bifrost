package block_gc

import (
	"bytes"
	"context"
	"maps"
	"slices"
	"strings"

	"github.com/s4wave/spacewave/db/kvtx"
)

// seedBatchRecords bounds the forward records one seeding commit reads.
const seedBatchRecords = 8192

// SeedRefsKnown adds an edge to NodeRefsKnown from every block that an edge
// names. A graph written before those edges existed treated a block with any
// edge as having known refs, so seeding keeps its reads unchanged. That rule
// also counts a block stored untracked under a tracked owner as known; the
// repair walk, which decodes each block, corrects such a block. Each commit
// seeds the blocks of up to seedBatchRecords forward records, and a rerun
// after a crash finds the edges already written.
func (g *RefGraph) SeedRefsKnown(ctx context.Context) error {
	for start := graphKey(kindForward); start != nil; {
		var err error
		start, err = g.seedRefsKnownBatch(ctx, start)
		if err != nil {
			return err
		}
	}
	return nil
}

// seedRefsKnownBatch seeds the blocks of the forward records at or after
// start in one commit, and returns the key to resume from, or nil at the end.
func (g *RefGraph) seedRefsKnownBatch(ctx context.Context, start []byte) ([]byte, error) {
	var next []byte
	err := g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
		// Collect the block nodes of the next forward records.
		blocks, last, more, err := readForwardBlocks(ctx, tx, start)
		if err != nil {
			return err
		}

		// Add the edge each block lacks, in key order, and resume after the
		// last record read.
		for _, node := range slices.Sorted(maps.Keys(blocks)) {
			if err := g.addRefTx(ctx, tx, node, NodeRefsKnown); err != nil {
				return err
			}
		}
		next = nil
		if more {
			next = append(last, 0)
		}
		return nil
	})
	return next, err
}

// readForwardBlocks returns the block nodes of up to seedBatchRecords forward
// records at or after start, the last key read, and whether more remain.
func readForwardBlocks(ctx context.Context, tx kvtx.Tx, start []byte) (map[string]struct{}, []byte, bool, error) {
	// Position the iterator at start.
	it := tx.Iterate(ctx, graphKey(kindForward), true, false)
	defer it.Close()
	if err := it.Seek(start); err != nil {
		return nil, nil, false, err
	}

	// Decode both nodes of each record and keep the blocks.
	blocks := make(map[string]struct{})
	var last []byte
	for read := 0; it.Valid() && read < seedBatchRecords; read++ {
		key := it.Key()
		subject, rest, err := readNode(key[1:])
		if err != nil {
			return nil, nil, false, err
		}
		object, _, err := readNode(rest)
		if err != nil {
			return nil, nil, false, err
		}
		for _, node := range []string{subject, object} {
			if strings.HasPrefix(node, prefixBlock) {
				blocks[node] = struct{}{}
			}
		}
		last = bytes.Clone(key)
		it.Next()
	}
	return blocks, last, it.Valid(), it.Err()
}
