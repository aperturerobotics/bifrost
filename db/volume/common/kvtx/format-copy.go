package kvtx

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/s4wave/spacewave/db/block"
	block_gc "github.com/s4wave/spacewave/db/block/gc"
	"github.com/s4wave/spacewave/db/kvtx"
	store_kvkey "github.com/s4wave/spacewave/db/store/kvkey"
)

// CopyProgress counts the work of a running CopyStore.
type CopyProgress struct {
	// Keys is the number of keys copied verbatim.
	Keys uint64
	// Bytes is the size of the keys and values copied verbatim.
	Bytes uint64
	// Edges is the number of Cayley log edges converted.
	Edges uint64
	// DroppedBlocks is the number of blocks a reachable copy left out.
	DroppedBlocks uint64
	// UnownedBlocks is the number of dropped blocks no graph record owned,
	// which a running collector would never have deleted.
	UnownedBlocks uint64
	// DroppedRecords is the number of graph records a reachable copy left out.
	DroppedRecords uint64
	// DroppedBytes is the size of the keys and values left out.
	DroppedBytes uint64
}

// CopyStore writes the Volume store src into the empty store dst at
// store_kvkey.FormatVersion. It copies every key outside the Cayley graph,
// encoding the version 2 graph records again, then converts the Cayley graph's
// live edges to edge records, so a store of format version 1 or 2, converted
// in part or not at all, never rewrites its graph in place.
//
// A graph copied from a store before format version 4 gets the refs-known
// edges the format migration seeds. src is read in one snapshot and must have
// no writer. The format version is written last in a durable commit; a dst
// without it is an incomplete copy to discard. progress, when set, is called
// after each commit.
func CopyStore(ctx context.Context, src, dst kvtx.Store, keys *store_kvkey.KVKey, progress func(CopyProgress)) error {
	// Accept the current format and the earlier graph formats it converts.
	version, err := readFormatVersion(ctx, src, keys)
	if err != nil {
		return err
	}
	if version == 0 || version > store_kvkey.FormatVersion {
		return fmt.Errorf("copy needs store format version 1 to %d, found %d", store_kvkey.FormatVersion, version)
	}
	return copyStore(ctx, src, dst, keys, false, version < store_kvkey.FormatVersion, progress)
}

// CopyReachable writes the Volume store src into the empty store dst like
// CopyStore, keeping only what a complete collection would keep: the blocks
// reachable from the graph root, the graph records of reachable nodes and
// every key outside the graph and the blocks. It drops the unreferenced queue.
//
// src must be at store_kvkey.FormatVersion: the walk reads only the current
// graph, so the records of a store not yet converted would be missed and
// their blocks dropped. A dropped block no graph record owned is counted in
// UnownedBlocks, since a running collector would never have deleted it.
func CopyReachable(ctx context.Context, src, dst kvtx.Store, keys *store_kvkey.KVKey, progress func(CopyProgress)) error {
	// Require a fully converted source.
	version, err := readFormatVersion(ctx, src, keys)
	if err != nil {
		return err
	}
	if version != store_kvkey.FormatVersion {
		return fmt.Errorf("reachable copy needs store format version %d, found %d; copy it to the current format first", store_kvkey.FormatVersion, version)
	}
	return copyStore(ctx, src, dst, keys, true, false, progress)
}

// copyStore copies src into the empty store dst and records the format
// version last. With reachable, it first marks the nodes reachable from the
// graph root and copies only what they keep. With seed, it marks the refs of
// every block the copied graph names as known.
func copyStore(
	ctx context.Context,
	src, dst kvtx.Store,
	keys *store_kvkey.KVKey,
	reachable, seed bool,
	progress func(CopyProgress),
) error {
	// Refuse a destination that already holds keys.
	if err := requireEmptyStore(ctx, dst); err != nil {
		return err
	}

	// Read the source in one snapshot.
	stx, err := src.NewTransaction(ctx, false)
	if err != nil {
		return err
	}
	defer stx.Discard()

	// Mark the reachable nodes in the same snapshot.
	w := &copyWriter{store: dst, progress: progress}
	if reachable {
		w.graph = openRefGraph(kvtx.NewTxStore(stx))
		w.blockPrefix = keys.GetBlockFullPrefix()
		w.live, err = w.graph.Reachable(ctx)
		if err != nil {
			return fmt.Errorf("mark reachable nodes: %w", err)
		}
	}

	// Copy the keys, then convert the Cayley graph.
	if err := w.copyKeys(ctx, stx, keys.GetFormatVersionKey()); err != nil {
		return err
	}
	if err := w.convertEdges(ctx, stx); err != nil {
		return err
	}
	if seed {
		if err := openRefGraph(dst).SeedRefsKnown(ctx); err != nil {
			return fmt.Errorf("seed refs-known edges: %w", err)
		}
	}

	// Record the version in a durable commit that also makes the ordered
	// commits durable.
	tx, err := dst.NewTransaction(ctx, true)
	if err != nil {
		return err
	}
	defer tx.Discard()
	if err := tx.Set(ctx, keys.GetFormatVersionKey(), store_kvkey.MarshalFormatVersion(store_kvkey.FormatVersion)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// requireEmptyStore returns an error unless store holds no keys.
func requireEmptyStore(ctx context.Context, store kvtx.Store) error {
	// Count the keys in a snapshot.
	tx, err := store.NewTransaction(ctx, false)
	if err != nil {
		return err
	}
	defer tx.Discard()
	size, err := tx.Size(ctx)
	if err != nil {
		return err
	}
	if size != 0 {
		return errors.New("copy destination is not empty")
	}
	return nil
}

// copyWriter writes the destination of a CopyStore in ordered commits.
type copyWriter struct {
	// store is the destination.
	store kvtx.Store
	// progress receives the counts after each commit, when set.
	progress func(CopyProgress)
	// counts is the work committed so far.
	counts CopyProgress
	// live holds the reachable nodes of a reachable copy; nil copies every
	// key.
	live *block_gc.NodeSet
	// blockPrefix is the prefix of the block keys a reachable copy filters.
	blockPrefix []byte
	// graph reads the source's reference graph in the copy's snapshot during
	// a reachable copy.
	graph *block_gc.RefGraph
}

// copyKeys copies every key of stx except the Cayley graph and versionKey,
// in commits of up to migrateBatchBytes. It writes each version 2 graph
// record in the current encoding.
func (w *copyWriter) copyKeys(ctx context.Context, stx kvtx.Tx, versionKey []byte) error {
	// Walk every key in order.
	it := stx.Iterate(ctx, nil, true, false)
	defer it.Close()
	if err := it.Seek(nil); err != nil {
		return err
	}

	// Fill each batch, seeking past the Cayley graph in one step.
	var batch []keyValue
	var size int
	for it.Valid() {
		key := it.Key()
		if bytes.HasPrefix(key, cayleyGraphPrefix) {
			if err := it.Seek(prefixEnd(cayleyGraphPrefix)); err != nil {
				return err
			}
			continue
		}
		if !bytes.Equal(key, versionKey) {
			value, err := it.ValueCopy(nil)
			if err != nil {
				return err
			}
			key = bytes.Clone(key)
			if record, ok := bytes.CutPrefix(key, legacyRefGraphPrefix); ok {
				current, err := block_gc.LegacyRecordKey(record)
				if err != nil {
					return err
				}
				key = append(bytes.Clone(refGraphPrefix), current...)
			}
			keep, err := w.keep(ctx, key)
			if err != nil {
				return err
			}
			if keep {
				batch = append(batch, keyValue{key: key, value: value})
				size += len(key) + len(value)
			} else {
				w.counts.DroppedBytes += uint64(len(key) + len(value)) //nolint:gosec // lengths are never negative.
			}
		}
		if size >= migrateBatchBytes {
			if err := w.commitKeys(ctx, batch, size); err != nil {
				return err
			}
			batch, size = batch[:0], 0
		}
		it.Next()
	}
	if err := it.Err(); err != nil {
		return err
	}
	return w.commitKeys(ctx, batch, size)
}

// keep reports whether the copy writes key, counting a dropped key. A
// reachable copy drops the graph records of unreachable nodes and the blocks
// no reachable node names.
func (w *copyWriter) keep(ctx context.Context, key []byte) (bool, error) {
	// A full copy keeps every key.
	if w.live == nil {
		return true, nil
	}

	// Keep a graph record whose subject is reachable.
	if record, ok := bytes.CutPrefix(key, refGraphPrefix); ok {
		subject, err := block_gc.RecordSubject(record)
		if err != nil {
			return false, err
		}
		if w.live.Has(subject) {
			return true, nil
		}
		w.counts.DroppedRecords++
		return false, nil
	}

	// Keep a reachable block and every key outside the blocks.
	data, ok := bytes.CutPrefix(key, w.blockPrefix)
	if !ok {
		return true, nil
	}
	ref := &block.BlockRef{}
	if err := ref.UnmarshalVT(data); err != nil {
		return false, fmt.Errorf("decode block key %q: %w", key, err)
	}
	node := block_gc.BlockIRI(ref)
	if w.live.Has(node) {
		return true, nil
	}

	// Count a dropped block that no record names as an object, not even the
	// unreferenced queue.
	w.counts.DroppedBlocks++
	owners, err := w.graph.GetIncomingRefs(ctx, node)
	if err != nil {
		return false, err
	}
	if len(owners) == 0 {
		w.counts.UnownedBlocks++
	}
	return false, nil
}

// commitKeys writes batch, holding size bytes, in one ordered commit.
func (w *copyWriter) commitKeys(ctx context.Context, batch []keyValue, size int) error {
	// Write the batch.
	if len(batch) == 0 {
		return nil
	}
	tx, err := w.store.NewTransaction(ctx, true)
	if err != nil {
		return err
	}
	defer tx.Discard()
	for _, m := range batch {
		if err := tx.Set(ctx, m.key, m.value); err != nil {
			return err
		}
	}
	if err := kvtx.CommitOrdered(ctx, tx); err != nil {
		return err
	}

	// Count the committed work.
	w.counts.Keys += uint64(len(batch))
	w.counts.Bytes += uint64(size) //nolint:gosec
	w.report()
	return nil
}

// convertEdges writes the live reference edges of the Cayley log in stx as
// edge records, convertBatchEdges log edges per commit.
func (w *copyWriter) convertEdges(ctx context.Context, stx kvtx.Tx) error {
	for start := cayleyLogPrefix; start != nil; {
		// Read the next log edges and their reference edges from the snapshot.
		edges, next, err := collectCayleyEdges(ctx, stx, start)
		if err != nil {
			return err
		}
		refs, err := resolveCayleyRefs(ctx, stx, edges)
		if err != nil {
			return err
		}

		// Write them in one commit.
		tx, err := w.store.NewTransaction(ctx, true)
		if err != nil {
			return err
		}
		err = openRefGraph(kvtx.NewTxStore(tx)).ApplyRefBatch(ctx, refs, nil)
		if err == nil {
			err = kvtx.CommitOrdered(ctx, tx)
		}
		tx.Discard()
		if err != nil {
			return err
		}
		w.counts.Edges += uint64(len(edges))
		w.report()
		start = next
	}
	return nil
}

// report sends the committed counts to the progress callback.
func (w *copyWriter) report() {
	if w.progress != nil {
		w.progress(w.counts)
	}
}

// prefixEnd returns the first key after every key with prefix, which must
// not end in 0xff.
func prefixEnd(prefix []byte) []byte {
	end := bytes.Clone(prefix)
	end[len(end)-1]++
	return end
}
