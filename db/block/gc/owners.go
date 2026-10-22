package block_gc

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// OwnerCount counts the owners of one kind a mark seeded from.
type OwnerCount struct {
	// Nodes is the number of owners of the kind.
	Nodes int
	// Edges is the number of their forward records.
	Edges int
}

// countOwner counts a forward record from owner. Forward records arrive in key
// order, grouped by subject, so an owner is new when it differs from the last.
func (m *BlockMark) countOwner(owner, last string) {
	// Count the record under the owner's kind, the name before its colon.
	if m.Owners == nil {
		m.Owners = make(map[string]OwnerCount)
	}
	kind, _, _ := strings.Cut(owner, ":")
	c := m.Owners[kind]
	c.Edges++
	if owner != last {
		c.Nodes++
	}
	m.Owners[kind] = c
}

// FormatOwners spells owner counts as kind=nodes/edges, sorted by kind.
func FormatOwners(owners map[string]OwnerCount) string {
	var b strings.Builder
	for _, kind := range slices.Sorted(maps.Keys(owners)) {
		if b.Len() != 0 {
			b.WriteByte(' ')
		}
		c := owners[kind]
		fmt.Fprintf(&b, "%s=%d/%d", kind, c.Nodes, c.Edges)
	}
	return b.String()
}
