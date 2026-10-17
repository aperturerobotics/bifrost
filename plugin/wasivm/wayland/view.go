package wayland

import (
	"image"
	"slices"
)

// maxDamage and maxCopies bound a view's lists. Past maxDamage areas the view
// draws their bounding box instead; past maxCopies moves it draws each
// move's destination instead of copying.
const (
	maxDamage = 16
	maxCopies = 16
)

// view is what one viewer's copy of the screen lacks: the moves to make on
// it, in order, then the areas to draw again from the screen. The compositor's
// lock guards it.
type view struct {
	// copies is the moves, oldest first.
	copies []move
	// damage is the areas to draw after the moves.
	damage []image.Rectangle
}

// move is an area of the screen that moves by an offset, both ends inside
// the screen.
type move struct {
	// src is the area before the move.
	src image.Rectangle
	// by is the offset it moves by.
	by image.Point
}

// dst returns the area after the move.
func (m move) dst() image.Rectangle {
	return m.src.Add(m.by)
}

// empty reports whether the viewer's copy is current.
func (v *view) empty() bool {
	return len(v.copies) == 0 && len(v.damage) == 0
}

// add marks the area r, inside the screen, to draw again.
func (v *view) add(r image.Rectangle) {
	// Skip an area already marked and drop those it covers.
	if r.Empty() {
		return
	}
	for _, d := range v.damage {
		if r.In(d) {
			return
		}
	}
	v.damage = slices.DeleteFunc(v.damage, func(d image.Rectangle) bool { return d.In(r) })
	v.damage = append(v.damage, r)

	// Fold a long list into its bounding box.
	if len(v.damage) > maxDamage {
		var u image.Rectangle
		for _, d := range v.damage {
			u = u.Union(d)
		}
		v.damage = append(v.damage[:0], u)
	}
}

// copy moves the area src of the viewer's copy by by. The caller marks the
// areas the move uncovers.
func (v *view) copy(src image.Rectangle, by image.Point) {
	// The stale pixels inside src land stale at the destination.
	var stale []image.Rectangle
	for _, d := range v.damage {
		if in := d.Intersect(src); !in.Empty() {
			stale = append(stale, in.Add(by))
		}
	}
	for _, r := range stale {
		v.add(r)
	}

	// Fold a move of the last move's destination into the last move. The
	// part of that destination the folded move leaves is not copied, so it
	// is drawn instead.
	if n := len(v.copies); n != 0 {
		if last := &v.copies[n-1]; src.In(last.dst()) {
			for _, r := range subtract(last.dst(), src.Add(by)) {
				v.add(r)
			}
			last.src, last.by = src.Sub(last.by), last.by.Add(by)
			return
		}
	}

	// Queue the move, or draw the destinations of too many.
	v.copies = append(v.copies, move{src: src, by: by})
	if len(v.copies) > maxCopies {
		for _, m := range v.copies {
			v.add(m.dst())
		}
		v.copies = v.copies[:0]
	}
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
