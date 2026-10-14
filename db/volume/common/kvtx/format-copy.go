package kvtx

import (
	"bytes"
	"context"
	"errors"
	"fmt"

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
}

// CopyStore writes the Volume store src into the empty store dst at
// store_kvkey.FormatVersion. It copies every key outside the Cayley graph,
// encoding the version 2 graph records again, then converts the Cayley graph's
// live edges to edge records, so a store of format version 1 or 2, converted
// in part or not at all, never rewrites its graph in place.
//
// src is read in one snapshot and must have no writer. The format version is
// written last in a durable commit; a dst without it is an incomplete copy to
// discard. progress, when set, is called after each commit.
func CopyStore(ctx context.Context, src, dst kvtx.Store, keys *store_kvkey.KVKey, progress func(CopyProgress)) error {
	// Accept the current format and the earlier graph formats it converts.
	version, err := readFormatVersion(ctx, src, keys)
	if err != nil {
		return err
	}
	if version == 0 || version > store_kvkey.FormatVersion {
		return fmt.Errorf("copy needs store format version 1 to %d, found %d", store_kvkey.FormatVersion, version)
	}
	if err := requireEmptyStore(ctx, dst); err != nil {
		return err
	}

	// Read the source in one snapshot.
	stx, err := src.NewTransaction(ctx, false)
	if err != nil {
		return err
	}
	defer stx.Discard()

	// Copy the keys, then convert the Cayley graph.
	w := &copyWriter{store: dst, progress: progress}
	if err := w.copyKeys(ctx, stx, keys.GetFormatVersionKey()); err != nil {
		return err
	}
	if err := w.convertEdges(ctx, stx); err != nil {
		return err
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
			batch = append(batch, keyValue{key: key, value: value})
			size += len(key) + len(value)
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
