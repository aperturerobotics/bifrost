package wayland

import (
	"image"
	"slices"

	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// grab is an interactive move or resize of a window: it follows the pointer
// until the buttons release.
type grab struct {
	// w is the window.
	w *surface
	// edges is the mask of the edges a resize moves, or zero for a move.
	edges uint32
	// pointer is the pointer's position when the grab started.
	pointer image.Point
	// rect is the window geometry's screen rectangle when the grab started.
	rect image.Rectangle
}

// Input sends a viewer's pointer or key event to the surfaces holding the
// focus. A keysym the keyboard lacks is dropped.
func (c *Compositor) Input(in *s4wave_vm.WasiVmDisplayInput) {
	c.update(func() {
		switch {
		case in.GetPointer() != nil:
			c.pointerInput(in.GetPointer())
		case in.GetKey() != nil:
			c.keyInput(in.GetKey())
		}
	})
}

// activate moves the keyboard focus to the window w, which may be nil, and
// configures the windows losing and gaining it. The caller holds the lock.
func (c *Compositor) activate(w *surface) {
	// Leave the old window.
	old := c.active
	if old == w {
		return
	}
	c.active = w
	if old != nil {
		serial := c.nextSerial()
		for _, k := range old.c.keyboards {
			old.c.send(newEvent(k.id, keyboardLeave).uint(serial).uint(old.id))
		}
		if old.xdg != nil && old.xdg.configured {
			old.xdg.configure()
		}
	}

	// Enter the new one.
	if w != nil {
		for _, k := range w.c.keyboards {
			c.enterKeyboard(w, k)
		}
		w.xdg.configure()
	}
}

// refocusPointer moves the pointer focus to the surface under the pointer.
// While a button is held, the focused surface keeps the pointer until it
// leaves the screen, and a grab keeps it from every surface. The caller
// holds the lock.
func (c *Compositor) refocusPointer() {
	switch {
	case c.grab != nil:
	case c.buttons&buttonBits == 0:
		c.setHover(c.surfaceAt(c.pointer))
	case c.hover != nil && (c.windowOf(c.hover) == nil || c.hover.image == nil):
		c.setHover(nil)
	}
}

// setHover moves the pointer focus to s, which may be nil, telling the
// client losing it and the client gaining it. The caller holds the lock.
func (c *Compositor) setHover(s *surface) {
	// Leave the old surface.
	old := c.hover
	if old == s {
		return
	}
	c.hover = s
	if old != nil {
		serial := c.nextSerial()
		for _, p := range old.c.pointers {
			old.c.send(newEvent(p.id, pointerLeave).uint(serial).uint(old.id))
			old.c.pointerFrame(p)
		}
	}

	// Enter the new one.
	if s != nil {
		for _, p := range s.c.pointers {
			c.enterPointer(s, p)
		}
	}
}

// forget drops the destroyed surface s from the focus, the grab and the
// popup grabs without telling its client. The caller holds the lock.
func (c *Compositor) forget(s *surface) {
	if c.hover == s {
		c.hover = nil
	}
	if c.active == s {
		c.active = nil
	}
	if c.grab != nil && c.grab.w == s {
		c.grab = nil
	}
	c.popupGrabs = slices.DeleteFunc(c.popupGrabs, func(p *surface) bool { return p == s })
}

// enterPointer sends the pointer p of the hovered surface s its enter event
// at the pointer's position on s. The caller holds the lock.
func (c *Compositor) enterPointer(s *surface, p resource) {
	at := c.pointer.Sub(s.origin())
	s.c.send(newEvent(p.id, pointerEnter).uint(c.nextSerial()).uint(s.id).
		fixed(int32(at.X)).fixed(int32(at.Y))) //nolint:gosec // near the screen.
	s.c.pointerFrame(p)
}

// enterKeyboard sends the keyboard k of the active window s its enter event
// with the held keys and the modifiers. The caller holds the lock.
func (c *Compositor) enterKeyboard(s *surface, k resource) {
	s.c.send(newEvent(k.id, keyboardEnter).uint(c.nextSerial()).uint(s.id).uints(c.keys...))
	s.c.send(c.modifiersEvent(k.id))
}

// modifiersEvent returns the modifiers event for keyboard id. The caller
// holds the lock.
func (c *Compositor) modifiersEvent(id uint32) *event {
	return newEvent(id, keyboardModifiers).uint(c.nextSerial()).
		uint(heldModifiers(c.keys)).uint(0).uint(c.locked).uint(0)
}

// pointerInput moves the pointer and presses, releases or scrolls with the
// buttons whose state changed. A grab takes the pointer from the surfaces.
// The caller holds the lock.
func (c *Compositor) pointerInput(ev *s4wave_vm.WasiVmPointerEvent) {
	// Move the pointer within the screen and note the buttons that changed.
	c.pointer = image.Pt(
		max(min(int(ev.GetX()), c.screen.Max.X-1), c.screen.Min.X),
		max(min(int(ev.GetY()), c.screen.Max.Y-1), c.screen.Min.Y),
	)
	changed := ev.GetButtons() ^ c.buttons
	held := c.buttons&buttonBits != 0
	c.buttons = ev.GetButtons()

	// Move or resize a grabbed window until the buttons release.
	if c.grab != nil {
		c.drag()
		if c.buttons&buttonBits == 0 {
			c.endGrab()
		}
		return
	}

	// Focus the surface under the pointer. A first press dismisses the
	// popups outside it, then raises and activates its window.
	if !held {
		c.setHover(c.surfaceAt(c.pointer))
		if c.buttons&changed&buttonBits != 0 {
			c.press()
		}
	}

	// Send the events, then follow the pointer once the buttons release.
	c.sendPointer(changed)
	if c.buttons&buttonBits == 0 {
		c.refocusPointer()
	}
}

// press dismisses the grabbing popups the hovered surface is outside of, then
// raises and activates its window. The caller holds the lock.
func (c *Compositor) press() {
	for n := len(c.popupGrabs); n != 0; n-- {
		top := c.popupGrabs[n-1]
		if c.hover != nil && c.hover.ancestor(top) {
			break
		}
		c.dismiss(top, true)
	}
	if w := c.windowOf(c.hover); w != nil {
		c.raise(w)
		c.activate(w)
	}
}

// sendPointer sends the hovered surface's pointers the motion and the
// changed buttons as one frame. The caller holds the lock.
func (c *Compositor) sendPointer(changed uint32) {
	// Find the pointer's position on the surface.
	s := c.hover
	if s == nil {
		return
	}
	at := c.pointer.Sub(s.origin())

	// Send each pointer its frame.
	time := c.now()
	for _, p := range s.c.pointers {
		s.c.send(newEvent(p.id, pointerMotion).uint(time).
			fixed(int32(at.X)).fixed(int32(at.Y))) //nolint:gosec // near the screen.
		for bit, code := range pointerButtonCodes {
			if changed&(1<<bit) == 0 {
				continue
			}
			pressed := c.buttons&(1<<bit) != 0
			switch {
			case code != 0:
				state := uint32(0)
				if pressed {
					state = 1
				}
				s.c.send(newEvent(p.id, pointerButton).uint(c.nextSerial()).uint(time).uint(code).uint(state))
			case pressed:
				s.c.send(newEvent(p.id, pointerAxis).uint(time).uint(0).fixed(scrollSteps[bit]))
			}
		}
		s.c.pointerFrame(p)
	}
}

// startGrab starts moving the window s, or resizing it by edges, when a
// button pressed on it is still held. A maximized or fullscreen window
// stays put. The caller holds the lock.
func (c *Compositor) startGrab(s *surface, edges uint32) {
	// Check the press and the window.
	t := s.xdg.toplevel
	if c.grab != nil || c.buttons&buttonBits == 0 || c.windowOf(c.hover) != s || t.maximized || t.fullscreen {
		return
	}

	// Take the pointer from the surfaces and tell a resized window.
	c.grab = &grab{
		w:       s,
		edges:   edges,
		pointer: c.pointer,
		rect:    image.Rectangle{Min: t.pos, Max: t.pos.Add(s.xdg.geometry().Size())},
	}
	c.setHover(nil)
	if edges != 0 {
		s.xdg.configure()
	}
}

// drag moves the grabbed window with the pointer, or configures it at the
// size the pointer drags it to. The caller holds the lock.
func (c *Compositor) drag() {
	// Move the window, keeping its top on the screen.
	g := c.grab
	t, d := g.w.xdg.toplevel, c.pointer.Sub(g.pointer)
	if g.edges == 0 {
		before := c.bounds(g.w)
		t.pos = g.rect.Min.Add(d)
		t.pos.Y = max(t.pos.Y, c.screen.Min.Y)
		c.damage(before.Union(c.bounds(g.w)))
		return
	}

	// Drag the grabbed edges, within the window's size bounds.
	size := g.rect.Size()
	if g.edges&resizeLeft != 0 {
		size.X -= d.X
	} else if g.edges&resizeRight != 0 {
		size.X += d.X
	}
	if g.edges&resizeTop != 0 {
		size.Y -= d.Y
	} else if g.edges&resizeBottom != 0 {
		size.Y += d.Y
	}
	size = image.Pt(clampSize(size.X, t.minSize.X, t.maxSize.X), clampSize(size.Y, t.minSize.Y, t.maxSize.Y))

	// Configure the window at the new size.
	if size != t.size {
		t.size = size
		g.w.xdg.configure()
	}
}

// clampSize bounds a window side to at least lo and one pixel, and at most
// hi when hi is set.
func clampSize(n, lo, hi int) int {
	n = max(n, lo, 1)
	if hi > 0 {
		n = min(n, hi)
	}
	return n
}

// endGrab ends the grab, tells a resized window it no longer resizes and
// gives the pointer back to the surface under it. The caller holds the lock.
func (c *Compositor) endGrab() {
	g := c.grab
	c.grab = nil
	if g.edges != 0 {
		g.w.xdg.configure()
	}
	c.refocusPointer()
}

// keyInput presses or releases the key that produces a keysym and sends the
// modifiers when they change. The caller holds the lock.
func (c *Compositor) keyInput(ev *s4wave_vm.WasiVmKeyEvent) {
	// Find the key and update the held keys and the locks.
	code, ok := keysymCodes[ev.GetKeysym()]
	if !ok {
		return
	}
	mods, locked := heldModifiers(c.keys), c.locked
	held := slices.Contains(c.keys, code)
	switch {
	case ev.GetDown() && !held:
		c.keys = append(c.keys, code)
		if code == keyCapsLock {
			c.locked ^= modCapsLock
		}
	case !ev.GetDown() && held:
		c.keys = slices.DeleteFunc(c.keys, func(k uint32) bool { return k == code })
	default:
		return
	}

	// Send the key and the new modifiers to the active window.
	s := c.active
	if s == nil {
		return
	}
	state := uint32(0)
	if ev.GetDown() {
		state = 1
	}
	for _, k := range s.c.keyboards {
		s.c.send(newEvent(k.id, keyboardKey).uint(c.nextSerial()).uint(c.now()).uint(code).uint(state))
		if heldModifiers(c.keys) != mods || c.locked != locked {
			s.c.send(c.modifiersEvent(k.id))
		}
	}
}
