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
