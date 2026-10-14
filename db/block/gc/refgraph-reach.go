package block_gc

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/s4wave/spacewave/net/hash"
)

// reachChunk bounds the nodes whose forward records one prefetch loads.
const reachChunk = 4096

// Reachable returns every node reachable by forward edges from the graph root
// and the registered roots, the roots included: the nodes the mark phase of a
// complete collection keeps. The unreferenced queue marks candidates and owns
// nothing, so neither it nor its edges are included. The walk reads one
// snapshot a level at a time, loading the index pages of each chunk of the
// level's records together in key order, so a cold store costs a batch of
// reads per chunk instead of one read per node.
func (g *RefGraph) Reachable(ctx context.Context) (map[string]struct{}, error) {
	var reached map[string]struct{}
	err := g.transaction(ctx, false, func(ctx context.Context, tx kvtx.Tx) error {
		// Seed the first level with the graph root and the registered roots.
		reached = map[string]struct{}{NodeGCRoot: {}}
		level := []string{NodeGCRoot}
		prefix := graphKey(kindRoot)
		err := tx.ScanPrefixKeys(ctx, prefix, func(key []byte) error {
			node := lastNode(prefix, key)
			if _, ok := reached[node]; !ok && node != NodeUnreferenced {
				reached[node] = struct{}{}
				level = append(level, node)
			}
			return nil
		})
		if err != nil {
			return err
		}

		// Walk the levels until one adds no node.
		for len(level) != 0 {
			level, err = reachLevel(ctx, tx, level, reached)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return reached, err
}

// reachLevel adds each target of the forward records of level that is not
// yet in reached, and returns those targets as the next level.
func reachLevel(ctx context.Context, tx kvtx.Tx, level []string, reached map[string]struct{}) ([]string, error) {
	// Order the level's record prefixes by key.
	prefixes := make([][]byte, 0, len(level))
	for _, node := range level {
		if node != NodeUnreferenced {
			prefixes = append(prefixes, graphKey(kindForward, node))
		}
	}
	slices.SortFunc(prefixes, bytes.Compare)

	// Load each chunk's index pages together, then scan its records.
	var next []string
	for chunk := range slices.Chunk(prefixes, reachChunk) {
		if err := kvtx.Prefetch(ctx, tx, chunk); err != nil {
			return nil, err
		}
		for _, prefix := range chunk {
			err := tx.ScanPrefixKeys(ctx, prefix, func(key []byte) error {
				node := lastNode(prefix, key)
				if _, ok := reached[node]; !ok {
					reached[node] = struct{}{}
					next = append(next, node)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return next, nil
}

// RecordSubject returns the node whose liveness decides whether a complete
// collection keeps the graph record key: the subject of an edge in either
// direction, or the node of a node or root record.
func RecordSubject(key []byte) (string, error) {
	// Split the kind byte from the nodes.
	if len(key) == 0 {
		return "", fmt.Errorf("decode empty reference graph record")
	}
	kind, rest := key[0], key[1:]

	// An incoming record names the object first.
	if kind == kindIncoming {
		var err error
		if _, rest, err = readNode(rest); err != nil {
			return "", fmt.Errorf("decode reference graph record %q: %w", key, err)
		}
	}
	node, _, err := readNode(rest)
	if err != nil {
		return "", fmt.Errorf("decode reference graph record %q: %w", key, err)
	}
	return node, nil
}

// readNode decodes the node at the start of enc, as appendNode wrote it, and
// returns the bytes after it.
func readNode(enc []byte) (string, []byte, error) {
	// Require the first byte of a node.
	if len(enc) == 0 {
		return "", nil, fmt.Errorf("missing node")
	}

	// A text node is a uvarint length and the IRI.
	if enc[0] == textNode {
		size, n := binary.Uvarint(enc[1:])
		if n <= 0 || uint64(len(enc)-1-n) < size { //nolint:gosec // the length is checked nonnegative first.
			return "", nil, fmt.Errorf("truncated text node")
		}
		end := 1 + n + int(size) //nolint:gosec // size is bounded by len(enc) above.
		return string(enc[1+n : end]), enc[end:], nil
	}

	// A block node is its hash type and a digest of that type's length.
	t := hash.HashType(enc[0])
	size := t.GetHashLen()
	if size <= 0 || len(enc)-1 < size {
		return "", nil, fmt.Errorf("truncated block node of hash type %d", enc[0])
	}
	end := 1 + size
	return prefixBlock + string(enc[:end]), enc[end:], nil
}
