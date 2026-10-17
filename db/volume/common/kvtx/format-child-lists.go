package kvtx

import (
	"bytes"
	"context"

	block_gc "github.com/s4wave/spacewave/db/block/gc"
	"github.com/s4wave/spacewave/db/kvtx"
	kvtx_prefixer "github.com/s4wave/spacewave/db/kvtx/prefixer"
	store_kvkey "github.com/s4wave/spacewave/db/store/kvkey"
)

// migrateChildLists converts the version 3 graph to child lists in place.
func migrateChildLists(ctx context.Context, store kvtx.Store, _ *store_kvkey.KVKey) error {
	return MigrateChildLists(ctx, store, refGraphPrefix)
}

// thenChildLists runs an upgrade that ends at the version 3 graph, then
// converts it to child lists.
func thenChildLists(migrate formatMigration) formatMigration {
	return func(ctx context.Context, store kvtx.Store, keys *store_kvkey.KVKey) error {
		if err := migrate(ctx, store, keys); err != nil {
			return err
		}
		return migrateChildLists(ctx, store, keys)
	}
}

// MigrateChildLists converts the reference graph under prefix from store
// format 3 to child lists, as block_gc.ConvertFormat3Record says. Each batch
// deletes its converted records and writes their edges in one commit, so a
// rerun resumes with the records not yet converted; a rerun over a converted
// record finds its edge already applied. Child lists sort before the edge
// records, so a batch never reads the lists earlier batches wrote.
func MigrateChildLists(ctx context.Context, store kvtx.Store, prefix []byte) error {
	for start := prefix; start != nil; {
		var err error
		start, err = convertChildListBatch(ctx, store, prefix, start)
		if err != nil {
			return err
		}
	}
	return nil
}

// convertChildListBatch converts up to convertBatchEdges records under prefix
// at or after start in one commit, and returns the key to resume from, or nil
// when none remain.
func convertChildListBatch(ctx context.Context, store kvtx.Store, prefix, start []byte) ([]byte, error) {
	// Read the next records.
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		return nil, err
	}
	defer tx.Discard()
	records, err := readLegacyRecords(ctx, tx, prefix, start)
	if err != nil || len(records) == 0 {
		return nil, err
	}

	// Delete each record the current encoding drops and collect the edges.
	var adds []block_gc.RefEdge
	for _, record := range records {
		keep, add, hasAdd, err := block_gc.ConvertFormat3Record(record[len(prefix):])
		if err != nil {
			return nil, err
		}
		if hasAdd {
			adds = append(adds, add)
		}
		if !keep {
			if err := tx.Delete(ctx, record); err != nil {
				return nil, err
			}
		}
	}

	// Write the edges in the same commit.
	graph := block_gc.NewRefGraph(kvtx_prefixer.NewPrefixer(kvtx.NewTxStore(tx), bytes.Clone(prefix)))
	if err := graph.ApplyRefBatch(ctx, adds, nil); err != nil {
		return nil, err
	}
	if err := kvtx.CommitOrdered(ctx, tx); err != nil || len(records) < convertBatchEdges {
		return nil, err
	}
	return keyAfter(records[len(records)-1]), nil
}
