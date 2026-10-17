package wayland

import (
	"image"
	"slices"
)

// region is a wl_region: the rectangles added and subtracted, in order. A
// point is in the region when the last rectangle containing it was added.
type region struct {
	// c is the client.
	c *client
	// id is the region's object id.
	id uint32
	// ops is the rectangles, oldest first.
	ops []regionOp
}

// regionOp is one rectangle added to or subtracted from a region.
type regionOp struct {
	// rect is the rectangle in surface coordinates.
	rect image.Rectangle
	// add is set when the rectangle was added, clear when subtracted.
	add bool
}

// request handles a wl_region request.
func (r *region) request(op uint16, m *message) error {
	switch op {
	case regionDestroy:
		r.c.destroy(r.id)
	case regionAdd, regionSubtract:
		x, y, w, h := m.int(), m.int(), m.int(), m.int()
		if m.err == nil {
			rect := image.Rect(int(x), int(y), int(x)+int(w), int(y)+int(h))
			r.ops = append(r.ops, regionOp{rect: rect, add: op == regionAdd})
		}
	}
	return nil
}

// contains reports whether pt is in the region.
func (r *region) contains(pt image.Point) bool {
	for i := len(r.ops) - 1; i >= 0; i-- {
		if pt.In(r.ops[i].rect) {
			return r.ops[i].add
		}
	}
	return false
}

// clone returns a copy of the region's rectangles that later requests do
// not change.
func (r *region) clone() *region {
	return &region{ops: slices.Clone(r.ops)}
}

// _ is a type assertion.
var _ object = (*region)(nil)
