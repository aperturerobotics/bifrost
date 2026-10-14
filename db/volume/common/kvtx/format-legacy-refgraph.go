package kvtx

import (
	"bytes"
	"context"

	block_gc "github.com/s4wave/spacewave/db/block/gc"
	"github.com/s4wave/spacewave/db/kvtx"
	store_kvkey "github.com/s4wave/spacewave/db/store/kvkey"
)

// legacyRefGraphPrefix holds the reference graph of format version 2, whose
// records spell each node as a 4-byte length and its full IRI.
var legacyRefGraphPrefix = []byte("refs/")

// migrateLegacyRefGraph moves the version 2 graph to the current encoding.
func migrateLegacyRefGraph(ctx context.Context, store kvtx.Store, _ *store_kvkey.KVKey) error {
	return MigrateLegacyRefGraph(ctx, store, legacyRefGraphPrefix, refGraphPrefix)
}

// MigrateLegacyRefGraph moves the reference graph records under legacy, in
// the encoding that spelled each node as a length and its full IRI, to the
// current encoding under prefix. Each batch writes its records and deletes the
// originals in one commit, so a rerun resumes with the records not yet moved.
// A store holding none costs one scan. Neither prefix may contain the other.
func MigrateLegacyRefGraph(ctx context.Context, store kvtx.Store, legacy, prefix []byte) error {
	for {
		done, err := moveLegacyRecords(ctx, store, legacy, prefix)
		if err != nil || done {
			return err
		}
	}
}

// moveLegacyRecords moves up to convertBatchEdges records under legacy in one
// commit, and reports whether none remain.
func moveLegacyRecords(ctx context.Context, store kvtx.Store, legacy, prefix []byte) (bool, error) {
	// Read the first records left under legacy.
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		return false, err
	}
	defer tx.Discard()
	records, err := readLegacyRecords(ctx, tx, legacy)
	if err != nil || len(records) == 0 {
		return true, err
	}

	// Write each record in the current encoding and delete the original.
	for _, record := range records {
		key, err := block_gc.LegacyRecordKey(record[len(legacy):])
		if err != nil {
			return false, err
		}
		if err := tx.Set(ctx, append(bytes.Clone(prefix), key...), nil); err != nil {
			return false, err
		}
		if err := tx.Delete(ctx, record); err != nil {
			return false, err
		}
	}
	return len(records) < convertBatchEdges, kvtx.CommitOrdered(ctx, tx)
}

// readLegacyRecords returns the keys of up to convertBatchEdges records under
// legacy.
func readLegacyRecords(ctx context.Context, tx kvtx.Tx, legacy []byte) ([][]byte, error) {
	// Seek to the first record.
	it := tx.Iterate(ctx, legacy, true, false)
	defer it.Close()
	if err := it.Seek(legacy); err != nil {
		return nil, err
	}

	// Collect the keys.
	var records [][]byte
	for ; it.Valid() && len(records) < convertBatchEdges; it.Next() {
		records = append(records, bytes.Clone(it.Key()))
	}
	return records, it.Err()
}
