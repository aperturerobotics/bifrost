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

	// Write every reference edge.
	adds, err := resolveCayleyRefs(ctx, tx, edges)
	if err != nil {
		return nil, err
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

// resolveCayleyRefs returns the reference edges among edges, reading each
// node's name from tx once.
func resolveCayleyRefs(ctx context.Context, tx kvtx.Tx, edges []cayleyEdge) ([]block_gc.RefEdge, error) {
	names := make(map[uint64]string)
	var refs []block_gc.RefEdge
	for _, e := range edges {
		// Skip edges of other predicates.
		predicate, err := cayleyNodeName(ctx, tx, names, e.predicate)
		if err != nil {
			return nil, err
		}
		if predicate != block_gc.PredGCRef {
			continue
		}

		// Name both ends.
		subject, err := cayleyNodeName(ctx, tx, names, e.subject)
		if err != nil {
			return nil, err
		}
		object, err := cayleyNodeName(ctx, tx, names, e.object)
		if err != nil {
			return nil, err
		}
		refs = append(refs, block_gc.RefEdge{Subject: subject, Object: object})
	}
	return refs, nil
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

// deletePrefix deletes every key under prefix in bounded commits. Each batch
// seeks past the keys earlier batches deleted.
func deletePrefix(ctx context.Context, store kvtx.Store, prefix []byte) error {
	for start := prefix; start != nil; {
		var err error
		start, err = deletePrefixBatch(ctx, store, prefix, start)
		if err != nil {
			return err
		}
	}
	return nil
}

// deletePrefixBatch deletes up to convertBatchEdges keys under prefix at or
// after start in one commit, and returns the key to resume from, or nil when
// none remain.
func deletePrefixBatch(ctx context.Context, store kvtx.Store, prefix, start []byte) ([]byte, error) {
	// Open the batch's write transaction.
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		return nil, err
	}
	defer tx.Discard()

	// Collect the next keys.
	it := tx.Iterate(ctx, prefix, true, false)
	var keys [][]byte
	err = it.Seek(start)
	for ; err == nil && it.Valid() && len(keys) < convertBatchEdges; it.Next() {
		keys = append(keys, bytes.Clone(it.Key()))
	}
	more := err == nil && it.Valid()
	if err == nil {
		err = it.Err()
	}
	it.Close()
	if err != nil || len(keys) == 0 {
		return nil, err
	}

	// Delete them in one commit.
	for _, key := range keys {
		if err := tx.Delete(ctx, key); err != nil {
			return nil, err
		}
	}
	if err := kvtx.CommitOrdered(ctx, tx); err != nil || !more {
		return nil, err
	}
	return keyAfter(keys[len(keys)-1]), nil
}

// keyAfter returns the least key greater than key.
func keyAfter(key []byte) []byte {
	return append(bytes.Clone(key), 0)
}
