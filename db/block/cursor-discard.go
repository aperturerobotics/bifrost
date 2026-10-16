package block

// DiscardDetached unlinks an abandoned cursor position from the positions it
// referenced, so they stop listing it as a parent. The transaction root and
// positions still referenced by another parent survive. It returns the
// traversed references of a discarded position without discarding their
// targets. Positions cleared by an earlier write have none left.
// The caller must no longer use the discarded cursor position.
func (c *Cursor) DiscardDetached() map[uint32]*Cursor {
	// Reject an unusable cursor position.
	if c == nil || c.t == nil || c.pos == nil {
		return nil
	}
	c.t.mtx.Lock()
	defer c.t.mtx.Unlock()

	// Keep the root and positions another parent still references.
	if c.pos == c.t.root || len(c.pos.parents) != 0 {
		return nil
	}
	refs := make(map[uint32]*Cursor, len(c.pos.refHandles))
	for id, ref := range c.pos.refHandles {
		refs[id] = newCursor(c.t, ref.target, c.store)
	}
	c.clearRefHandles()
	return refs
}
