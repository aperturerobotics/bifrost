// Package wayland is a Wayland compositor for the clients in a WasiVM. It
// listens on a socket in the guest's filesystem, reads each client's shared
// memory buffers through the kernel, streams the screen to viewers outside
// the guest and turns their pointer and key input into Wayland events.
//
// Toplevels are kiosk style: each fills the screen, fullscreen and without
// decorations, and the newest mapped one is shown and holds the pointer and
// keyboard focus. Subsurfaces and popups are accepted but never drawn.
package wayland

import (
	"context"
	"image"
	"slices"
	"time"

	"github.com/aperturerobotics/util/broadcast"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// SocketPath is the compositor's socket in the guest. A client finds it with
// XDG_RUNTIME_DIR=/run and WAYLAND_DISPLAY=wayland-0.
const SocketPath = "/run/wayland-0"

// Compositor serves the Wayland clients of one WasiVM kernel.
type Compositor struct {
	// listener accepts the guest's connections.
	listener *kernel.UnixListener
	// screen is the output's size in pixels.
	screen image.Rectangle
	// start is the origin of event timestamps.
	start time.Time
	// keymap is the keyboard's xkbcommon keymap, sent to each keyboard.
	keymap *kernel.File

	// bcast guards the fields below and wakes the viewers when the shown
	// toplevel changes.
	bcast broadcast.Broadcast
	// clients is the connected clients.
	clients map[*client]struct{}
	// closed is set once Close runs, so no client connects after it.
	closed bool
	// serial numbers input and configure events.
	serial uint32
	// toplevels is the mapped toplevels, the shown one last.
	toplevels []*surface
	// focus is the surface holding the pointer and keyboard focus: the
	// shown toplevel, or nil.
	focus *surface
	// viewers holds each viewer's damage: the screen area changed since its
	// last frame.
	viewers map[*image.Rectangle]struct{}
	// pointer is the pointer's position on the screen.
	pointer image.Point
	// buttons is the mask of held pointer buttons, in viewer bit order.
	buttons uint32
	// keys is the held keys, as evdev key codes.
	keys []uint32
	// locked is the mask of locked modifiers: Caps Lock.
	locked uint32
}

// Listen starts a compositor with a screen of width by height on the
// kernel's socket at SocketPath. Serve accepts its clients.
func Listen(k *kernel.Kernel, width, height int) (*Compositor, error) {
	l, err := k.ListenUnix(SocketPath)
	if err != nil {
		return nil, errors.Wrap(err, "listen for wayland clients")
	}
	return &Compositor{
		listener: l,
		screen:   image.Rect(0, 0, width, height),
		start:    time.Now(),
		keymap:   k.NewMemfd("keymap", append([]byte(keymap), 0)),
		clients:  make(map[*client]struct{}),
		viewers:  make(map[*image.Rectangle]struct{}),
	}, nil
}

// Serve accepts and serves clients until Close or ctx ends.
func (c *Compositor) Serve(ctx context.Context) error {
	// End the accept loop with ctx.
	stop := context.AfterFunc(ctx, func() { _ = c.listener.Close() })
	defer stop()

	// Serve each connection.
	for {
		conn, err := c.listener.AcceptUnix()
		if err != nil {
			if ctx.Err() != nil {
				return context.Cause(ctx)
			}
			return err
		}
		cl := newClient(c, conn)
		var closed bool
		c.bcast.HoldLock(func(func(), func() <-chan struct{}) {
			closed = c.closed
			if !closed {
				c.clients[cl] = struct{}{}
			}
		})
		if closed {
			_ = conn.Close()
			return nil
		}
		go cl.serve()
	}
}

// Close stops listening and disconnects every client.
func (c *Compositor) Close() {
	// Stop accepting.
	_ = c.listener.Close()

	// Disconnect the clients, whose serve loops then clean up.
	var clients []*client
	c.bcast.HoldLock(func(func(), func() <-chan struct{}) {
		c.closed = true
		for cl := range c.clients {
			clients = append(clients, cl)
		}
	})
	for _, cl := range clients {
		_ = cl.conn.Close()
	}
	_ = c.keymap.Close()
}

// Run sends the whole screen to send, then each area that changes, until ctx
// ends. Each area is the bounding box of the changes since the last one.
func (c *Compositor) Run(ctx context.Context, send func(*s4wave_vm.WasiVmDisplayFrame) error) error {
	// Register the viewer with the whole screen damaged.
	damage := c.screen
	c.bcast.HoldLock(func(func(), func() <-chan struct{}) {
		c.viewers[&damage] = struct{}{}
	})
	defer c.bcast.HoldLock(func(func(), func() <-chan struct{}) {
		delete(c.viewers, &damage)
	})

	// Send the damaged area each time it grows.
	for {
		var frame *s4wave_vm.WasiVmDisplayFrame
		var waitCh <-chan struct{}
		c.bcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			if !damage.Empty() {
				frame = c.readScreen(damage)
				damage = image.Rectangle{}
			}
			waitCh = getWaitCh()
		})
		if frame != nil {
			if err := send(frame); err != nil {
				return err
			}
		}
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-waitCh:
		}
	}
}

// readScreen copies the area r of the screen: the shown toplevel's pixels,
// black past its edges. The caller holds the lock.
func (c *Compositor) readScreen(r image.Rectangle) *s4wave_vm.WasiVmDisplayFrame {
	// Start from black.
	pix := make([]byte, 4*r.Dx()*r.Dy())
	for i := 3; i < len(pix); i += 4 {
		pix[i] = 0xff
	}

	// Copy the rows the toplevel covers.
	if len(c.toplevels) != 0 {
		s := c.toplevels[len(c.toplevels)-1]
		in := r.Intersect(image.Rectangle{Max: s.size})
		for y := in.Min.Y; y < in.Max.Y; y++ {
			from := 4 * (y*s.size.X + in.Min.X)
			to := 4 * ((y-r.Min.Y)*r.Dx() + in.Min.X - r.Min.X)
			copy(pix[to:], s.image[from:from+4*in.Dx()])
		}
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

// damage adds r to every viewer's damage and wakes the viewers. The caller
// holds the lock.
func (c *Compositor) damage(r image.Rectangle, broadcast func()) {
	r = r.Intersect(c.screen)
	if r.Empty() {
		return
	}
	for d := range c.viewers {
		*d = d.Union(r)
	}
	broadcast()
}

// show maps s on top when it is not mapped, or unmaps it, then moves the
// focus to the shown toplevel and damages the screen when it changed. The
// caller holds the lock.
func (c *Compositor) show(s *surface, mapped bool, broadcast func()) {
	// Update the stack.
	top := c.shown()
	c.toplevels = slices.DeleteFunc(c.toplevels, func(t *surface) bool { return t == s })
	if mapped {
		c.toplevels = append(c.toplevels, s)
	}
	if c.shown() == top {
		return
	}

	// Redraw the screen and move the focus.
	c.damage(c.screen, broadcast)
	c.setFocus(c.shown())
}

// shown returns the shown toplevel, or nil. The caller holds the lock.
func (c *Compositor) shown() *surface {
	if len(c.toplevels) == 0 {
		return nil
	}
	return c.toplevels[len(c.toplevels)-1]
}

// setFocus moves the pointer and keyboard focus to s, which may be nil,
// telling the client losing it and the client gaining it. The caller holds
// the lock.
func (c *Compositor) setFocus(s *surface) {
	// Leave the old surface.
	if old := c.focus; old != nil {
		serial := c.nextSerial()
		for _, p := range old.c.pointers {
			old.c.send(newEvent(p.id, pointerLeave).uint(serial).uint(old.id))
			old.c.pointerFrame(p)
		}
		for _, k := range old.c.keyboards {
			old.c.send(newEvent(k.id, keyboardLeave).uint(serial).uint(old.id))
		}
	}

	// Enter the new one.
	c.focus = s
	if s != nil {
		for _, p := range s.c.pointers {
			c.enterPointer(s, p)
		}
		for _, k := range s.c.keyboards {
			c.enterKeyboard(s, k)
		}
	}
}

// enterPointer sends the pointer p of the focused surface s its enter event.
// The caller holds the lock.
func (c *Compositor) enterPointer(s *surface, p resource) {
	s.c.send(newEvent(p.id, pointerEnter).uint(c.nextSerial()).uint(s.id).
		fixed(int32(c.pointer.X)).fixed(int32(c.pointer.Y))) //nolint:gosec // inside the screen.
	s.c.pointerFrame(p)
}

// enterKeyboard sends the keyboard k of the focused surface s its enter
// event with the held keys and the modifiers. The caller holds the lock.
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

// nextSerial returns a new event serial. The caller holds the lock.
func (c *Compositor) nextSerial() uint32 {
	c.serial++
	return c.serial
}

// now returns the event timestamp: milliseconds since the start.
func (c *Compositor) now() uint32 {
	return uint32(time.Since(c.start).Milliseconds()) //nolint:gosec // wraps as Wayland timestamps do.
}

// Input sends a viewer's pointer or key event to the focused surface. A
// keysym the keyboard lacks is dropped.
func (c *Compositor) Input(in *s4wave_vm.WasiVmDisplayInput) {
	c.bcast.HoldLock(func(func(), func() <-chan struct{}) {
		switch {
		case in.GetPointer() != nil:
			c.pointerInput(in.GetPointer())
		case in.GetKey() != nil:
			c.keyInput(in.GetKey())
		}
	})
}

// pointerInput moves the pointer and presses, releases or scrolls with the
// buttons whose state changed. The caller holds the lock.
func (c *Compositor) pointerInput(ev *s4wave_vm.WasiVmPointerEvent) {
	// Move the pointer within the screen.
	c.pointer = image.Pt(
		max(min(int(ev.GetX()), c.screen.Max.X-1), 0),
		max(min(int(ev.GetY()), c.screen.Max.Y-1), 0),
	)
	changed := ev.GetButtons() ^ c.buttons
	c.buttons = ev.GetButtons()
	s := c.focus
	if s == nil {
		return
	}

	// Send each pointer the motion and the button changes as one frame.
	time := c.now()
	for _, p := range s.c.pointers {
		s.c.send(newEvent(p.id, pointerMotion).uint(time).
			fixed(int32(c.pointer.X)).fixed(int32(c.pointer.Y))) //nolint:gosec // inside the screen.
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

	// Send the key and the new modifiers to the focused surface.
	s := c.focus
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

// remove forgets a disconnected client and its toplevels. The caller holds
// the lock.
func (c *Compositor) remove(cl *client, broadcast func()) {
	delete(c.clients, cl)
	if c.focus != nil && c.focus.c == cl {
		c.focus = nil
	}
	for _, s := range slices.Clone(c.toplevels) {
		if s.c == cl {
			c.show(s, false, broadcast)
		}
	}
}
