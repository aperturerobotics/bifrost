package wayland

import (
	"image"
)

// surface is a wl_surface. Its pending state builds up until commit, which
// copies the attached buffer's damaged pixels into the surface's image,
// releases the buffer and maps or unmaps its toplevel.
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

	// xdgSurface is the xdg_surface's id, which receives configure events.
	xdgSurface uint32
	// toplevel is the xdg_toplevel's id, or zero when the surface has none.
	toplevel uint32
	// configured is set once the toplevel's first configure is sent.
	configured bool
	// mapped is set while the toplevel shows on the compositor's stack.
	mapped bool

	// size is the image's size: the committed buffer's, clipped to the
	// screen. The serve goroutine writes it and image under the
	// compositor's lock, so it reads them without the lock.
	size image.Point
	// image is the committed pixels in RGBA, or nil with no buffer.
	image []byte
}

// request handles a wl_surface request.
func (s *surface) request(op uint16, m *message) error {
	switch op {
	case surfaceDestroy:
		s.c.comp.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
			if s.c.comp.focus == s {
				s.c.comp.focus = nil
			}
			s.c.comp.show(s, false, broadcast)
		})
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

	// Update the image and the toplevel.
	s.c.comp.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		if s.attached {
			s.update(buf != nil, size, area, pix)
		}
		if s.toplevel != 0 {
			s.present(area, broadcast)
		}
	})

	// Clear the pending state and tell the client to draw its next frame.
	s.attached, s.buffer, s.damaged = false, nil, image.Rectangle{}
	now := s.c.comp.now()
	for _, id := range s.callbacks {
		s.c.done(id, now)
	}
	s.callbacks = nil
	return nil
}

// update replaces the image when the surface lost its buffer or changed
// size, then copies pix, the pixels of area, into it. The caller holds the
// compositor's lock.
func (s *surface) update(buffered bool, size image.Point, area image.Rectangle, pix []byte) {
	// Clear or resize the image.
	switch {
	case !buffered:
		s.size, s.image = image.Point{}, nil
		return
	case size != s.size || s.image == nil:
		s.size, s.image = size, make([]byte, 4*size.X*size.Y)
	}

	// Copy the area's rows.
	for y := range area.Dy() {
		to := 4 * ((area.Min.Y+y)*size.X + area.Min.X)
		copy(s.image[to:to+4*area.Dx()], pix[4*y*area.Dx():])
	}
}

// present sends a new toplevel its first configure, then maps or unmaps it
// to match its image, or redraws its changed area when it is shown. The
// caller holds the compositor's lock.
func (s *surface) present(area image.Rectangle, broadcast func()) {
	// Configure the toplevel to fill the screen, fullscreen and active.
	comp := s.c.comp
	if !s.configured {
		s.configured = true
		screen := comp.screen.Size()
		s.c.send(newEvent(s.toplevel, toplevelConfigure).
			int(int32(screen.X)).int(int32(screen.Y)). //nolint:gosec // screen sizes are small.
			uints(toplevelFullscreen, toplevelActivated))
		s.c.send(newEvent(s.xdgSurface, xdgSurfaceConfigure).uint(comp.nextSerial()))
	}

	// Map, unmap or redraw it.
	switch mapped := s.image != nil; {
	case mapped != s.mapped:
		s.mapped = mapped
		comp.show(s, mapped, broadcast)
	case mapped && comp.shown() == s:
		comp.damage(area, broadcast)
	}
}

// xdgSurfaceRequest handles a request to the surface's xdg_surface id.
func (s *surface) xdgSurfaceRequest(id uint32, op uint16, m *message) error {
	switch op {
	case xdgSurfaceDestroy:
		s.c.destroy(id)
	case xdgSurfaceGetToplevel:
		tid := m.uint()
		if m.err != nil {
			return nil
		}
		if s.toplevel != 0 {
			return s.c.fail(id, errInvalidObject, "surface already has a toplevel")
		}
		if err := s.c.add(tid, requestFunc(s.toplevelRequest)); err != nil {
			return err
		}
		s.toplevel, s.xdgSurface = tid, id
	case xdgSurfaceGetPopup:
		// Dismiss the popup at once: popups are never drawn.
		pid, _, _ := m.uint(), m.uint(), m.uint()
		if m.err != nil {
			return nil
		}
		if err := s.c.add(pid, s.c.inert(pid, popupDestroy)); err != nil {
			return err
		}
		s.c.send(newEvent(pid, popupDone))
	}
	return nil
}

// toplevelRequest handles an xdg_toplevel request. Only destroy, which
// unmaps the toplevel, has an effect: every toplevel fills the screen.
func (s *surface) toplevelRequest(op uint16, _ *message) error {
	// Ignore the window management requests.
	if op != toplevelDestroy {
		return nil
	}

	// Unmap the toplevel, then forget it.
	s.c.comp.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		s.mapped = false
		s.c.comp.show(s, false, broadcast)
	})
	s.c.destroy(s.toplevel)
	s.toplevel, s.configured = 0, false
	return nil
}

// _ is a type assertion.
var _ object = (*surface)(nil)
