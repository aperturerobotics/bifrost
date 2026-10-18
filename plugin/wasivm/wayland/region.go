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
	for _, v := range slices.Backward(r.ops) {
		if pt.In(v.rect) {
			return v.add
		}
	}
	return false
}

// covers reports whether the region holds every point of rect. Like
// contains, it lets the newest rectangle over each point decide.
func (r *region) covers(rect image.Rectangle) bool {
	left := []image.Rectangle{rect}
	for _, v := range slices.Backward(r.ops) {
		var next []image.Rectangle
		for _, l := range left {
			if !v.add && l.Overlaps(v.rect) {
				return false
			}
			next = append(next, subtract(l, v.rect)...)
		}
		if left = next; len(left) == 0 {
			return true
		}
	}
	return rect.Empty()
}

// clone returns a copy of the region's rectangles that later requests do
// not change.
func (r *region) clone() *region {
	return &region{ops: slices.Clone(r.ops)}
}

// subtract returns the parts of a outside b, as up to four rectangles.
func subtract(a, b image.Rectangle) []image.Rectangle {
	// Keep a whole when b misses it.
	in := a.Intersect(b)
	switch {
	case a.Empty():
		return nil
	case in.Empty():
		return []image.Rectangle{a}
	}

	// Take the bands above and below the overlap, then the sides beside it.
	var out []image.Rectangle
	for _, r := range [...]image.Rectangle{
		{Min: a.Min, Max: image.Pt(a.Max.X, in.Min.Y)},
		{Min: image.Pt(a.Min.X, in.Max.Y), Max: a.Max},
		{Min: image.Pt(a.Min.X, in.Min.Y), Max: image.Pt(in.Min.X, in.Max.Y)},
		{Min: image.Pt(in.Max.X, in.Min.Y), Max: image.Pt(a.Max.X, in.Max.Y)},
	} {
		if !r.Empty() {
			out = append(out, r)
		}
	}
	return out
}

// _ is a type assertion.
var _ object = (*region)(nil)
