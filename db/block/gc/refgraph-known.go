package block_gc

import (
	"bytes"
	"context"
	"strings"

	"github.com/s4wave/spacewave/db/kvtx"
)

// seedBatchRecords bounds the records one seeding commit reads.
const seedBatchRecords = 8192

// SeedRefsKnown adds an edge to NodeRefsKnown from every block that an edge
// names. A graph written before those edges existed treated a block with any
// edge as having known refs, so seeding keeps its reads unchanged. That rule
// also counts a block stored untracked under a tracked owner as known; the
// repair walk, which decodes each block, corrects such a block.
//
// Seeding scans the forward records for their subjects, then the incoming
// records for their objects. Both scans meet their first node in key order,
// so the new records arrive in key order too, and each checkpoint rewrites
// only the index pages around them instead of pages across the whole index.
// Each commit seeds the nodes of up to seedBatchRecords records, and a rerun
// after a crash finds the edges already written.
func (g *RefGraph) SeedRefsKnown(ctx context.Context) error {
	for _, kind := range []byte{kindForward, kindIncoming} {
		for start := graphKey(kind); start != nil; {
			var err error
			start, err = g.seedRefsKnownBatch(ctx, kind, start)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// seedRefsKnownBatch seeds the first nodes of the records of kind at or after
// start in one commit, and returns the key to resume from, or nil at the end.
func (g *RefGraph) seedRefsKnownBatch(ctx context.Context, kind byte, start []byte) ([]byte, error) {
	var next []byte
	err := g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
		// Collect the block nodes that lead the next records.
		blocks, last, more, err := readFirstBlocks(ctx, tx, kind, start)
		if err != nil {
			return err
		}

		// Add the edge each block lacks, and resume after the last record
		// read.
		for _, node := range blocks {
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

// readFirstBlocks returns the distinct block nodes that lead up to
// seedBatchRecords records of kind at or after start, in key order, with the
// last key read and whether more records remain.
func readFirstBlocks(ctx context.Context, tx kvtx.Tx, kind byte, start []byte) ([]string, []byte, bool, error) {
	// Position the iterator at start.
	it := tx.Iterate(ctx, graphKey(kind), true, false)
	defer it.Close()
	if err := it.Seek(start); err != nil {
		return nil, nil, false, err
	}

	// Keep each record's first node when it is a block. Records sharing a
	// first node are adjacent, so comparing with the previous node removes
	// the repeats.
	var blocks []string
	var last []byte
	for read := 0; it.Valid() && read < seedBatchRecords; read++ {
		key := it.Key()
		node, _, err := readNode(key[1:])
		if err != nil {
			return nil, nil, false, err
		}
		if strings.HasPrefix(node, prefixBlock) && (len(blocks) == 0 || blocks[len(blocks)-1] != node) {
			blocks = append(blocks, node)
		}
		last = bytes.Clone(key)
		it.Next()
	}
	return blocks, last, it.Valid(), it.Err()
}
