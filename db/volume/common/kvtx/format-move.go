package kvtx

import (
	"context"
	"slices"

	block_gc "github.com/s4wave/spacewave/db/block/gc"
	"github.com/s4wave/spacewave/db/kvtx"
)

// AdoptRefGraph moves a reference graph kept beside a Volume into the Volume.
// root holds the Volume's store under volume, the graph's version 2 records
// under legacy and its later records under current. They move to the Volume's
// graph prefixes in the same encoding, so opening the Volume afterwards brings
// them to the current format by the store's recorded version. The node
// inventory of the retired mark-sweep collector is dropped. A rerun after a
// crash resumes the move. A root holding no graph costs two scans.
func AdoptRefGraph(ctx context.Context, root kvtx.Store, volume, legacy, current []byte) error {
	if err := movePrefix(ctx, root, legacy, slices.Concat(volume, legacyRefGraphPrefix), nil); err != nil {
		return err
	}
	return movePrefix(ctx, root, current, slices.Concat(volume, refGraphPrefix), block_gc.InventoryRecord)
}

// movePrefix moves every record under from to the same suffix under to,
// deleting instead each record whose suffix drop, if set, reports. Each batch
// writes its records and deletes the originals in one commit, so a rerun
// resumes with the records not yet moved, and seeks past the records earlier
// batches moved. Neither prefix may contain the other.
func movePrefix(ctx context.Context, store kvtx.Store, from, to []byte, drop func(suffix []byte) bool) error {
	for start := from; start != nil; {
		var err error
		start, err = moveBatch(ctx, store, from, to, start, drop)
		if err != nil {
			return err
		}
	}
	return nil
}

// moveBatch moves or drops up to migrateBatchBytes of the records under from
// at or after start in one commit, and returns the key to resume from, or nil
// when none remain.
func moveBatch(ctx context.Context, store kvtx.Store, from, to, start []byte, drop func(suffix []byte) bool) ([]byte, error) {
	// Collect the next batch under the writer.
	tx, err := store.NewTransaction(ctx, true)
	if err != nil {
		return nil, err
	}
	defer tx.Discard()
	moves, more, err := collectMoves(ctx, tx, from, start)
	if err != nil || len(moves) == 0 {
		return nil, err
	}

	// Rewrite each kept record under to and delete the original.
	for _, m := range moves {
		suffix := m.key[len(from):]
		if drop == nil || !drop(suffix) {
			if err := tx.Set(ctx, slices.Concat(to, suffix), m.value); err != nil {
				return nil, err
			}
		}
		if err := tx.Delete(ctx, m.key); err != nil {
			return nil, err
		}
	}
	if err := kvtx.CommitOrdered(ctx, tx); err != nil || !more {
		return nil, err
	}
	return keyAfter(moves[len(moves)-1].key), nil
}
