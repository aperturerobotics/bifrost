package mysql

import (
	"bytes"
	"math"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// TestTableKeyEquality verifies canonical conversions, collation and tuple framing.
func TestTableKeyEquality(t *testing.T) {
	// Compare encoded equality with the SQL column comparison contract.
	ctx := sql.NewContext(t.Context())
	text := types.MustCreateString(types.Text.Type(), 64, sql.Collation_utf8mb4_general_ci)
	cases := []struct {
		// name identifies the equality contract.
		name string
		// schema defines the compared columns.
		schema sql.Schema
		// left is the stored row.
		left sql.Row
		// right is a lookup tuple.
		right sql.Row
		// equal indicates whether the tuples compare equal.
		equal bool
	}{
		{"integer conversion", sql.Schema{{Type: types.Int64}}, sql.Row{int64(10)}, sql.Row{int32(10)}, true},
		{"collation", sql.Schema{{Type: text}}, sql.Row{"Hello"}, sql.Row{"hello"}, true},
		{"tuple framing", sql.Schema{{Type: types.Text}, {Type: types.Text}}, sql.Row{"a", "bc"}, sql.Row{"ab", "c"}, false},
		{"null and empty", sql.Schema{{Type: types.Text}}, sql.Row{nil}, sql.Row{""}, false},
		{"signed zero", sql.Schema{{Type: types.Float64}}, sql.Row{float64(0)}, sql.Row{math.Copysign(0, -1)}, true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			// Compare canonical tuples rather than the source Go representations.
			left, _, err := encodeTableKey(ctx, test.schema, test.left, nil)
			if err != nil {
				t.Fatal(err)
			}
			right, _, err := encodeTableKey(ctx, test.schema, test.right, nil)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(left, right) != test.equal {
				t.Fatalf("encoded equality = %v, want %v", bytes.Equal(left, right), test.equal)
			}
		})
	}
}

// TestNumberKeyOrder verifies that number keys sort as SQL sorts their values.
func TestNumberKeyOrder(t *testing.T) {
	ctx := sql.NewContext(t.Context())
	for _, test := range []struct {
		// typ is the column type.
		typ sql.Type
		// values are in ascending SQL order.
		values []any
	}{
		{types.Int64, []any{nil, int64(math.MinInt64), int64(-1), int64(0), int64(1), int64(math.MaxInt64)}},
		{types.Int8, []any{nil, int8(-128), int8(-1), int8(0), int8(127)}},
		{types.Uint64, []any{nil, uint64(0), uint64(1), uint64(1 << 63), uint64(math.MaxUint64)}},
		{types.Float64, []any{nil, math.Inf(-1), -1e300, -1.5, -math.SmallestNonzeroFloat64, float64(0), math.SmallestNonzeroFloat64, 1.5, math.Inf(1)}},
		{types.Float32, []any{nil, float32(-3.5), float32(-1), float32(0), float32(2)}},
	} {
		t.Run(test.typ.String(), func(t *testing.T) {
			// Each key sorts after the key of the value before it.
			schema := sql.Schema{{Type: test.typ}}
			var prev []byte
			for i, value := range test.values {
				key, _, err := encodeTableKey(ctx, schema, sql.Row{value}, nil)
				if err != nil {
					t.Fatal(err)
				}
				if i != 0 && bytes.Compare(prev, key) >= 0 {
					t.Fatalf("key of %v sorts before or at key of %v", value, test.values[i-1])
				}
				prev = key
			}
		})
	}
}
