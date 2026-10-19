package wayland

import (
	"image"

	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// The desktop's window frame, in screen units: the title bar above a
// window's geometry, and the band around the frame that resizes it.
const (
	titleHeight = 28
	frameBorder = 4
)

// The title bar buttons.
const (
	buttonClose    = s4wave_vm.WasiVmSceneFrameButtonKind_WASI_VM_SCENE_FRAME_BUTTON_KIND_CLOSE
	buttonMaximize = s4wave_vm.WasiVmSceneFrameButtonKind_WASI_VM_SCENE_FRAME_BUTTON_KIND_MAXIMIZE
)

// frameButton is a title bar button and its screen rectangle.
type frameButton struct {
	kind s4wave_vm.WasiVmSceneFrameButtonKind
	r    image.Rectangle
}

// frameHit is a part of a window's frame under the pointer: a button, the
// resize band's edges, or else the title bar. The zero value is no frame.
type frameHit struct {
	// w is the window, or nil.
	w *surface
	// button is the button, or unknown.
	button s4wave_vm.WasiVmSceneFrameButtonKind
	// edges is the mask of the edges the band resizes, or zero.
	edges uint32
}

// bindDecorationManager binds zxdg_decoration_manager_v1, which lets a
// window's client ask the compositor to draw its decorations.
func (c *client) bindDecorationManager(r resource) object {
	return requestFunc(func(op uint16, m *message) error {
		switch op {
		case decorationManagerDestroy:
			c.destroy(r.id)
		case decorationManagerGetToplevelDecoration:
			// Attach the decoration to the toplevel.
			id, tid := m.uint(), m.uint()
			if m.err != nil {
				return nil
			}
			t, ok := c.objects[tid].(*toplevel)
			switch {
			case !ok:
				return c.fail(r.id, errInvalidObject, "bad toplevel")
			case t.decoration != 0:
				return c.fail(r.id, decorationErrAlreadyConstructed, "toplevel has a decoration")
			}
			if err := c.add(id, requestFunc(t.decorationRequest)); err != nil {
				return err
			}
			c.comp.update(func() { t.decoration = id })
		}
		return nil
	})
}

// decorationRequest handles a zxdg_toplevel_decoration_v1 request. A mode
// change configures the window again, as the protocol requires.
func (t *toplevel) decorationRequest(op uint16, m *message) error {
	c := t.x.s.c
	switch op {
	case decorationDestroy:
		c.destroy(t.decoration)
		c.comp.update(func() {
			t.decoration = 0
			c.comp.setMode(t, 0)
		})
	case decorationSetMode:
		mode := m.uint()
		if m.err == nil {
			c.comp.update(func() { c.comp.setMode(t, mode) })
		}
	case decorationUnsetMode:
		c.comp.update(func() { c.comp.setMode(t, 0) })
	}
	return nil
}

// decorationMode returns the decoration mode the window uses: the one its
// client asked for, else server side. The caller holds the lock.
func (t *toplevel) decorationMode() uint32 {
	if t.mode == decorationClientSide {
		return decorationClientSide
	}
	return decorationServerSide
}

// titleHeight returns the height of the title bar the desktop draws above
// the window: none unless the window has server side decorations and shows
// on the desktop short of the full screen. The caller holds the lock.
func (t *toplevel) titleHeight() int {
	if t.decoration == 0 || t.decorationMode() != decorationServerSide || t.own != nil || t.fullscreen {
		return 0
	}
	return titleHeight
}

// setMode sets the decoration mode the client of t asked for and
// configures the window with it, fitting a maximized window to its new
// title bar. The caller holds the lock.
func (c *Compositor) setMode(t *toplevel, mode uint32) {
	t.mode = mode
	c.changed = true
	switch x := t.x; {
	case !x.configured:
	case t.maximized:
		c.setState(x.s, t.maximized, t.fullscreen)
	default:
		x.configure()
	}
}

// frameOf returns the screen rectangle of the window w with its title bar,
// and whether the desktop draws a frame for it. The caller holds the lock.
func frameOf(w *surface) (image.Rectangle, bool) {
	// Check the window has a frame, then extend its geometry up by the
	// title bar.
	t := w.xdg.toplevel
	th := t.titleHeight()
	if th == 0 {
		return image.Rectangle{}, false
	}
	r := image.Rectangle{Min: t.pos, Max: t.pos.Add(w.xdg.geometry().Size())}
	r.Min.Y -= th
	return r, true
}

// frameButtons returns the buttons of the title bar of the frame f, from
// the right: close, then maximize, each a square as tall as the title bar.
func frameButtons(f image.Rectangle) []frameButton {
	closeR := image.Rect(f.Max.X-titleHeight, f.Min.Y, f.Max.X, f.Min.Y+titleHeight)
	return []frameButton{
		{buttonClose, closeR},
		{buttonMaximize, closeR.Sub(image.Pt(titleHeight, 0))},
	}
}

// frameAt returns the part of the frame of the window w at the screen
// point pt, or no frame. The band resizes only a window that can resize.
// The caller holds the lock.
func frameAt(w *surface, pt image.Point) frameHit {
	// Check the point is on the frame or its band, outside the content.
	f, ok := frameOf(w)
	t := w.xdg.toplevel
	border := frameBorder
	if t.maximized {
		border = 0
	}
	content := image.Rectangle{Min: t.pos, Max: f.Max}
	if !ok || !pt.In(f.Inset(-border)) || pt.In(content) {
		return frameHit{}
	}

	// Find the band's edges, then a button on the title bar.
	hit := frameHit{w: w}
	switch {
	case pt.X < f.Min.X:
		hit.edges |= resizeLeft
	case pt.X >= f.Max.X:
		hit.edges |= resizeRight
	}
	switch {
	case pt.Y < f.Min.Y:
		hit.edges |= resizeTop
	case pt.Y >= f.Max.Y:
		hit.edges |= resizeBottom
	}
	if hit.edges != 0 {
		return hit
	}
	for _, b := range frameButtons(f) {
		if pt.In(b.r) {
			hit.button = b.kind
		}
	}
	return hit
}

// pressFrame starts what a press on the frame part f does: a press on the
// title bar moves the window and one on the band resizes it at once, and a
// press on a button waits for its release. The caller holds the lock.
func (c *Compositor) pressFrame(f frameHit) {
	if f.button != 0 {
		c.framePress = f
		return
	}
	c.beginGrab(f.w, f.edges)
}

// releaseFrame acts on a button pressed on a frame when the pointer, on the
// space sp, is released over the same button. The caller holds the lock.
func (c *Compositor) releaseFrame(sp *space) {
	// Take the pressed button and check the pointer is still on it.
	f := c.framePress
	c.framePress = frameHit{}
	if f.w == nil {
		return
	}
	if _, at := c.hitAt(sp, c.pointer); at != f {
		return
	}

	// Ask the window to close, or maximize or restore it.
	t := f.w.xdg.toplevel
	switch f.button {
	case buttonClose:
		f.w.c.send(newEvent(t.id, toplevelClose))
	case buttonMaximize:
		c.setState(f.w, !t.maximized, t.fullscreen)
	}
}

// sceneFrame returns the frame of the window w as the space sp shows it, or
// nil when the desktop draws none. The caller holds the lock.
func (c *Compositor) sceneFrame(sp *space, w *surface) *s4wave_vm.WasiVmSceneFrame {
	// Find the frame, relative to the root.
	f, ok := frameOf(w)
	if !ok {
		return nil
	}
	t, origin := w.xdg.toplevel, w.origin()
	border := uint32(frameBorder)
	if t.maximized {
		border = 0
	}

	// Describe it with its buttons.
	rel := f.Sub(origin)
	frame := &s4wave_vm.WasiVmSceneFrame{
		Root:        w.sceneID,
		Title:       t.title,
		Active:      c.active == w,
		Maximized:   t.maximized,
		X:           int32(rel.Min.X), //nolint:gosec // screen positions are small.
		Y:           int32(rel.Min.Y), //nolint:gosec // screen positions are small.
		Width:       uint32(rel.Dx()), //nolint:gosec // window sizes are small.
		Height:      uint32(rel.Dy()), //nolint:gosec // window sizes are small.
		TitleHeight: titleHeight,
		Border:      border,
		MinY:        int32(c.minTop(w) + origin.Y - t.pos.Y - sp.area.Min.Y), //nolint:gosec // screen positions are small.
	}
	for _, b := range frameButtons(f) {
		r := b.r.Sub(origin)
		frame.Buttons = append(frame.Buttons, &s4wave_vm.WasiVmSceneFrameButton{
			Kind:   b.kind,
			X:      int32(r.Min.X), //nolint:gosec // screen positions are small.
			Y:      int32(r.Min.Y), //nolint:gosec // screen positions are small.
			Width:  uint32(r.Dx()), //nolint:gosec // button sizes are small.
			Height: uint32(r.Dy()), //nolint:gosec // button sizes are small.
		})
	}
	return frame
}
