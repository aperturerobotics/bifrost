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

// mapWindow shows the toplevel s on top of the stack, numbering and placing
// it when it is new, and activates it. The caller holds the lock.
func (c *Compositor) mapWindow(s *surface) {
	// Number and place a new window on the cascade.
	x, t := s.xdg, s.xdg.toplevel
	if !t.placed {
		t.placed = true
		t.pos = c.usable.Min.Add(image.Pt(cascadeStep, cascadeStep).Mul(1 + c.placed%cascadeSlots))
		c.placed++
		c.lastWindow++
		t.windowID = c.lastWindow
	}

	// Show and activate it.
	x.mapped = true
	c.windows = append(c.windows, s)
	c.changed = true
	c.damage(c.desktop, c.bounds(s))
	c.activate(s)
	c.refocusPointer()
}

// mapLayer shows the layer surface s on top of its layer and rearranges the
// layers around it. The caller holds the lock.
func (c *Compositor) mapLayer(s *surface) {
	// Show it, then place the layers and move the focus around it.
	l := s.layer
	l.mapped = true
	c.layers[l.layer] = append(c.layers[l.layer], s)
	c.arrange()
	c.damage(c.desktop, c.bounds(s))
	c.refocusKeyboard()
	c.refocusPointer()
}

// unmap takes the window or layer surface s off the screen, dismisses its
// popups and moves the focus it held. A window in its own space ends it and
// maps again on the desktop. The client must configure it again before it
// maps. The caller holds the lock.
func (c *Compositor) unmap(s *surface) {
	// Hide it and its popups.
	if !s.mapped() {
		return
	}
	c.redrawTree(s)
	for _, p := range slices.Clone(s.popups) {
		c.dismiss(p, true)
	}

	// Take a window off the stack or out of its space, or a layer surface
	// off its layer, making room for the windows.
	if x := s.xdg; x != nil {
		x.mapped, x.configured = false, false
		if x.toplevel.own != nil {
			c.leaveSpace(s)
		}
		c.windows = slices.DeleteFunc(c.windows, func(w *surface) bool { return w == s })
		c.changed = true
	} else {
		l := s.layer
		l.mapped, l.configured = false, false
		c.layers[l.layer] = slices.DeleteFunc(c.layers[l.layer], func(t *surface) bool { return t == s })
		c.arrange()
	}

	// Move the focus.
	if c.grab != nil && c.grab.w == s {
		c.grab = nil
	}
	if c.layerFocus == s {
		c.layerFocus = nil
	}
	if c.active == s {
		c.activate(nil)
		if n := len(c.windows); n != 0 {
			c.activate(c.windows[n-1])
		}
	}
	c.refocusKeyboard()
	c.refocusPointer()
}

// arrange places the mapped layer surfaces and gives the windows the area
// the exclusive zones leave. The surfaces reserving an edge go first, from
// the top layer down, each taking its edge from the area the ones before
// left. Maximized windows then fill the new usable area. The caller holds
// the lock.
func (c *Compositor) arrange() {
	// Place the surfaces reserving an edge.
	usable := c.desktop.area
	for _, layer := range slices.Backward(c.layers[:]) {
		for _, s := range layer {
			if s.layer.exclusiveEdge() != 0 {
				s.layer.place(usable)
				usable = s.layer.reserve(usable)
			}
		}
	}

	// Place the rest in the usable area, or on the whole screen for a
	// negative zone.
	for _, layer := range c.layers {
		for _, s := range layer {
			switch l := s.layer; {
			case l.exclusiveEdge() != 0:
			case l.zone < 0:
				l.place(c.desktop.area)
			default:
				l.place(usable)
			}
		}
	}

	// Fit the maximized windows to a changed usable area.
	if usable == c.usable {
		return
	}
	c.usable = usable
	for _, w := range c.windows {
		if t := w.xdg.toplevel; t.maximized && !t.fullscreen {
			before := c.bounds(w)
			t.pos, t.size = usable.Min, usable.Size()
			c.damage(c.desktop, before.Union(c.bounds(w)))
			w.xdg.configure()
		}
	}
}

// raise moves the mapped window w to the top of the stack, drawing again
// the parts of it the windows above covered. The caller holds the lock.
func (c *Compositor) raise(w *surface) {
	i := slices.Index(c.windows, w)
	b := c.bounds(w)
	for _, above := range c.windows[i+1:] {
		c.damage(c.desktop, b.Intersect(c.bounds(above)))
	}
	c.windows = append(slices.Delete(c.windows, i, i+1), w)
}

// moveWindow moves the window w so its geometry's corner sits at pos. When
// its surface is opaque, the viewers move the pixels they have and draw only
// what the move uncovers and the trees above it. The caller holds the lock.
func (c *Compositor) moveWindow(w *surface, pos image.Point) {
	// Move it.
	before, from := c.bounds(w), w.origin()
	w.xdg.toplevel.pos = pos
	after, by := c.bounds(w), w.origin().Sub(from)
	if by == (image.Point{}) {
		return
	}

	// Move the pixels of an opaque surface. The copy carries along the trees
	// drawn above its source and covers those above its destination, so draw
	// both again.
	var moved image.Rectangle
	if w.opaque && w.image != nil {
		moved = c.shift(image.Rectangle{Min: from, Max: from.Add(w.size)}, by)
		roots := c.roots()
		for _, above := range roots[slices.Index(roots, w)+1:] {
			b := c.bounds(above)
			c.damage(c.desktop, moved.Intersect(b))
			c.damage(c.desktop, moved.Intersect(b.Add(by)))
		}
	}

	// Draw the rest of both places.
	for _, r := range slices.Concat(subtract(before, moved), subtract(after, moved)) {
		c.damage(c.desktop, r)
	}
}

// setState maximizes the window s to the usable area, fills the screen with
// it, or restores it, and configures it at the new size. A window in its own
// space keeps filling it. The caller holds the lock.
func (c *Compositor) setState(s *surface, maximized, fullscreen bool) {
	// Answer a window in its own space with its current state.
	t := s.xdg.toplevel
	if t.own != nil {
		if s.xdg.configured {
			s.xdg.configure()
		}
		return
	}

	// Remember the window's place before it first leaves it.
	wasNormal := !t.maximized && !t.fullscreen
	if wasNormal && (maximized || fullscreen) {
		t.restore = image.Rectangle{Min: t.pos, Max: t.pos.Add(s.xdg.geometry().Size())}
	}
	t.maximized, t.fullscreen = maximized, fullscreen

	// Fill the usable area or the screen, or go back to the remembered
	// place. The client's last image shows at the new place until it draws
	// at the new size, so draw both places again.
	before := c.bounds(s)
	switch {
	case fullscreen:
		t.pos, t.size = c.desktop.area.Min, c.desktop.area.Size()
	case maximized:
		t.pos, t.size = c.usable.Min, c.usable.Size()
	case !wasNormal:
		t.pos, t.size = t.restore.Min, t.restore.Size()
	}
	c.damage(c.desktop, before.Union(c.bounds(s)))
	if s.xdg.configured {
		s.xdg.configure()
	}
}

// rootOf returns the mapped window or layer surface whose tree holds s, or
// nil. The caller holds the lock.
func (c *Compositor) rootOf(s *surface) *surface {
	// Find the root of the tree and check it shows.
	if s == nil {
		return nil
	}
	root := s
	for root.parent != nil {
		root = root.parent
	}
	if !root.mapped() {
		return nil
	}
	return root
}

// roots returns the mapped windows and layer surfaces bottom to top: the
// background and bottom layers, the windows, then the top and overlay
// layers. The caller holds the lock.
func (c *Compositor) roots() []*surface {
	return slices.Concat(c.layers[layerBackground], c.layers[layerBottom], c.windows, c.layers[layerTop], c.layers[layerOverlay])
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

// surfaceAt returns the topmost surface under the point pt of the space sp,
// or nil. The caller holds the lock.
func (c *Compositor) surfaceAt(sp *space, pt image.Point) *surface {
	for _, w := range slices.Backward(c.spaceRoots(sp)) {
		// Note the last surface of the tree, the topmost, under pt.
		var hit *surface
		c.walk(w, w.origin(), func(t *surface, at image.Point) bool {
			if t.takesInput(pt.Sub(at)) {
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

// readFrame takes what the view v of the space sp lacks as a frame: its
// moves, then its areas read from the space. The caller holds the lock.
func (c *Compositor) readFrame(sp *space, v *view) *s4wave_vm.WasiVmDisplayFrame {
	// List the moves, then read each area after them.
	f := &s4wave_vm.WasiVmDisplayFrame{
		ScreenWidth:  uint32(sp.area.Dx()), //nolint:gosec // space sizes are small.
		ScreenHeight: uint32(sp.area.Dy()), //nolint:gosec // space sizes are small.
	}
	for _, m := range v.copies {
		src, dst := m.src, m.dst()
		f.Copies = append(f.Copies, &s4wave_vm.WasiVmDisplayCopy{
			X:      uint32(src.Min.X), //nolint:gosec // inside the screen.
			Y:      uint32(src.Min.Y), //nolint:gosec // inside the screen.
			Width:  uint32(src.Dx()),  //nolint:gosec // inside the screen.
			Height: uint32(src.Dy()),  //nolint:gosec // inside the screen.
			ToX:    uint32(dst.Min.X), //nolint:gosec // inside the screen.
			ToY:    uint32(dst.Min.Y), //nolint:gosec // inside the screen.
		})
	}
	for _, r := range v.damage {
		f.Regions = append(f.Regions, c.readSpace(sp, r))
	}
	v.copies, v.damage = v.copies[:0], v.damage[:0]
	return f
}

// readSpace composites the area r of the space sp: its windows and layer
// surfaces, bottom to top, over the background. The caller holds the lock.
func (c *Compositor) readSpace(sp *space, r image.Rectangle) *s4wave_vm.WasiVmDisplayRegion {
	// Fill the background: one pixel, then double the filled prefix, so a
	// large area takes a few long copies instead of one per pixel.
	pix := make([]byte, 4*r.Dx()*r.Dy())
	for n := copy(pix, background[:]); n < len(pix); {
		n += copy(pix[n:], pix[:n])
	}

	// Draw each surface of each tree.
	for _, w := range c.spaceRoots(sp) {
		c.walk(w, w.origin(), func(s *surface, at image.Point) bool {
			blend(pix, r, s, at)
			return false
		})
	}
	return &s4wave_vm.WasiVmDisplayRegion{
		X:      uint32(r.Min.X), //nolint:gosec // inside the screen.
		Y:      uint32(r.Min.Y), //nolint:gosec // inside the screen.
		Width:  uint32(r.Dx()),  //nolint:gosec // inside the screen.
		Height: uint32(r.Dy()),  //nolint:gosec // inside the screen.
		Rgba:   pix,
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
				to[i+j] = from[i+j] + byte((uint32(to[i+j])*keep+127)/255) //nolint:gosec // at most 255.
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
	c.redrawTree(s)
	s.unparent()
	c.popupGrabs = slices.DeleteFunc(c.popupGrabs, func(p *surface) bool { return p == s })
	if notify && s.xdg != nil && s.xdg.popup != nil {
		s.c.send(newEvent(s.xdg.popup.id, popupDone))
	}
	c.refocusPointer()
}
