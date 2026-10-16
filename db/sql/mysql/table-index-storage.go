package mysql

import (
	"bytes"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/kvtx"
)

// indexTree opens the secondary index's tree within the table transaction.
func (t *Table) indexTree(ctx *sql.Context, index *TableIndex, write bool) (kvtx.BlockTx, error) {
	for i, candidate := range t.root.GetIndexes() {
		if candidate == index {
			cursor := t.bcs.FollowSubBlock(8).FollowSubBlock(uint32(i)).FollowSubBlock(5) //nolint:gosec
			return index.GetKeyValue().BuildKvTransaction(GetDbContext(ctx), cursor, write)
		}
	}
	return nil, errors.New("index is not attached to table")
}

// secondaryKey encodes index columns, suffixing nonunique and NULL keys by row.
func (t *Table) secondaryKey(ctx *sql.Context, index *TableIndex, row sql.Row, rowKey []byte) ([]byte, bool, error) {
	// Encode indexed values using the table's canonical column representation.
	ordinals, err := t.indexColumnOrdinals(index.GetColumns())
	if err != nil {
		return nil, false, err
	}
	key, hasNull, err := t.columnKey(ctx, row, ordinals, index.GetColumns())
	if err != nil {
		return nil, false, err
	}

	// Keep each duplicate separately while unique non-null values have one key.
	unique := index.GetUnique() && !hasNull
	if !unique {
		key = append(key, rowKey...)
	}
	return key, unique, nil
}

// writeIndexes adds or removes every secondary reference for a stored row.
func (t *Table) writeIndexes(ctx *sql.Context, row sql.Row, rowKey []byte, remove bool) error {
	for _, index := range t.root.GetIndexes() {
		// Compute the index entry before opening its tree.
		key, _, err := t.secondaryKey(ctx, index, row, rowKey)
		if err != nil {
			return err
		}
		tree, err := t.indexTree(ctx, index, true)
		if err != nil {
			return err
		}

		// Mutate the tree through the enclosing block transaction.
		if remove {
			err = tree.Delete(GetDbContext(ctx), key)
		} else {
			err = tree.Set(GetDbContext(ctx), key, rowKey)
		}
		tree.Discard()
		if err != nil {
			return err
		}
	}
	return nil
}

// rowAtKey reads one primary tree key and decodes only that row.
func (t *Table) rowAtKey(ctx *sql.Context, key []byte) (sql.Row, error) {
	// Open the partition that stores every row.
	part, err := t.rowPartition()
	if err != nil {
		return nil, err
	}
	tree, err := part.BuildTreeTx(GetDbContext(ctx), false, false)
	if err != nil {
		return nil, err
	}
	defer tree.Discard()

	// Resolve the single row reference without opening an iterator.
	cursor, err := tree.GetCursorAtKey(GetDbContext(ctx), key)
	if err != nil || cursor == nil {
		return nil, err
	}
	return t.decodeRow(ctx, cursor)
}

// decodeRow decodes the stored row at cursor, enforcing the table schema.
func (t *Table) decodeRow(ctx *sql.Context, cursor *block.Cursor) (sql.Row, error) {
	// Detach the cursor so the decoded row outlives the tree transaction.
	cursor = cursor.DetachTransaction()
	stored, err := UnmarshalTableRow(GetDbContext(ctx), cursor)
	if err != nil {
		return nil, err
	}

	// Enforce the same row/schema contract as ordinary table iteration.
	row, err := stored.FetchSqlRow(GetDbContext(ctx), cursor)
	if err != nil {
		return nil, err
	}
	if err := t.schema.CheckRow(ctx, row); err != nil {
		return nil, err
	}
	return row, nil
}

// scanRows returns every stored row. A null-safe join probes with a NULL-excluding
// range before it binds a key, and the engine keeps its predicate filters
// because PreciseMatch is false.
func (t *Table) scanRows(ctx *sql.Context) ([]sql.Row, error) {
	// Iterate the one partition that stores every row.
	part, err := t.rowPartition()
	if err != nil {
		return nil, err
	}
	iter, err := part.IterateRows(ctx)
	if err != nil {
		return nil, err
	}
	return sql.RowIterToRows(ctx, iter)
}

// lookupRows resolves equality ranges through the primary or secondary tree.
func (t *Table) lookupRows(ctx *sql.Context, lookup sql.IndexLookup) ([]sql.Row, error) {
	// Require ranges supported by the declared index.
	index, ok := lookup.Index.(*tableIndex)
	if !ok {
		return nil, ErrUnexpectedType
	}
	ranges, ok := lookup.Ranges.(sql.MySQLRangeCollection)
	if !ok {
		return nil, ErrUnexpectedType
	}
	rows := make([]sql.Row, 0, len(ranges))
	seen := make(map[string]struct{}, len(ranges))
	for _, r := range ranges {
		// Convert the full equality tuple into its encoded tree key.
		values, ok := index.pointValues(ctx, r)
		if !ok {
			return t.scanRows(ctx)
		}
		ordinals, err := t.indexColumnOrdinals(index.columns)
		if err != nil {
			return nil, err
		}
		key, hasNull, err := t.encodeColumns(ctx, ordinals, values, index.columns)
		if err != nil {
			return nil, err
		}

		// Fetch primary keys directly or resolve them through a secondary tree.
		keys := [][]byte{key}
		if !index.primary {
			stored := t.indexByName(index.name)
			tree, err := t.indexTree(ctx, stored, false)
			if err != nil {
				return nil, err
			}
			keys, err = secondaryRowKeys(ctx, tree, key, index.unique && !hasNull)
			tree.Discard()
			if err != nil {
				return nil, err
			}
		}

		// Decode each matching row once even when equality ranges overlap.
		for _, primary := range keys {
			if _, ok := seen[string(primary)]; ok {
				continue
			}
			row, err := t.rowAtKey(ctx, primary)
			if err != nil {
				return nil, err
			}
			if row == nil {
				continue
			}

			// Prefix indexes and coerced probes may resolve a wider equality bucket.
			// Check only fetched candidates, never unrelated table rows.
			matches := true
			for i, ord := range ordinals {
				cmp, err := r[i].Typ.Compare(ctx, row[ord], values[i])
				if err != nil {
					return nil, err
				}
				if cmp != 0 {
					matches = false
					break
				}
			}
			if matches {
				seen[string(primary)] = struct{}{}
				rows = append(rows, row)
			}
		}
	}
	return rows, nil
}

// secondaryRowKeys resolves one unique key or the duplicate keys of one prefix.
func secondaryRowKeys(ctx *sql.Context, tree kvtx.BlockTx, key []byte, unique bool) ([][]byte, error) {
	// A unique index equality reads exactly one index entry.
	if unique {
		value, found, err := tree.Get(GetDbContext(ctx), key)
		if err != nil || !found {
			return nil, err
		}
		return [][]byte{value}, nil
	}

	// Nonunique equality walks only entries with the complete encoded tuple.
	it := tree.Iterate(GetDbContext(ctx), key, true, false)
	defer it.Close()
	var keys [][]byte
	for it.Next() {
		value, err := it.Value()
		if err != nil {
			return nil, err
		}
		keys = append(keys, bytes.Clone(value))
	}
	return keys, it.Err()
}
