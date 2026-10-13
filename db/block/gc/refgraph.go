package block_gc

import (
	"context"
	"encoding/binary"
	"slices"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/kvtx"
)

// maxBatchEdges bounds the input edges committed in one transaction.
const maxBatchEdges = 128

// Record kinds. Each kind byte begins a key followed by length-prefixed node
// IRIs, so no node can be read as a prefix of another. Records hold no value:
// a scan reads the last node from each key.
const (
	// kindForward keys an edge by subject then object.
	kindForward byte = 'f'
	// kindIncoming keys an edge by object then subject.
	kindIncoming byte = 'i'
	// kindNode keys the node inventory of a collector graph.
	kindNode byte = 'n'
	// kindRoot keys the registered roots of a collector graph.
	kindRoot byte = 'r'
)

// RefGraph stores the garbage collection ownership graph as plain records in
// a transactional key-value store. Each edge is one record per direction keyed
// by the full node IRIs, so membership is a point read, traversal is a prefix
// scan and a removed edge leaves nothing behind. Opening a RefGraph reads
// nothing.
type RefGraph struct {
	// store holds the graph records.
	store kvtx.Store
	// inventory records every node and root for the mark-sweep collector.
	inventory bool
}

// NewRefGraph constructs a reference graph over store. The incremental
// collector finds garbage through orphan marks, so this graph keeps no node
// inventory and does not support the CollectorGraph listing methods.
func NewRefGraph(store kvtx.Store) *RefGraph {
	return &RefGraph{store: store}
}

// NewCollectorRefGraph constructs a reference graph over store that also
// records every node and root, as the mark-sweep collector requires.
func NewCollectorRefGraph(store kvtx.Store) *RefGraph {
	return &RefGraph{store: store, inventory: true}
}

// graphKey encodes a record kind followed by unambiguous full node IRIs.
func graphKey(kind byte, nodes ...string) []byte {
	// Size the key buffer for the kind byte and each length-prefixed node.
	size := 1
	for _, node := range nodes {
		size += 4 + len(node)
	}

	// Append the kind byte and each node with its 4-byte length prefix.
	key := make([]byte, 1, size)
	key[0] = kind
	for _, node := range nodes {
		key = binary.BigEndian.AppendUint32(key, uint32(len(node))) //nolint:gosec // node IRIs are far shorter than 4 GiB.
		key = append(key, node...)
	}
	return key
}

// lastNode returns the node that follows prefix in a record key, where
// prefix is a graphKey with one node fewer than key.
func lastNode(prefix, key []byte) string {
	return string(key[len(prefix)+4:])
}

// transaction retries complete graph operations against a fresh consistent view.
func (g *RefGraph) transaction(ctx context.Context, write bool, body func(context.Context, kvtx.Tx) error) error {
	return kvtx.RunTransaction(ctx, write, func(ctx context.Context) (kvtx.Tx, error) {
		return g.store.NewTransaction(ctx, write)
	}, body)
}

// RegisterEntityChain registers a chain of gc/ref edges between nodes.
// Each adjacent pair gets an AddRef call: nodes[0]->nodes[1],
// nodes[1]->nodes[2], etc. At least 2 nodes required. Idempotent.
func RegisterEntityChain(ctx context.Context, rg RefGraphOps, nodes ...string) error {
	// Require enough nodes to define a chain.
	if len(nodes) < 2 {
		return errors.New("RegisterEntityChain requires at least 2 nodes")
	}

	// Register every adjacent pair in chain order.
	for i := 0; i < len(nodes)-1; i++ {
		if err := rg.AddRef(ctx, nodes[i], nodes[i+1]); err != nil {
			return err
		}
	}
	return nil
}

// AddRef atomically inserts both edge directions. Idempotent.
func (g *RefGraph) AddRef(ctx context.Context, subject, object string) error {
	return g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
		return g.addRefTx(ctx, tx, subject, object)
	})
}

// RemoveRef removes an exact edge without creating an orphan mark.
func (g *RefGraph) RemoveRef(ctx context.Context, subject, object string) error {
	return g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
		_, err := g.removeRefTx(ctx, tx, subject, object)
		return err
	})
}

// batchRemainderError reports the suffix left after an earlier slice committed.
type batchRemainderError struct {
	// err is the failure that stopped the ownership transition.
	err error
	// adds is the uncommitted suffix of the original additions.
	adds []RefEdge
	// removes is the uncommitted suffix of the original removals.
	removes []RefEdge
}

// Error describes the failed partially committed ownership transition.
func (e *batchRemainderError) Error() string {
	return "apply ref batch after partial commit: " + e.err.Error()
}

// Unwrap returns the storage failure that interrupted the transition.
func (e *batchRemainderError) Unwrap() error {
	return e.err
}

// RefBatchRemainder returns copies of the uncommitted input suffixes.
func (e *batchRemainderError) RefBatchRemainder() ([]RefEdge, []RefEdge) {
	return slices.Clone(e.adds), slices.Clone(e.removes)
}

// ApplyRefBatch commits a bounded ownership transition atomically. Larger
// transitions commit slices in add-before-remove order and report an exact
// uncommitted suffix if a later slice fails.
func (g *RefGraph) ApplyRefBatch(ctx context.Context, adds, removes []RefEdge) error {
	addIndex, removeIndex := 0, 0
	committed := false
	for addIndex < len(adds) || removeIndex < len(removes) {
		// Take the next slice, finishing the additions before any removal.
		startAdds, startRemoves := addIndex, removeIndex
		remaining := maxBatchEdges
		addIndex = min(addIndex+remaining, len(adds))
		remaining -= addIndex - startAdds
		if addIndex == len(adds) {
			removeIndex = min(removeIndex+remaining, len(removes))
		}

		// Commit the slice, reporting the uncommitted suffix after the first.
		err := g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
			return g.applyRefBatchTx(ctx, tx, adds[startAdds:addIndex], removes[startRemoves:removeIndex])
		})
		if err != nil {
			if !committed {
				return err
			}
			return &batchRemainderError{
				err:     err,
				adds:    slices.Clone(adds[startAdds:]),
				removes: slices.Clone(removes[startRemoves:]),
			}
		}
		committed = true
	}
	return nil
}

// applyRefBatchTx applies one transaction slice in ownership order, marking
// each object that loses its last owner as unreferenced.
func (g *RefGraph) applyRefBatchTx(ctx context.Context, tx kvtx.Tx, adds, removes []RefEdge) error {
	// Add every edge first, so an edge in both lists ends removed.
	for _, edge := range adds {
		if err := g.addRefTx(ctx, tx, edge.Subject, edge.Object); err != nil {
			return err
		}
	}

	// Remove every edge, noting the objects that lost an owner and the
	// objects whose staging edge the caller removed.
	var released []string
	unstaged := make(map[string]struct{})
	for _, edge := range removes {
		existed, err := g.removeRefTx(ctx, tx, edge.Subject, edge.Object)
		switch {
		case err != nil:
			return err
		case !existed || IsPermanentRoot(edge.Object):
		case edge.Subject == NodeUnreferenced:
			unstaged[edge.Object] = struct{}{}
		default:
			released = append(released, edge.Object)
		}
	}

	// Mark each released object left without an owner, unless the caller
	// removed its staging edge.
	for _, object := range released {
		if _, ok := unstaged[object]; ok {
			continue
		}
		owned, err := hasIncomingTx(ctx, tx, object, nil)
		if err != nil {
			return err
		}
		if !owned {
			if err := g.addRefTx(ctx, tx, NodeUnreferenced, object); err != nil {
				return err
			}
		}
	}
	return nil
}

// addRefTx adds both edge directions, and the endpoint inventory of a
// collector graph, unless the exact forward edge already exists.
func (g *RefGraph) addRefTx(ctx context.Context, tx kvtx.Tx, subject, object string) error {
	// Skip the edge when its exact forward record already exists.
	forward := graphKey(kindForward, subject, object)
	exists, err := tx.Exists(ctx, forward)
	if err != nil || exists {
		return err
	}

	// Record both endpoints in the inventory of a collector graph.
	if g.inventory {
		if err := tx.Set(ctx, graphKey(kindNode, subject), nil); err != nil {
			return err
		}
		if err := tx.Set(ctx, graphKey(kindNode, object), nil); err != nil {
			return err
		}
	}

	// Write the two edge directions.
	if err := tx.Set(ctx, forward, nil); err != nil {
		return err
	}
	return tx.Set(ctx, graphKey(kindIncoming, object, subject), nil)
}

// removeRefTx removes both directions of an exact edge and reports whether it
// existed.
func (g *RefGraph) removeRefTx(ctx context.Context, tx kvtx.Tx, subject, object string) (bool, error) {
	// Check the forward record, then delete both directions.
	forward := graphKey(kindForward, subject, object)
	exists, err := tx.Exists(ctx, forward)
	if err != nil || !exists {
		return exists, err
	}
	if err := tx.Delete(ctx, forward); err != nil {
		return false, err
	}
	if err := tx.Delete(ctx, graphKey(kindIncoming, object, subject)); err != nil {
		return false, err
	}
	return true, nil
}

// hasIncomingTx reports whether object has an owner other than the
// unreferenced root and the ignored subjects, stopping at the first one.
func hasIncomingTx(ctx context.Context, tx kvtx.Tx, object string, ignored map[string]struct{}) (bool, error) {
	// Walk the object's incoming records; each key ends in its subject.
	prefix := graphKey(kindIncoming, object)
	it := tx.Iterate(ctx, prefix, true, false)
	defer it.Close()
	for it.Next() {
		subject := lastNode(prefix, it.Key())
		if subject == NodeUnreferenced {
			continue
		}
		if _, skip := ignored[subject]; !skip {
			return true, nil
		}
	}
	return false, it.Err()
}

// RemoveNodeRefs removes the observed outgoing targets in bounded transactions.
// New unrelated edges may be added after this operation's initial snapshot.
func (g *RefGraph) RemoveNodeRefs(ctx context.Context, node string, markOrphaned bool) ([]string, error) {
	targets, err := g.GetOutgoingRefs(ctx, node)
	if err != nil {
		return nil, err
	}
	for offset := 0; offset < len(targets); offset += maxBatchEdges {
		chunk := targets[offset:min(offset+maxBatchEdges, len(targets))]
		err := g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
			if markOrphaned {
				removes := make([]RefEdge, 0, len(chunk))
				for _, target := range chunk {
					removes = append(removes, RefEdge{Subject: node, Object: target})
				}
				return g.applyRefBatchTx(ctx, tx, nil, removes)
			}
			for _, target := range chunk {
				if _, err := g.removeRefTx(ctx, tx, node, target); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return targets, nil
}

// HasIncomingRefs ignores the bookkeeping edge from the unreferenced root.
func (g *RefGraph) HasIncomingRefs(ctx context.Context, node string) (bool, error) {
	return g.HasIncomingRefsExcluding(ctx, node)
}

// HasIncomingRefsExcluding stops at the first owner outside the unreferenced
// root and the caller's exclusions.
func (g *RefGraph) HasIncomingRefsExcluding(ctx context.Context, node string, excluded ...string) (bool, error) {
	// Index the exclusions, then scan the node's owners.
	ignored := make(map[string]struct{}, len(excluded))
	for _, source := range excluded {
		ignored[source] = struct{}{}
	}
	var found bool
	err := g.transaction(ctx, false, func(ctx context.Context, tx kvtx.Tx) error {
		var err error
		found, err = hasIncomingTx(ctx, tx, node, ignored)
		return err
	})
	return found, err
}

// list returns the last node of every record under prefix.
func (g *RefGraph) list(ctx context.Context, prefix []byte) ([]string, error) {
	var answer []string
	err := g.transaction(ctx, false, func(ctx context.Context, tx kvtx.Tx) error {
		answer = nil
		return tx.ScanPrefixKeys(ctx, prefix, func(key []byte) error {
			answer = append(answer, lastNode(prefix, key))
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return answer, nil
}

// GetOutgoingRefs returns all current targets of a node.
func (g *RefGraph) GetOutgoingRefs(ctx context.Context, node string) ([]string, error) {
	return g.list(ctx, graphKey(kindForward, node))
}

// GetIncomingRefs returns all current owners of a node.
func (g *RefGraph) GetIncomingRefs(ctx context.Context, node string) ([]string, error) {
	return g.list(ctx, graphKey(kindIncoming, node))
}

// GetUnreferencedNodes returns the nodes marked for collector reconsideration.
func (g *RefGraph) GetUnreferencedNodes(ctx context.Context) ([]string, error) {
	return g.GetOutgoingRefs(ctx, NodeUnreferenced)
}

// AddBlockRef adds gc/ref from source block to target block.
func (g *RefGraph) AddBlockRef(ctx context.Context, source, target *block.BlockRef) error {
	from, to := BlockIRI(source), BlockIRI(target)
	if from == "" || to == "" {
		return nil
	}
	return g.AddRef(ctx, from, to)
}

// AddObjectRoot links an object identity to its current content root.
func (g *RefGraph) AddObjectRoot(ctx context.Context, objectKey string, ref *block.BlockRef) error {
	target := BlockIRI(ref)
	if target == "" {
		return nil
	}
	return g.AddRef(ctx, ObjectIRI(objectKey), target)
}

// RemoveObjectRoot removes an exact object-to-block ownership edge.
func (g *RefGraph) RemoveObjectRoot(ctx context.Context, objectKey string, ref *block.BlockRef) error {
	target := BlockIRI(ref)
	if target == "" {
		return nil
	}
	return g.RemoveRef(ctx, ObjectIRI(objectKey), target)
}

// AddRoot records a node and its root membership in a collector graph.
func (g *RefGraph) AddRoot(ctx context.Context, node string) error {
	return g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
		for _, kind := range []byte{kindNode, kindRoot} {
			key := graphKey(kind, node)
			found, err := tx.Exists(ctx, key)
			if err != nil {
				return err
			}
			if found {
				continue
			}
			if err := tx.Set(ctx, key, nil); err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveRoot removes root membership without changing the node's edges.
func (g *RefGraph) RemoveRoot(ctx context.Context, node string) error {
	return g.deleteRecord(ctx, graphKey(kindRoot, node))
}

// RemoveNode removes inventory after the collector has handled its edges.
func (g *RefGraph) RemoveNode(ctx context.Context, node string) error {
	return g.deleteRecord(ctx, graphKey(kindNode, node))
}

// deleteRecord treats a missing record as a no-op.
func (g *RefGraph) deleteRecord(ctx context.Context, key []byte) error {
	return g.transaction(ctx, true, func(ctx context.Context, tx kvtx.Tx) error {
		found, err := tx.Exists(ctx, key)
		if err != nil || !found {
			return err
		}
		return tx.Delete(ctx, key)
	})
}

// IterateNodes returns the node inventory of a collector graph.
func (g *RefGraph) IterateNodes(ctx context.Context) ([]string, error) {
	return g.list(ctx, graphKey(kindNode))
}

// GetRootNodes returns the registered roots of a collector graph.
func (g *RefGraph) GetRootNodes(ctx context.Context) ([]string, error) {
	return g.list(ctx, graphKey(kindRoot))
}

// Close leaves the store's lifetime with its owner.
func (g *RefGraph) Close() error { return nil }

// _ is a type assertion
var (
	_ RefGraphOps            = (*RefGraph)(nil)
	_ CollectorGraph         = (*RefGraph)(nil)
	_ RefBatchRemainderError = (*batchRemainderError)(nil)
)
