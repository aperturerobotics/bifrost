package wayland

import (
	"image"
	"slices"

	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// background is the desktop's color behind the windows, in RGBA.
var background = [4]byte{0x2e, 0x34, 0x40, 0xff}

// cascadeStep is the offset between the corners of successive new windows,
// and cascadeSlots is how many fit before the cascade starts over.
const (
	cascadeStep  = 32
	cascadeSlots = 8
)

// mapWindow shows the toplevel s on top of the stack, placing it when it is
// new, and activates it. The caller holds the lock.
func (c *Compositor) mapWindow(s *surface) {
	// Place a new window on the cascade.
	x, t := s.xdg, s.xdg.toplevel
	if !t.placed {
		t.placed = true
		t.pos = c.screen.Min.Add(image.Pt(cascadeStep, cascadeStep).Mul(1 + c.placed%cascadeSlots))
		c.placed++
	}

	// Show and activate it.
	x.mapped = true
	c.windows = append(c.windows, s)
	c.damage(c.bounds(s))
	c.activate(s)
	c.refocusPointer()
}

// unmap takes the toplevel s off the stack, dismisses its popups and
// activates the window below. The client must configure it again before it
// maps. The caller holds the lock.
func (c *Compositor) unmap(s *surface) {
	// Take it off the stack.
	x := s.xdg
	if !x.mapped {
		return
	}
	c.damage(c.bounds(s))
	for _, p := range slices.Clone(s.popups) {
		c.dismiss(p, true)
	}
	x.mapped, x.configured = false, false
	c.windows = slices.DeleteFunc(c.windows, func(w *surface) bool { return w == s })

	// Move the focus.
	if c.grab != nil && c.grab.w == s {
		c.grab = nil
	}
	if c.active == s {
		c.activate(nil)
		if n := len(c.windows); n != 0 {
			c.activate(c.windows[n-1])
		}
	}
	c.refocusPointer()
}

// raise moves the window w to the top of the stack. The caller holds the
// lock.
func (c *Compositor) raise(w *surface) {
	if n := len(c.windows); n == 0 || c.windows[n-1] == w {
		return
	}
	c.windows = slices.DeleteFunc(c.windows, func(t *surface) bool { return t == w })
	c.windows = append(c.windows, w)
	c.damage(c.bounds(w))
}

// setState maximizes, fills the screen with, or restores the window s, and
// configures it at the new size. The caller holds the lock.
func (c *Compositor) setState(s *surface, maximized, fullscreen bool) {
	// Remember the window's place before it first leaves it.
	t := s.xdg.toplevel
	wasNormal := !t.maximized && !t.fullscreen
	if wasNormal && (maximized || fullscreen) {
		t.restore = image.Rectangle{Min: t.pos, Max: t.pos.Add(s.xdg.geometry().Size())}
	}
	t.maximized, t.fullscreen = maximized, fullscreen

	// Fill the screen, or go back to the remembered place.
	switch {
	case maximized || fullscreen:
		t.pos, t.size = c.screen.Min, c.screen.Size()
	case !wasNormal:
		t.pos, t.size = t.restore.Min, t.restore.Size()
	}
	if s.xdg.configured {
		s.xdg.configure()
	}
}

// windowOf returns the mapped window whose tree holds s, or nil. The caller
// holds the lock.
func (c *Compositor) windowOf(s *surface) *surface {
	// Find the root of the tree, which is the window when it is a mapped
	// toplevel.
	if s == nil {
		return nil
	}
	root := s
	for root.parent != nil {
		root = root.parent
	}
	if root.xdg == nil || !root.xdg.mapped {
		return nil
	}
	return root
}

// bounds returns the screen area of the tree of s, or nothing for nil. The
// caller holds the lock.
func (c *Compositor) bounds(s *surface) image.Rectangle {
	var r image.Rectangle
	if s == nil {
		return r
	}
	c.walk(s, s.origin(), func(t *surface, at image.Point) bool {
		r = r.Union(image.Rectangle{Min: at, Max: at.Add(t.size)})
		return false
	})
	return r
}

// walk calls fn on each surface of the tree of s, bottom to top, with its
// screen position at when s sits at at, until fn returns true. It reports
// whether fn did. The caller holds the lock.
func (c *Compositor) walk(s *surface, at image.Point, fn func(t *surface, at image.Point) bool) bool {
	for _, t := range s.stack {
		if t == s {
			if s.image != nil && fn(s, at) {
				return true
			}
		} else if c.walk(t, at.Add(t.sub.offset), fn) {
			return true
		}
	}
	for _, p := range s.popups {
		if c.walk(p, p.origin(), fn) {
			return true
		}
	}
	return false
}

// surfaceAt returns the topmost surface under the screen point pt, or nil.
// The caller holds the lock.
func (c *Compositor) surfaceAt(pt image.Point) *surface {
	for _, w := range slices.Backward(c.windows) {
		// Note the last surface of the window, the topmost, under pt.
		var hit *surface
		c.walk(w, w.origin(), func(t *surface, at image.Point) bool {
			if pt.In(image.Rectangle{Min: at, Max: at.Add(t.size)}) {
				hit = t
			}
			return false
		})
		if hit != nil {
			return hit
		}
	}
	return nil
}

// readScreen composites the area r of the screen: the windows, bottom to
// top, over the background. The caller holds the lock.
func (c *Compositor) readScreen(r image.Rectangle) *s4wave_vm.WasiVmDisplayFrame {
	// Fill the background.
	pix := make([]byte, 4*r.Dx()*r.Dy())
	for i := 0; i < len(pix); i += 4 {
		copy(pix[i:i+4], background[:])
	}

	// Draw each surface of each window.
	for _, w := range c.windows {
		c.walk(w, w.origin(), func(s *surface, at image.Point) bool {
			blend(pix, r, s, at)
			return false
		})
	}
	return &s4wave_vm.WasiVmDisplayFrame{
		ScreenWidth:  uint32(c.screen.Dx()), //nolint:gosec // screen sizes are small.
		ScreenHeight: uint32(c.screen.Dy()), //nolint:gosec // screen sizes are small.
		X:            uint32(r.Min.X),       //nolint:gosec // inside the screen.
		Y:            uint32(r.Min.Y),       //nolint:gosec // inside the screen.
		Width:        uint32(r.Dx()),        //nolint:gosec // inside the screen.
		Height:       uint32(r.Dy()),        //nolint:gosec // inside the screen.
		Rgba:         pix,
	}
}

// blend draws the image of s, at screen position at, over pix, the opaque
// RGBA pixels of the screen area r. Each source pixel is premultiplied, so
// the result is the source plus the destination scaled by the source's
// transparency.
func blend(pix []byte, r image.Rectangle, s *surface, at image.Point) {
	// Find the area both cover.
	in := r.Intersect(image.Rectangle{Min: at, Max: at.Add(s.size)})
	for y := in.Min.Y; y < in.Max.Y; y++ {
		from := s.image[4*((y-at.Y)*s.size.X+in.Min.X-at.X):][:4*in.Dx()]
		to := pix[4*((y-r.Min.Y)*r.Dx()+in.Min.X-r.Min.X):][:4*in.Dx()]

		// Copy an opaque row, or blend each pixel.
		if s.opaque {
			copy(to, from)
			continue
		}
		for i := 0; i < len(from); i += 4 {
			keep := 255 - uint32(from[i+3])
			for j := range 3 {
				to[i+j] = from[i+j] + byte((uint32(to[i+j])*keep+127)/255)
			}
		}
	}
}

// followResize moves the window s so the edges opposite a resize grab's stay
// put as its size changes, and reports whether it moved. The caller holds
// the lock.
func (c *Compositor) followResize(s *surface) bool {
	// Find the corner the grab holds.
	g := c.grab
	if g == nil || g.w != s || g.edges == 0 {
		return false
	}
	t, size := s.xdg.toplevel, s.xdg.geometry().Size()
	pos := t.pos
	if g.edges&resizeLeft != 0 {
		pos.X = g.rect.Max.X - size.X
	}
	if g.edges&resizeTop != 0 {
		pos.Y = g.rect.Max.Y - size.Y
	}

	// Move the window there.
	if pos == t.pos {
		return false
	}
	t.pos = pos
	return true
}

// dismiss hides the popup s and the popups above it, telling the client
// when notify is set. The caller holds the lock.
func (c *Compositor) dismiss(s *surface, notify bool) {
	// Dismiss the popups on this one first.
	for _, p := range slices.Clone(s.popups) {
		c.dismiss(p, notify)
	}

	// Take it off its parent.
	c.damage(c.bounds(c.windowOf(s)))
	s.unparent()
	c.popupGrabs = slices.DeleteFunc(c.popupGrabs, func(p *surface) bool { return p == s })
	if notify && s.xdg != nil && s.xdg.popup != nil {
		s.c.send(newEvent(s.xdg.popup.id, popupDone))
	}
	c.refocusPointer()
}
