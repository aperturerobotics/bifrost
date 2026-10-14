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
// Each batch seeks past the records earlier batches moved, since the deleted
// keys stay in the store's layers until a checkpoint folds them and stepping
// over them again would make the migration quadratic. A store holding none
// costs one scan. Neither prefix may contain the other.
func MigrateLegacyRefGraph(ctx context.Context, store kvtx.Store, legacy, prefix []byte) error {
	for start := legacy; start != nil; {
		var err error
		start, err = moveLegacyRecords(ctx, store, legacy, prefix, start)
		if err != nil {
			return err
		}
	}
	return nil
}

// moveLegacyRecords moves up to convertBatchEdges records under legacy at or
// after start in one commit, and returns the key to resume from, or nil when
// none remain.
func moveLegacyRecords(ctx context.Context, store kvtx.Store, legacy, prefix, start []byte) ([]byte, error) {
	// Read the next records left under legacy.
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		return nil, err
	}
	defer tx.Discard()
	records, err := readLegacyRecords(ctx, tx, legacy, start)
	if err != nil || len(records) == 0 {
		return nil, err
	}

	// Write each record in the current encoding and delete the original.
	for _, record := range records {
		key, err := block_gc.LegacyRecordKey(record[len(legacy):])
		if err != nil {
			return nil, err
		}
		if err := tx.Set(ctx, append(bytes.Clone(prefix), key...), nil); err != nil {
			return nil, err
		}
		if err := tx.Delete(ctx, record); err != nil {
			return nil, err
		}
	}
	if err := kvtx.CommitOrdered(ctx, tx); err != nil || len(records) < convertBatchEdges {
		return nil, err
	}
	return keyAfter(records[len(records)-1]), nil
}

// readLegacyRecords returns the keys of up to convertBatchEdges records under
// legacy at or after start.
func readLegacyRecords(ctx context.Context, tx kvtx.Tx, legacy, start []byte) ([][]byte, error) {
	// Seek to the first record.
	it := tx.Iterate(ctx, legacy, true, false)
	defer it.Close()
	if err := it.Seek(start); err != nil {
		return nil, err
	}

	// Collect the keys.
	var records [][]byte
	for ; it.Valid() && len(records) < convertBatchEdges; it.Next() {
		records = append(records, bytes.Clone(it.Key()))
	}
	return records, it.Err()
}
