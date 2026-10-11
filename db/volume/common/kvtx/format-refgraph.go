package kvtx

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	cayley_proto "github.com/aperturerobotics/cayley/graph/proto"
	"github.com/aperturerobotics/cayley/quad"
	"github.com/aperturerobotics/cayley/quad/pquads"
	block_gc "github.com/s4wave/spacewave/db/block/gc"
	"github.com/s4wave/spacewave/db/kvtx"
	store_kvkey "github.com/s4wave/spacewave/db/store/kvkey"
)

// cayleyGraphPrefix holds the Cayley quad store that format version 1 kept
// the reference graph in.
var cayleyGraphPrefix = []byte("gc/")

// cayleyLogPrefix holds the Cayley primitive log: one record per live node or
// edge, keyed by its decimal primitive ID.
var cayleyLogPrefix = []byte("gc/l/")

// convertBatchEdges bounds the edges one conversion commit moves.
const convertBatchEdges = 8192

// migrateBase58Store upgrades a version 0 store: it rekeys the blocks, then
// converts the reference graph.
func migrateBase58Store(ctx context.Context, store kvtx.Store, keys *store_kvkey.KVKey) error {
	if err := migrateBase58BlockKeys(ctx, store, keys); err != nil {
		return err
	}
	return migrateCayleyRefGraph(ctx, store, keys)
}

// migrateCayleyRefGraph moves every reference edge from the Cayley primitive
// log to the edge records of block_gc.RefGraph, then deletes the Cayley store.
// Each batch writes its edges and deletes their log records in one commit,
// and node records are deleted only after every edge has moved, so a rerun
// resumes with every remaining edge's nodes intact.
func migrateCayleyRefGraph(ctx context.Context, store kvtx.Store, _ *store_kvkey.KVKey) error {
	start := cayleyLogPrefix
	for start != nil {
		var err error
		start, err = convertCayleyEdgeBatch(ctx, store, start)
		if err != nil {
			return err
		}
	}
	return deletePrefix(ctx, store, cayleyGraphPrefix)
}

// cayleyEdge is one edge primitive of the Cayley log.
type cayleyEdge struct {
	// key is the edge's log key.
	key []byte
	// subject, predicate and object are the node primitive IDs.
	subject, predicate, object uint64
}

// convertCayleyEdgeBatch moves up to convertBatchEdges edges at or after
// start, and returns the key to resume from, or nil at the end of the log.
func convertCayleyEdgeBatch(ctx context.Context, store kvtx.Store, start []byte) ([]byte, error) {
	// Open the batch's write transaction.
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		return nil, err
	}
	defer tx.Discard()

	// Collect the next edges, skipping node records.
	edges, next, err := collectCayleyEdges(ctx, tx, start)
	if err != nil || len(edges) == 0 {
		return next, err
	}

	// Resolve each node once and write every reference edge.
	names := make(map[uint64]string)
	var adds []block_gc.RefEdge
	for _, e := range edges {
		predicate, err := cayleyNodeName(ctx, tx, names, e.predicate)
		if err != nil {
			return nil, err
		}
		if predicate != block_gc.PredGCRef {
			continue
		}
		subject, err := cayleyNodeName(ctx, tx, names, e.subject)
		if err != nil {
			return nil, err
		}
		object, err := cayleyNodeName(ctx, tx, names, e.object)
		if err != nil {
			return nil, err
		}
		adds = append(adds, block_gc.RefEdge{Subject: subject, Object: object})
	}
	if err := openRefGraph(kvtx.NewTxStore(tx)).ApplyRefBatch(ctx, adds, nil); err != nil {
		return nil, err
	}

	// Delete the moved edges from the log in the same commit.
	for _, e := range edges {
		if err := tx.Delete(ctx, e.key); err != nil {
			return nil, err
		}
	}
	return next, kvtx.CommitOrdered(ctx, tx)
}

// collectCayleyEdges reads up to convertBatchEdges live edge primitives at or
// after start, and returns the key after the last one read, or nil when the
// log ends.
func collectCayleyEdges(ctx context.Context, tx kvtx.Tx, start []byte) ([]cayleyEdge, []byte, error) {
	// Seek to the batch start.
	it := tx.Iterate(ctx, cayleyLogPrefix, true, false)
	defer it.Close()
	if err := it.Seek(start); err != nil {
		return nil, nil, err
	}

	// Collect live edges, skipping node and deleted records.
	var edges []cayleyEdge
	for ; it.Valid(); it.Next() {
		if len(edges) == convertBatchEdges {
			return edges, bytes.Clone(it.Key()), nil
		}
		value, err := it.Value()
		if err != nil {
			return nil, nil, err
		}
		var p cayley_proto.Primitive
		if err := p.UnmarshalVT(value); err != nil {
			return nil, nil, fmt.Errorf("decode reference graph record %q: %w", it.Key(), err)
		}
		if p.Deleted || p.IsNode() {
			continue
		}
		edges = append(edges, cayleyEdge{
			key:       bytes.Clone(it.Key()),
			subject:   p.Subject,
			predicate: p.Predicate,
			object:    p.Object,
		})
	}
	return edges, nil, it.Err()
}

// cayleyNodeName returns the IRI of a node primitive, reading it from the log
// on first use.
func cayleyNodeName(ctx context.Context, tx kvtx.Tx, names map[uint64]string, id uint64) (string, error) {
	// Reuse a name this batch already resolved.
	if name, ok := names[id]; ok {
		return name, nil
	}

	// Read the node's log record.
	data, found, err := tx.Get(ctx, strconv.AppendUint(bytes.Clone(cayleyLogPrefix), id, 10))
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("reference graph node %d is missing", id)
	}

	// Decode the record's IRI.
	var p cayley_proto.Primitive
	if err := p.UnmarshalVT(data); err != nil {
		return "", fmt.Errorf("decode reference graph node %d: %w", id, err)
	}
	value, err := pquads.UnmarshalValue(ctx, p.Value)
	if err != nil {
		return "", fmt.Errorf("decode reference graph node %d: %w", id, err)
	}
	iri, ok := value.(quad.IRI)
	if !ok {
		return "", fmt.Errorf("reference graph node %d is %T, not an IRI", id, value)
	}

	// Cache the name for the rest of the batch.
	names[id] = string(iri)
	return string(iri), nil
}

// deletePrefix deletes every key under prefix in bounded commits.
func deletePrefix(ctx context.Context, store kvtx.Store, prefix []byte) error {
	for {
		more, err := deletePrefixBatch(ctx, store, prefix)
		if err != nil || !more {
			return err
		}
	}
}

// deletePrefixBatch deletes up to convertBatchEdges keys under prefix in one
// commit, and reports whether more remain.
func deletePrefixBatch(ctx context.Context, store kvtx.Store, prefix []byte) (bool, error) {
	// Open the batch's write transaction.
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		return false, err
	}
	defer tx.Discard()

	// Collect the next keys.
	it := tx.Iterate(ctx, prefix, false, false)
	var keys [][]byte
	for it.Next() && len(keys) < convertBatchEdges {
		keys = append(keys, bytes.Clone(it.Key()))
	}
	more := it.Valid()
	err = it.Err()
	it.Close()
	if err != nil || len(keys) == 0 {
		return false, err
	}

	// Delete them in one commit.
	for _, key := range keys {
		if err := tx.Delete(ctx, key); err != nil {
			return false, err
		}
	}
	return more, kvtx.CommitOrdered(ctx, tx)
}
