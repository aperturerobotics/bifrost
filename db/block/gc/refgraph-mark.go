package block_gc

import (
	"bytes"
	"context"
	"encoding/binary"
	"slices"
	"time"

	"github.com/s4wave/spacewave/db/kvtx"
)

// unknownSample bounds the blocks of unknown references a mark reports.
const unknownSample = 16

// BlockMark is the result of marking the blocks of a graph in one snapshot.
type BlockMark struct {
	// Live holds every node the mark reached. A rescue adds to it.
	Live *NodeSet
	// Unreached lists the tracked blocks the mark did not reach, in key order.
	Unreached []string
	// Unknown counts the reached blocks without a child list as last read,
	// whose references are not recorded. A sweep removes nothing while any
	// remains, since such a block may reference any unreached one.
	Unknown int
	// UnknownSample lists up to unknownSample of those blocks.
	UnknownSample []string
	// UnknownNamers lists, for each block in UnknownSample, the node whose
	// record made it live: a root, an owner or a parent block, or empty for
	// a block a write adopted during the pass.
	UnknownNamers []string
	// Scans counts the scans of the child lists a scan mark made, zero for
	// a walk.
	Scans int
	// LongestRead is the longest read transaction of a scan mark, zero for
	// a walk.
	LongestRead time.Duration
	// Roots counts the registered roots the mark seeded.
	Roots int
	// Owners counts, by kind, the nodes that are not blocks whose forward
	// records seeded the mark, and those records. A kind that grows from
	// pass to pass while the World's data does not is an owner never
	// released.
	Owners map[string]OwnerCount

	// awaiting lists the blocks Unknown counts, each with its namer, up to
	// settleLimit of them. A write can commit the edge that adopts a block
	// before the staging edge that creates its list, so the list may yet
	// appear, and each rescue reads it again. The bound keeps that read
	// within a rescue's writer transaction small.
	awaiting []namedNode
	// logSeen is the first adopt log record no rescue or catch-up walked.
	logSeen uint64
}

// namedNode is a reached node with the node whose record reached it.
type namedNode struct {
	node  string
	namer string
}

// Awaitable reports whether every unknown block is listed, so a sweep can
// wait for their child lists instead of giving up.
func (m *BlockMark) Awaitable() bool {
	return m.Unknown == len(m.awaiting)
}

// await lists a reached block without a child list, up to settleLimit, and
// samples it, without counting it.
func (m *BlockMark) await(b namedNode) {
	if len(m.awaiting) < settleLimit {
		m.awaiting = append(m.awaiting, b)
	}
	if len(m.UnknownSample) < unknownSample {
		m.UnknownSample = append(m.UnknownSample, b.node)
		m.UnknownNamers = append(m.UnknownNamers, b.namer)
	}
}

// noteUnknown counts, lists and samples a reached block without a child list.
func (m *BlockMark) noteUnknown(b namedNode) {
	m.Unknown++
	m.await(b)
}

// takeAwaiting returns the listed unknown blocks for a walk to read again and
// stops counting them, sampling afresh when every unknown block is listed.
func (m *BlockMark) takeAwaiting() []namedNode {
	// Hand the list over and uncount it.
	stack := m.awaiting
	if m.Awaitable() {
		m.UnknownSample, m.UnknownNamers = nil, nil
	}
	m.Unknown -= len(stack)
	m.awaiting = nil
	return stack
}

// MarkBlocks marks the blocks of the graph in one snapshot. Every node that is
// not a block keeps the reference count and its queue, so the mark treats each
// one as live along with the graph root and the registered roots, and walks
// child lists and forward edges from the blocks they own. The tracked blocks
// it does not reach are the sweep's candidates. Opening a pass with
// BeginBlockPass before the mark lets a sweep rescue the blocks writes adopt
// after the snapshot.
func (g *RefGraph) MarkBlocks(ctx context.Context) (*BlockMark, error) {
	var mark *BlockMark
	err := g.transaction(ctx, false, func(ctx context.Context, tx kvtx.Tx) error {
		// Mark the seeds, stacking the blocks to walk.
		mark = &BlockMark{Live: NewNodeSet()}
		live := mark.Live
		var stack []namedNode
		err := seedMarkTx(ctx, tx, mark, func(node, namer string) {
			if node != NodeUnreferenced && live.Add(node) && isBlockNode(node) {
				stack = append(stack, namedNode{node, namer})
			}
		})
		if err != nil {
			return err
		}

		// Walk the owned blocks, noting those of unknown references.
		err = walkTx(ctx, tx, live, stack, mark.noteUnknown)
		if err != nil {
			return err
		}

		// List the tracked blocks the walk did not reach.
		lists := graphKey(kindChildren)
		return tx.ScanPrefixKeys(ctx, lists, func(key []byte) error {
			if block := lastNode(lists, key); !live.Has(block) {
				mark.Unreached = append(mark.Unreached, block)
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return mark, nil
}

// seedMarkTx calls seed with the graph root, each registered root, and both
// nodes of each forward record whose subject is not a block, each with the
// node that names it, counting the roots and owners in mark.
func seedMarkTx(ctx context.Context, tx kvtx.Tx, mark *BlockMark, seed func(node, namer string)) error {
	// Seed the graph root and the registered roots.
	seed(NodeGCRoot, "")
	roots := graphKey(kindRoot)
	err := tx.ScanPrefixKeys(ctx, roots, func(key []byte) error {
		mark.Roots++
		seed(lastNode(roots, key), NodeGCRoot)
		return nil
	})
	if err != nil {
		return err
	}

	// Seed both nodes of each edge from a text node.
	owners := []byte{kindForward, textNode}
	var last string
	return tx.ScanPrefixKeys(ctx, owners, func(key []byte) error {
		// Decode the subject and the object.
		subject, rest, err := readNode(key[1:])
		if err != nil {
			return err
		}
		object, _, err := readNode(rest)
		if err != nil {
			return err
		}

		// Count and seed them.
		mark.countOwner(subject, last)
		last = subject
		seed(subject, "")
		seed(object, subject)
		return nil
	})
}

// walkTx walks child lists and forward edges from the nodes on stack, adding
// each node it reaches to reached. untracked, when set, is called with each
// block it reaches that has no child list, named by the node whose record
// reached it. It pops chunks of nodes and loads the index pages of each
// chunk's records together in key order, so a cold store costs a batch of
// reads per chunk instead of one read per node, and the work list stays small
// however wide the graph is.
func walkTx(ctx context.Context, tx kvtx.Tx, reached *NodeSet, stack []namedNode, untracked func(namedNode)) error {
	// Stack each node on its first visit, with the node that reached it.
	visit := func(node, namer string) {
		if reached.Add(node) {
			stack = append(stack, namedNode{node, namer})
		}
	}

	// Walk chunks until the stack is empty.
	records := make([]walkRecord, 0, reachChunk)
	keys := make([][]byte, 0, reachChunk)
	for len(stack) != 0 {
		// Pop a chunk and order the keys of its records.
		chunk := stack[max(0, len(stack)-reachChunk):]
		records = records[:0]
		for _, n := range chunk {
			switch {
			case n.node == NodeUnreferenced:
			case isBlockNode(n.node):
				records = append(records, walkRecord{graphKey(kindChildren, n.node), n})
			default:
				records = append(records, walkRecord{graphKey(kindForward, n.node), n})
			}
		}
		stack = stack[:len(stack)-len(chunk)]
		slices.SortFunc(records, func(a, b walkRecord) int { return bytes.Compare(a.key, b.key) })

		// Load the chunk's index pages together, then push each new child
		// and target.
		keys = keys[:0]
		for _, r := range records {
			keys = append(keys, r.key)
		}
		if err := kvtx.Prefetch(ctx, tx, keys); err != nil {
			return err
		}
		for _, r := range records {
			if err := walkRecordsTx(ctx, tx, r, visit, untracked); err != nil {
				return err
			}
		}
	}
	return nil
}

// walkRecord is the key of the records of a node a walk reached.
type walkRecord struct {
	key  []byte
	node namedNode
}

// walkRecordsTx calls visit with each node that r names as a target, named by
// r's node: the children in a child list record, or the targets of the
// forward records under a forward prefix. untracked, when set, is called with
// the block of a missing child list.
func walkRecordsTx(ctx context.Context, tx kvtx.Tx, r walkRecord, visit func(node, namer string), untracked func(namedNode)) error {
	// Read the children of a child list.
	if r.key[0] == kindChildren {
		value, found, err := tx.Get(ctx, r.key)
		if err != nil {
			return err
		}
		if !found {
			if untracked != nil {
				untracked(r.node)
			}
			return nil
		}
		children, err := decodeChildren(value)
		if err != nil {
			return err
		}
		for _, child := range children {
			visit(child, r.node.node)
		}
		return nil
	}

	// Scan the targets of the forward records.
	return tx.ScanPrefixKeys(ctx, r.key, func(record []byte) error {
		visit(lastNode(r.key, record), r.node.node)
		return nil
	})
}

// passKey marks a running block pass. Its value is the pass's adopt log
// range, so no write scans the log: deleted records stay in the store until
// it compacts, and a scan steps past each one while it holds the writer.
var passKey = []byte{kindPass}

// adoptRange is the adopt log of a running pass: records lo up to hi.
type adoptRange struct {
	// lo is the first record not yet consumed.
	lo uint64
	// hi is the sequence of the next record.
	hi uint64
	// legacy marks a pass record written without a range, whose log only a
	// scan can find.
	legacy bool
}

// encode returns the pass record value of r.
func (r adoptRange) encode() []byte {
	return binary.BigEndian.AppendUint64(binary.BigEndian.AppendUint64(nil, r.lo), r.hi)
}

// readPassTx returns the adopt log range of the running pass, or false when
// no pass runs.
func readPassTx(ctx context.Context, tx kvtx.Tx) (adoptRange, bool, error) {
	// A pass record without a range predates the range.
	value, open, err := tx.Get(ctx, passKey)
	if err != nil || !open {
		return adoptRange{}, false, err
	}
	if len(value) != 16 {
		return adoptRange{legacy: true}, true, nil
	}
	r := adoptRange{lo: binary.BigEndian.Uint64(value), hi: binary.BigEndian.Uint64(value[8:])}
	if r.lo > r.hi {
		return adoptRange{legacy: true}, true, nil
	}
	return r, true, nil
}

// BeginBlockPass opens a block pass. From its commit until EndBlockPass, each
// write appends the blocks it adopts to the adopt log, so a sweep after the
// mark can rescue them with RescueAdopted. It first drops the log of an
// earlier pass that never ended. The caller runs one pass at a time per graph.
func (g *RefGraph) BeginBlockPass(ctx context.Context) error {
	return g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
		// The new log starts where the earlier one ended, so it never
		// rewrites a sequence the earlier pass used.
		r, open, err := readPassTx(ctx, tx)
		if err != nil {
			return err
		}
		next, err := deleteAdoptLogTx(ctx, tx, r, open)
		if err != nil {
			return err
		}
		return tx.Set(ctx, passKey, adoptRange{lo: next, hi: next}.encode())
	})
}

// EndBlockPass closes the block pass and deletes its adopt log, which holds
// only what writes logged since the last rescue.
func (g *RefGraph) EndBlockPass(ctx context.Context) error {
	return g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
		// Delete the log, then the marker.
		r, open, err := readPassTx(ctx, tx)
		if err != nil || !open {
			return err
		}
		if _, err := deleteAdoptLogTx(ctx, tx, r, open); err != nil {
			return err
		}
		return tx.Delete(ctx, passKey)
	})
}

// CatchUpAdopted walks, like RescueAdopted, from the unknown blocks the mark
// listed and from the blocks of the adopt log records no rescue or catch-up
// walked yet, in read transactions of at most MarkReadBudget each, and
// returns the number of nodes it added to mark.Live. It leaves the log for
// RescueAdopted to consume, which then walks only the records written since,
// so the walk a sweep transaction holds the writer for scales with the writes
// since the catch-up, not with the mark's length. Reading a list before the
// rescue is safe as the scan mark's later reads are: an edge added after the
// read logs its object.
func (g *RefGraph) CatchUpAdopted(ctx context.Context, mark *BlockMark) (int, error) {
	before := mark.Live.Len()
	stack := mark.takeAwaiting()
	for done := false; !done; {
		err := g.transaction(ctx, false, func(ctx context.Context, tx kvtx.Tx) (err error) {
			done, err = catchUpTx(ctx, tx, mark, stack, MarkReadBudget)
			return err
		})
		if err != nil {
			return 0, err
		}
		stack = nil
	}
	return mark.Live.Len() - before, nil
}

// catchUpTx walks from stack, then from the blocks of the unwalked adopt log
// records in order until budget, and reports whether it walked them all. A
// legacy log has no order to resume in, so it leaves that to RescueAdopted.
func catchUpTx(ctx context.Context, tx kvtx.Tx, mark *BlockMark, stack []namedNode, budget time.Duration) (bool, error) {
	// Read the log's range, then walk from the blocks read again.
	start := time.Now()
	r, open, err := readPassTx(ctx, tx)
	if err != nil {
		return false, err
	}
	if err := walkTx(ctx, tx, mark.Live, stack, mark.noteUnknown); err != nil {
		return false, err
	}
	if !open || r.legacy {
		return true, nil
	}

	// Walk from each record's blocks until the budget runs out.
	for seq := max(r.lo, mark.logSeen); seq < r.hi; seq++ {
		if time.Since(start) >= budget {
			return false, nil
		}
		blocks, err := readAdoptLogTx(ctx, tx, adoptRange{lo: seq, hi: seq + 1})
		if err != nil {
			return false, err
		}
		if err := walkTx(ctx, tx, mark.Live, adoptedStack(mark, blocks), mark.noteUnknown); err != nil {
			return false, err
		}
		mark.logSeen = seq + 1
	}
	return true, nil
}

// adoptedStack adds the blocks to mark.Live and returns those it added, for a
// walk from them.
func adoptedStack(mark *BlockMark, blocks []string) []namedNode {
	var stack []namedNode
	for _, block := range blocks {
		if mark.Live.Add(block) {
			stack = append(stack, namedNode{node: block})
		}
	}
	return stack
}

// RescueAdopted consumes the adopt log, adding each block it names and every
// node reachable from one to mark.Live, and returns the number of nodes added.
// A sweep runs it first in each transaction that removes blocks, so no block
// a committed write adopted is removed. It skips the records CatchUpAdopted
// walked. It reads again the lists of the unknown blocks the mark listed and
// walks those it finds. A reached block without a list counts in
// mark.Unknown, and the sweep must remove nothing while any does. An empty
// log is left unwritten, so a rescue that finds nothing new commits nothing.
func (g *RefGraph) RescueAdopted(ctx context.Context, mark *BlockMark) (int, error) {
	var rescued int
	err := g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
		// Consume the log and advance the pass past it.
		r, open, err := readPassTx(ctx, tx)
		if err != nil || !open {
			return err
		}
		var blocks []string
		if r.legacy || r.lo != r.hi {
			unseen := adoptRange{lo: max(r.lo, mark.logSeen), hi: r.hi, legacy: r.legacy}
			if blocks, err = readAdoptLogTx(ctx, tx, unseen); err != nil {
				return err
			}
			next, err := deleteAdoptLogTx(ctx, tx, r, open)
			if err != nil {
				return err
			}
			if err := tx.Set(ctx, passKey, adoptRange{lo: next, hi: next}.encode()); err != nil {
				return err
			}
			mark.logSeen = next
		}

		// Walk from the listed unknown blocks and the logged blocks the mark
		// and earlier rescues did not reach.
		before := mark.Live.Len()
		stack := append(mark.takeAwaiting(), adoptedStack(mark, blocks)...)
		err = walkTx(ctx, tx, mark.Live, stack, mark.noteUnknown)
		rescued = mark.Live.Len() - before
		return err
	})
	return rescued, err
}

// readAdoptLogTx returns the blocks of every record in the adopt log r.
func readAdoptLogTx(ctx context.Context, tx kvtx.Tx, r adoptRange) ([]string, error) {
	// A legacy log is found by its prefix.
	var blocks []string
	if r.legacy {
		err := tx.ScanPrefix(ctx, []byte{kindAdopt}, func(_, value []byte) error {
			logged, err := decodeChildren(value)
			blocks = append(blocks, logged...)
			return err
		})
		return blocks, err
	}

	// Read each record of the range.
	for seq := r.lo; seq < r.hi; seq++ {
		value, found, err := tx.Get(ctx, adoptKey(seq))
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		logged, err := decodeChildren(value)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, logged...)
	}
	return blocks, nil
}

// logAdoptedTx appends the blocks a write adopted to the adopt log while a
// block pass runs. It reads and writes only the pass record and the new log
// record, so its cost does not grow with the log or its deleted records.
func logAdoptedTx(ctx context.Context, tx kvtx.Tx, blocks []string) error {
	// Log only during a pass.
	if len(blocks) == 0 {
		return nil
	}
	r, open, err := readPassTx(ctx, tx)
	if err != nil || !open {
		return err
	}

	// A legacy pass numbers its log once by a scan, then keeps the range.
	if r.legacy {
		if r, err = legacyRangeTx(ctx, tx); err != nil {
			return err
		}
	}

	// Write the sorted, distinct blocks at the next sequence.
	blocks = slices.Clone(blocks)
	slices.Sort(blocks)
	if err := tx.Set(ctx, adoptKey(r.hi), encodeChildren(slices.Compact(blocks))); err != nil {
		return err
	}
	r.hi++
	return tx.Set(ctx, passKey, r.encode())
}

// legacyRangeTx returns the range of a legacy pass's log, from its first
// record to after its last.
func legacyRangeTx(ctx context.Context, tx kvtx.Tx) (adoptRange, error) {
	// Find the first and the last record.
	var r adoptRange
	for _, reverse := range []bool{false, true} {
		it := tx.Iterate(ctx, []byte{kindAdopt}, false, reverse)
		if it.Next() {
			seq := binary.BigEndian.Uint64(it.Key()[1:])
			if reverse {
				r.hi = seq + 1
			} else {
				r.lo = seq
			}
		}
		err := it.Err()
		it.Close()
		if err != nil {
			return adoptRange{}, err
		}
	}
	return r, nil
}

// adoptKey returns the key of the adopt log record seq.
func adoptKey(seq uint64) []byte {
	return binary.BigEndian.AppendUint64([]byte{kindAdopt}, seq)
}

// deleteAdoptLogTx deletes the adopt log r of a running pass, when open, and
// returns the sequence after it.
func deleteAdoptLogTx(ctx context.Context, tx kvtx.Tx, r adoptRange, open bool) (uint64, error) {
	// Without a pass there is no log.
	if !open {
		return 0, nil
	}

	// A legacy log is found by its prefix.
	if r.legacy {
		var keys [][]byte
		var next uint64
		err := tx.ScanPrefixKeys(ctx, []byte{kindAdopt}, func(key []byte) error {
			keys = append(keys, slices.Clone(key))
			next = max(next, binary.BigEndian.Uint64(key[1:])+1)
			return nil
		})
		if err != nil {
			return 0, err
		}
		for _, key := range keys {
			if err := tx.Delete(ctx, key); err != nil {
				return 0, err
			}
		}
		return next, nil
	}

	// Delete each record of the range.
	for seq := r.lo; seq < r.hi; seq++ {
		if err := tx.Delete(ctx, adoptKey(seq)); err != nil {
			return 0, err
		}
	}
	return r.hi, nil
}
