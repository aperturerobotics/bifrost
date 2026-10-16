package mysql

import "github.com/dolthub/go-mysql-server/sql"

// tableLookupPartition holds rows fetched by one index lookup.
type tableLookupPartition struct {
	// rows contains only the matching stored rows.
	rows []sql.Row
}

// Key identifies the single logical partition of an index lookup.
func (p *tableLookupPartition) Key() []byte {
	return []byte("lookup")
}

// _ is a type assertion.
var _ sql.Partition = (*tableLookupPartition)(nil)
