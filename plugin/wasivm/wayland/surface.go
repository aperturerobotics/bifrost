package wayland

import (
	"image"
	"slices"
)

// surface is a wl_surface. Its pending state builds up until commit, which
// copies the attached buffer's damaged pixels into the surface's image and
// releases the buffer. A role places it on the screen: an xdg surface is a
// window or a popup, a layer surface is part of the desktop, and a
// subsurface is drawn on its parent.
type surface struct {
	// c is the client.
	c *client
	// id is the surface's object id.
	id uint32

	// attached is set when the pending state attaches a buffer, or none.
	attached bool
	// buffer is the attached buffer, or nil to unmap.
	buffer *shmBuffer
	// damaged is the pending damage in surface coordinates.
	damaged image.Rectangle
	// callbacks is the pending frame callbacks.
	callbacks []uint32

	// xdg is the surface's xdg_surface role, or nil.
	xdg *xdgSurface
	// sub is the surface's wl_subsurface role, or nil.
	sub *subsurface
	// layer is the surface's zwlr_layer_surface_v1 role, or nil.
	layer *layerSurface

	// The compositor's lock guards the fields below. The client's serve
	// goroutine alone writes them, so it reads them without the lock.

	// size is the image's size: the committed buffer's, clipped to the
	// screen.
	size image.Point
	// image is the committed pixels in premultiplied RGBA, or nil with no
	// buffer.
	image []byte
	// opaque is set when every pixel of the image is opaque.
	opaque bool
	// parent is the surface a subsurface or a popup is drawn on, or nil.
	parent *surface
	// stack is the surface and its subsurfaces, bottom to top.
	stack []*surface
	// popups is the popups on the surface, drawn above its stack, bottom to
	// top.
	popups []*surface
}

// subsurface is the wl_subsurface role: a surface drawn at an offset on its
// parent, in the parent's stack. Its own commits apply at once, as in the
// desynchronized mode, while its position applies at the parent's commit.
type subsurface struct {
	// id is the wl_subsurface's object id.
	id uint32
	// offset is the position in the parent's coordinates. The compositor's
	// lock guards it.
	offset image.Point
	// pending is the position set for the parent's next commit.
	pending image.Point
}

// newSurface makes a surface with no role.
func newSurface(c *client, id uint32) *surface {
	s := &surface{c: c, id: id}
	s.stack = []*surface{s}
	return s
}

// request handles a wl_surface request.
func (s *surface) request(op uint16, m *message) error {
	switch op {
	case surfaceDestroy:
		s.c.comp.update(s.detach)
		s.c.destroy(s.id)
	case surfaceAttach:
		bid, _, _ := m.uint(), m.int(), m.int()
		if m.err != nil {
			return nil
		}
		buf, ok := s.c.objects[bid].(*shmBuffer)
		if bid != 0 && !ok {
			return s.c.fail(s.id, errInvalidObject, "attached a non-buffer")
		}
		s.attached, s.buffer = true, buf
	case surfaceDamage, surfaceDamageBuffer:
		x, y, w, h := m.int(), m.int(), m.int(), m.int()
		if m.err == nil {
			s.damaged = s.damaged.Union(image.Rect(int(x), int(y), int(x)+int(w), int(y)+int(h)))
		}
	case surfaceFrame:
		id := m.uint()
		if m.err != nil {
			return nil
		}
		if err := s.c.add(id, noRequests); err != nil {
			return err
		}
		s.callbacks = append(s.callbacks, id)
	case surfaceCommit:
		return s.commit()
	}
	return nil
}

// commit applies the pending state.
func (s *surface) commit() error {
	// Check the role's pending state.
	if s.layer != nil {
		if err := s.layer.check(); err != nil {
			return err
		}
	}

	// Read the attached buffer's damaged pixels, or all of them when its
	// size changed, then release it.
	buf := s.buffer
	if buf != nil && buf.destroyed {
		buf = nil
	}
	size, area := s.size, image.Rectangle{}
	var pix []byte
	if s.attached && buf != nil {
		screen := s.c.comp.screen.Size()
		size = image.Pt(min(buf.width, screen.X), min(buf.height, screen.Y))
		area = s.damaged.Intersect(image.Rectangle{Max: size})
		if size != s.size || s.image == nil {
			area = image.Rectangle{Max: size}
		}
		var err error
		if pix, err = buf.read(area); err != nil {
			return s.c.fail(s.id, errImplementation, "read buffer: "+err.Error())
		}
		s.c.send(newEvent(buf.id, bufferRelease))
	}

	// Apply the state, then redraw the area it changed: the pixels alone,
	// or the whole window when its layout moved.
	comp := s.c.comp
	comp.update(func() {
		// Note where the surface's tree was.
		w := comp.rootOf(s)
		before := comp.bounds(w)

		// Apply the image, the subsurfaces' positions and the role.
		relayout := s.attached && (buf == nil || size != s.size)
		if s.attached {
			s.update(buf, size, area, pix)
		}
		for _, t := range s.stack {
			if t.sub != nil && t.sub.offset != t.sub.pending {
				t.sub.offset, relayout = t.sub.pending, true
			}
		}
		if s.xdg != nil && s.xdg.commit() {
			relayout = true
		}
		if s.layer != nil && s.layer.commit() {
			relayout = true
		}

		// Redraw, unless the tree mapped or unmapped, which redraws it.
		switch {
		case w != comp.rootOf(s):
		case relayout:
			comp.damage(before.Union(comp.bounds(w)))
			comp.refocusPointer()
		case w != nil:
			comp.damage(area.Add(s.origin()))
		}
	})

	// Clear the pending state and tell the client to draw its next frame.
	s.attached, s.buffer, s.damaged = false, nil, image.Rectangle{}
	now := comp.now()
	for _, id := range s.callbacks {
		s.c.done(id, now)
	}
	s.callbacks = nil
	return nil
}

// update replaces the image when the surface lost its buffer or changed
// size, then copies pix, the pixels of area, into it. The caller holds the
// compositor's lock.
func (s *surface) update(buf *shmBuffer, size image.Point, area image.Rectangle, pix []byte) {
	// Clear or resize the image.
	switch {
	case buf == nil:
		s.size, s.image = image.Point{}, nil
		return
	case size != s.size || s.image == nil:
		s.size, s.image = size, make([]byte, 4*size.X*size.Y)
	}
	s.opaque = buf.opaque

	// Copy the area's rows.
	for y := range area.Dy() {
		to := 4 * ((area.Min.Y+y)*size.X + area.Min.X)
		copy(s.image[to:to+4*area.Dx()], pix[4*y*area.Dx():])
	}
}

// origin returns the screen position of the surface's top left corner. The
// caller holds the compositor's lock.
func (s *surface) origin() image.Point {
	switch {
	case s.sub != nil && s.parent != nil:
		return s.parent.origin().Add(s.sub.offset)
	case s.layer != nil:
		return s.layer.pos
	case s.xdg == nil:
		return image.Point{}
	case s.xdg.toplevel != nil:
		return s.xdg.toplevel.pos.Sub(s.xdg.geometry().Min)
	case s.xdg.popup != nil && s.parent != nil:
		at := s.parent.geometry().Min.Add(s.xdg.popup.place.Min)
		return s.parent.origin().Add(at).Sub(s.xdg.geometry().Min)
	}
	return image.Point{}
}

// geometry returns the window geometry of an xdg surface, or the whole
// surface for another role, in surface coordinates. The caller holds the
// compositor's lock or is the client's serve goroutine.
func (s *surface) geometry() image.Rectangle {
	if s.xdg != nil {
		return s.xdg.geometry()
	}
	return image.Rectangle{Max: s.size}
}

// mapped reports whether the surface shows as a window or in a layer. The
// caller holds the compositor's lock.
func (s *surface) mapped() bool {
	return (s.xdg != nil && s.xdg.mapped) || (s.layer != nil && s.layer.mapped)
}

// detach takes the destroyed surface off the screen: it leaves its parent,
// its subsurfaces and popups lose theirs, and it unmaps. The caller holds
// the compositor's lock.
func (s *surface) detach() {
	// Redraw where the surface's tree was.
	comp := s.c.comp
	comp.damage(comp.bounds(comp.rootOf(s)))

	// Unmap it, leave the parent and orphan the children.
	comp.unmap(s)
	s.unparent()
	for _, t := range s.stack {
		if t != s {
			t.parent = nil
		}
	}
	for _, p := range s.popups {
		p.parent = nil
	}
	s.stack, s.popups = []*surface{s}, nil
	comp.forget(s)
	comp.refocusPointer()
}

// unparent removes the surface from its parent's stack or popups. The
// caller holds the compositor's lock.
func (s *surface) unparent() {
	if p := s.parent; p != nil {
		p.stack = slices.DeleteFunc(p.stack, func(t *surface) bool { return t == s })
		p.popups = slices.DeleteFunc(p.popups, func(t *surface) bool { return t == s })
		s.parent = nil
	}
}

// ancestor reports whether a is s or holds s in its tree of subsurfaces and
// popups. The caller holds the compositor's lock.
func (s *surface) ancestor(a *surface) bool {
	for t := s; t != nil; t = t.parent {
		if t == a {
			return true
		}
	}
	return false
}

// getSubsurface gives s the subsurface role on parent, on top of the
// parent's stack, as wl_subsurface id.
func (c *client) getSubsurface(id uint32, s, parent *surface) error {
	// Check the role and that the parent is not inside s.
	if s.sub != nil || s.xdg != nil || s.layer != nil || parent.ancestor(s) {
		return c.fail(id, subcompositorErrBadSurface, "bad subsurface")
	}

	// Add it to the parent's stack.
	if err := c.add(id, requestFunc(func(op uint16, m *message) error {
		return s.subsurfaceRequest(op, m)
	})); err != nil {
		return err
	}
	c.comp.update(func() {
		s.sub = &subsurface{id: id}
		s.parent = parent
		parent.stack = append(parent.stack, s)
	})
	return nil
}

// subsurfaceRequest handles a wl_subsurface request. Restacking applies at
// once, not at the parent's commit.
func (s *surface) subsurfaceRequest(op uint16, m *message) error {
	switch op {
	case subsurfaceDestroy:
		// Take the surface off its parent.
		id := s.sub.id
		s.c.comp.update(func() {
			// Redraw where its tree was, then remove it.
			s.c.comp.damage(s.c.comp.bounds(s.c.comp.rootOf(s)))
			s.unparent()
			s.sub = nil
			s.c.comp.refocusPointer()
		})
		s.c.destroy(id)
	case subsurfaceSetPosition:
		x, y := m.int(), m.int()
		if m.err == nil {
			s.sub.pending = image.Pt(int(x), int(y))
		}
	case subsurfacePlaceAbove, subsurfacePlaceBelow:
		// Find the sibling: the parent or another of its subsurfaces.
		sid := m.uint()
		sibling, ok := s.c.objects[sid].(*surface)
		if m.err != nil || s.parent == nil {
			return nil
		}
		if !ok || sibling == s || !slices.Contains(s.parent.stack, sibling) {
			return s.c.fail(s.sub.id, subcompositorErrBadSurface, "bad sibling")
		}

		// Move the surface beside it.
		s.c.comp.update(func() {
			// Insert it above or below the sibling and redraw its tree.
			p := s.parent
			p.stack = slices.DeleteFunc(p.stack, func(t *surface) bool { return t == s })
			at := slices.Index(p.stack, sibling)
			if op == subsurfacePlaceAbove {
				at++
			}
			p.stack = slices.Insert(p.stack, at, s)
			s.c.comp.damage(s.c.comp.bounds(s.c.comp.rootOf(s)))
		})
	}
	return nil
}

// _ is a type assertion.
var _ object = (*surface)(nil)
