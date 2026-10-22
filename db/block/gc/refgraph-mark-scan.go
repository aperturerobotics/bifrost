package block_gc

import (
	"context"
	"errors"
	"hash/maphash"
	"slices"
	"time"

	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/s4wave/spacewave/net/hash"
)

// MarkReadBudget bounds how long MarkBlocksScan holds one read transaction.
// While a transaction is open the store cannot reuse pages freed after it
// began, so a short one keeps freed space reusable during a pass.
const MarkReadBudget = 100 * time.Millisecond

// markBudgetCheck is how many records a scan visits between clock reads.
const markBudgetCheck = 64

// markLimit bounds one read transaction of a scan mark: at most records
// records when nonzero, and at most budget when nonzero.
type markLimit struct {
	records int
	budget  time.Duration
}

// MarkBlocksScan marks like MarkBlocks, but reads the child lists in key order
// in read transactions of at most MarkReadBudget each, so a mark costs about
// one read per index page of the lists instead of one per block, and holds no
// snapshot for long. A block pass must be open, as BeginBlockPass describes:
// its adopt log lets the sweep's rescue keep every block a write adopts
// between the mark's transactions.
//
// It never falls back to MarkBlocks to name the reached blocks without a
// child list: that walk would hold one snapshot for the whole graph.
func (g *RefGraph) MarkBlocksScan(ctx context.Context) (*BlockMark, error) {
	return g.markScan(ctx, markLimit{budget: MarkReadBudget}, nil)
}

// markScan marks like MarkBlocks, but reads the child lists in key order:
// each scan expands every list whose block is live and was not yet expanded,
// and the scans repeat until one adds no child behind its position, since it
// expanded every child it added after its position. A scan reads each index
// page of the lists once, so a mark costs scans times pages instead of a page
// a block. It counts the blocks without a list in mark.Unknown and names a
// sample of them, as sampleUnknown describes.
//
// Each read transaction stays within limit, and the next resumes after the
// last record read, calling between first when set. The adopt log of an open
// pass makes that safe: a write that adds an edge to a block after the mark
// read the edge's source logs the block, and the sweep rescues it. Seeding
// takes both ends of every edge from a node that is not a block, so a seed
// read early misses only edges added later, which the log covers.
func (g *RefGraph) markScan(ctx context.Context, limit markLimit, between func() error) (*BlockMark, error) {
	// Mark seeds live, counting the blocks among them.
	mark := &BlockMark{Live: NewNodeSet()}
	live := mark.Live
	var liveBlocks int
	seed := func(node string) {
		if node != NodeUnreferenced && live.Add(node) && isBlockNode(node) {
			liveBlocks++
		}
	}

	// Seed the graph root and the registered roots.
	seed(NodeGCRoot)
	roots := graphKey(kindRoot)
	err := g.scanRecords(ctx, mark, roots, limit, between, func(key []byte, _ kvtx.Iterator) error {
		seed(lastNode(roots, key))
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Seed both ends of each edge from a text node.
	owners := []byte{kindForward, textNode}
	err = g.scanRecords(ctx, mark, owners, limit, between, func(key []byte, _ kvtx.Iterator) error {
		// Decode the subject and the object.
		subject, rest, err := readNode(key[1:])
		if err != nil {
			return err
		}
		object, _, err := readNode(rest)
		if err != nil {
			return err
		}

		// Seed them.
		seed(subject)
		seed(object)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Scan until a scan adds no child behind its position.
	expanded := NewNodeSet()
	lists := graphKey(kindChildren)
	for {
		mark.Scans++
		var behind int
		err := g.scanRecords(ctx, mark, lists, limit, between, func(key []byte, it kvtx.Iterator) error {
			// Expand a live list once, reading its value only then.
			enc := key[len(lists):]
			if !live.hasBlock(enc) || !expanded.addBlock(enc) {
				return nil
			}
			value, err := it.Value()
			if err != nil {
				return err
			}

			// Add its children, counting those behind the position, which
			// only a later scan can expand.
			for len(value) != 0 {
				n := blockNodeLen(value)
				if n == 0 || n > len(value) {
					return errChildNotBlock
				}
				if live.addBlock(value[:n]) {
					liveBlocks++
					if string(value[:n]) < string(enc) {
						behind++
					}
				}
				value = value[n:]
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if behind == 0 {
			break
		}
	}

	// A live block without a list has unknown references. List the
	// unreached blocks with one more key scan.
	mark.Unknown = liveBlocks - expanded.Len()
	if mark.Unknown != 0 {
		if err := g.sampleUnknown(ctx, mark, expanded, limit, between); err != nil {
			return nil, err
		}
	}
	err = g.scanRecords(ctx, mark, lists, limit, between, func(key []byte, _ kvtx.Iterator) error {
		if block := lastNode(lists, key); !live.Has(block) {
			mark.Unreached = append(mark.Unreached, block)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mark, nil
}

// errSampleFull stops sampleUnknown's scans once the sample is full.
var errSampleFull = errors.New("unknown sample full")

// sampleUnknown names up to unknownSample of the live blocks a scan mark left
// unexpanded, which have no child list. Only the records that made them live
// name them, so it scans the roots, the edges from text nodes and the
// expanded lists, within limit like the mark, and stops once the sample is
// full.
func (g *RefGraph) sampleUnknown(ctx context.Context, mark *BlockMark, expanded *NodeSet, limit markLimit, between func() error) error {
	// Note each unexpanded block once.
	noted := NewNodeSet()
	note := func(enc []byte) error {
		if len(enc) == 0 || enc[0] == textNode || expanded.hasBlock(enc) || !noted.addBlock(enc) {
			return nil
		}
		mark.UnknownSample = append(mark.UnknownSample, prefixBlock+string(enc))
		if len(mark.UnknownSample) == unknownSample {
			return errSampleFull
		}
		return nil
	}

	// Scan the roots, the objects of edges from text nodes, then the
	// children of the expanded lists.
	roots, owners, lists := graphKey(kindRoot), []byte{kindForward, textNode}, graphKey(kindChildren)
	scans := []struct {
		prefix []byte
		visit  func(key []byte, it kvtx.Iterator) error
	}{
		{roots, func(key []byte, _ kvtx.Iterator) error { return note(key[len(roots):]) }},
		{owners, func(key []byte, _ kvtx.Iterator) error {
			_, rest, err := readNode(key[1:])
			if err != nil {
				return err
			}
			if n := blockNodeLen(rest); n != 0 && n <= len(rest) {
				return note(rest[:n])
			}
			return nil
		}},
		{lists, func(key []byte, it kvtx.Iterator) error {
			// Read only the lists the mark expanded.
			if !expanded.hasBlock(key[len(lists):]) {
				return nil
			}
			value, err := it.Value()
			if err != nil {
				return err
			}
			for len(value) != 0 {
				n := blockNodeLen(value)
				if n == 0 || n > len(value) {
					return errChildNotBlock
				}
				if err := note(value[:n]); err != nil {
					return err
				}
				value = value[n:]
			}
			return nil
		}},
	}
	for _, scan := range scans {
		err := g.scanRecords(ctx, mark, scan.prefix, limit, between, scan.visit)
		if errors.Is(err, errSampleFull) {
			return nil
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// scanRecords calls visit with each record under prefix in key order, in read
// transactions within limit, calling between before each transaction after
// the first. It records the longest transaction in mark.LongestRead.
func (g *RefGraph) scanRecords(ctx context.Context, mark *BlockMark, prefix []byte, limit markLimit, between func() error, visit func(key []byte, it kvtx.Iterator) error) error {
	var resume []byte
	for {
		// Let the caller act between transactions.
		if resume != nil && between != nil {
			if err := between(); err != nil {
				return err
			}
		}

		// Visit records after the last one visited until the limit.
		var more bool
		err := g.transaction(ctx, false, func(ctx context.Context, tx kvtx.Tx) error {
			// Position after the last record visited.
			start := time.Now()
			defer func() { mark.LongestRead = max(mark.LongestRead, time.Since(start)) }()
			it := tx.Iterate(ctx, prefix, true, false)
			defer it.Close()
			valid := it.Next()
			if resume != nil {
				if err := it.Seek(append(slices.Clone(resume), 0)); err != nil {
					return err
				}
				valid = it.Valid()
			}

			// Visit the records, stopping at the limit.
			for n := 0; valid; n++ {
				if n != 0 && limitReached(limit, n, start) {
					more = true
					break
				}
				key := it.Key()
				if err := visit(key, it); err != nil {
					return err
				}
				resume = append(resume[:0], key...)
				valid = it.Next()
			}
			return it.Err()
		})
		if err != nil || !more {
			return err
		}
	}
}

// limitReached reports whether a transaction that began at start and visited
// n records has reached limit. It reads the clock every markBudgetCheck
// records.
func limitReached(limit markLimit, n int, start time.Time) bool {
	if limit.records != 0 && n >= limit.records {
		return true
	}
	return limit.budget != 0 && n%markBudgetCheck == 0 && time.Since(start) >= limit.budget
}

// blockNodeLen returns the length of the block encoding that b starts with,
// or zero when b does not start with one.
func blockNodeLen(b []byte) int {
	if len(b) == 0 || b[0] == textNode {
		return 0
	}
	size := hash.HashType(b[0]).GetHashLen()
	if size <= 0 {
		return 0
	}
	return 1 + size
}

// blockHash returns the hash Has and Add give the block IRI whose encoding
// is enc, without building the IRI.
func (s *NodeSet) blockHash(enc []byte) [2]uint64 {
	h := [2]uint64{maphash.Bytes(s.blockSeeds[0], enc), maphash.Bytes(s.blockSeeds[1], enc)}
	if h == ([2]uint64{}) {
		h[1] = 1
	}
	return h
}

// hasBlock reports whether the block encoded as enc is in the set.
func (s *NodeSet) hasBlock(enc []byte) bool {
	_, found := s.find(s.blockHash(enc))
	return found
}

// addBlock inserts the block encoded as enc and reports whether it was
// absent.
func (s *NodeSet) addBlock(enc []byte) bool {
	return s.addHash(s.blockHash(enc))
}
