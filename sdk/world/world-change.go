package s4wave_world

import (
	"github.com/s4wave/spacewave/db/block/quad"
	"github.com/s4wave/spacewave/db/world"
)

// NewWorldChange encodes the changes that reach seqno for WatchChanges.
func NewWorldChange(seqno uint64, changes *world.ChangeSet) *WorldChange {
	// Send an unknown change without its partial lists.
	if changes.Unknown {
		return &WorldChange{Seqno: seqno, Unknown: true}
	}

	// Encode each changed quad with its stored spelling.
	quads := make([]*quad.Quad, len(changes.Quads))
	for i, q := range changes.Quads {
		quads[i] = &quad.Quad{
			Subject:   q.GetSubject(),
			Predicate: q.GetPredicate(),
			Obj:       q.GetObj(),
			Label:     q.GetLabel(),
		}
	}
	return &WorldChange{Seqno: seqno, Keys: changes.Keys, Quads: quads}
}

// ToChangeSet decodes the changes the message lists.
func (c *WorldChange) ToChangeSet() *world.ChangeSet {
	// Decode an unknown change without its lists.
	if c.GetUnknown() {
		return world.NewUnknownChangeSet()
	}

	// Expose the quads through the World graph interface.
	quads := make([]world.GraphQuad, len(c.GetQuads()))
	for i, q := range c.GetQuads() {
		quads[i] = q
	}
	return &world.ChangeSet{Keys: c.GetKeys(), Quads: quads}
}
