package mysql

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// EncodeTableKey encodes SQL equality keys with length-delimited components.
// Values are converted to their column type; text uses the column's collation.
// The encoding preserves equality, not SQL sort order.
// The second result reports whether any component is NULL. Row and schema must
// have equal lengths; columns is nil or supplies one prefix definition per column.
func EncodeTableKey(ctx *sql.Context, schema sql.Schema, row sql.Row, columns []*TableIndexColumn) ([]byte, bool, error) {
	// Encode each component independently so NULL and composite keys cannot collide.
	key := make([]byte, 0, len(row)*8)
	hasNull := false
	for ord, col := range schema {
		// Represent NULL separately from every non-null value, including empty text.
		if row[ord] == nil {
			key = append(key, 0)
			hasNull = true
			continue
		}

		// Convert values before encoding so probes and stored rows share an identity.
		value, _, err := col.Type.Convert(ctx, row[ord])
		if err != nil {
			return nil, false, err
		}
		value, err = sql.UnwrapAny(ctx, value)
		if err != nil {
			return nil, false, err
		}

		// Normalize signed zero, which SQL compares equal to positive zero.
		switch v := value.(type) {
		case float32:
			if v == 0 {
				value = float32(0)
			}
		case float64:
			if v == 0 {
				value = float64(0)
			}
		}

		// Use SQL wire values for scalar types and collation weights for strings.
		wire, err := col.Type.SQL(ctx, nil, value)
		if err != nil {
			return nil, false, err
		}
		data := wire.Raw()
		if typ, ok := col.Type.(sql.StringType); ok && typ.IsStringType() {
			// Apply index prefix lengths in characters, or bytes for binary columns.
			text := string(data)
			if len(columns) != 0 && columns[ord].GetLength() > 0 {
				length := int(columns[ord].GetLength())
				if types.IsBinaryType(typ) {
					text = text[:min(length, len(text))]
				} else {
					runes := []rune(text)
					text = string(runes[:min(length, len(runes))])
				}
			}

			// CHAR comparison discards trailing spaces before applying the collation.
			if types.IsChar(typ) {
				text = strings.TrimRight(text, " ")
			}
			var weights bytes.Buffer
			if err := typ.Collation().WriteWeightString(&weights, text); err != nil {
				return nil, false, err
			}
			data = weights.Bytes()
		}

		// Frame the non-null component with its byte length.
		key = append(key, 1)
		key = binary.AppendUvarint(key, uint64(len(data)))
		key = append(key, data...)
	}
	return key, hasNull, nil
}

// rowKey encodes the primary columns, or all columns for a keyless row prefix.
func (t *Table) rowKey(ctx *sql.Context, row sql.Row) ([]byte, error) {
	if len(t.schema.PkOrdinals) == 0 {
		key, _, err := EncodeTableKey(ctx, t.schema.Schema, row, nil)
		return key, err
	}
	key, _, err := t.columnKey(ctx, row, t.schema.PkOrdinals, nil)
	return key, err
}

// columnKey encodes the selected columns in index order.
func (t *Table) columnKey(ctx *sql.Context, row sql.Row, ordinals []int, columns []*TableIndexColumn) ([]byte, bool, error) {
	// Project the row and schema onto the index columns.
	schema := make(sql.Schema, len(ordinals))
	values := make(sql.Row, len(ordinals))
	for i, ord := range ordinals {
		schema[i] = t.schema.Schema[ord]
		values[i] = row[ord]
	}
	return EncodeTableKey(ctx, schema, values, columns)
}
