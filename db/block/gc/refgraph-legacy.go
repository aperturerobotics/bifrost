package block_gc

import (
	"encoding/binary"
	"fmt"
)

// LegacyRecordKey returns the current key of a graph record written by the
// earlier encoding, which spelled each node as a 4-byte big-endian length and
// its full IRI. record starts at the kind byte.
func LegacyRecordKey(record []byte) ([]byte, error) {
	// Split the kind byte from the nodes.
	if len(record) == 0 {
		return nil, fmt.Errorf("decode empty reference graph record")
	}
	kind, rest := record[0], record[1:]

	// Read each length-prefixed node, then encode them again.
	var nodes []string
	for len(rest) != 0 {
		if len(rest) < 4 || uint64(len(rest)-4) < uint64(binary.BigEndian.Uint32(rest)) { //nolint:gosec // the length is checked nonnegative first.
			return nil, fmt.Errorf("decode reference graph record %q", record)
		}
		n := binary.BigEndian.Uint32(rest)
		nodes = append(nodes, LegacyBlockIRI(string(rest[4:4+n])))
		rest = rest[4+n:]
	}
	return graphKey(kind, nodes...), nil
}
