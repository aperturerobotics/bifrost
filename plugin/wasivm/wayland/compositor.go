// Package wayland is a Wayland compositor for the clients in a WasiVM. It
// listens on a socket in the guest's filesystem, reads each client's shared
// memory buffers through the kernel, streams the screen to viewers outside
// the guest and turns their pointer and key input into Wayland events.
//
// The compositor is a stacking window manager. Windows open on a cascade at
// the size their clients choose, with the decorations the clients draw, and
// show over a plain background with their subsurfaces and popups. A press
// raises and activates the window under the pointer, which then holds the
// keyboard, and a client's move or resize request drags its window with the
// pointer. Layer shell clients draw the rest of the desktop, such as a
// wallpaper, a panel or a launcher, below or above the windows, and an
// exclusive zone keeps the windows off a panel's edge.
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

	// bcast guards the fields below and the surfaces' screen state, and
	// wakes the viewers when the screen changes.
	bcast broadcast.Broadcast
	// clients is the connected clients.
	clients map[*client]struct{}
	// closed is set once Close runs, so no client connects after it.
	closed bool
	// connected is set once a client has connected.
	connected bool
	// serial numbers input and configure events.
	serial uint32
	// windows is the mapped toplevels, bottom to top.
	windows []*surface
	// layers is the mapped layer surfaces of each layer, bottom to top.
	layers [layerCount][]*surface
	// usable is the screen area the exclusive zones leave to the windows.
	usable image.Rectangle
	// placed counts the windows placed on the cascade.
	placed int
	// active is the activated window, or nil.
	active *surface
	// focus is the surface holding the keyboard focus, or nil: the active
	// window, or a layer surface that takes the keyboard.
	focus *surface
	// layerFocus is the layer surface a press gave the keyboard, or nil.
	layerFocus *surface
	// hover is the surface holding the pointer focus, or nil.
	hover *surface
	// grab is the window move or resize in progress, or nil.
	grab *grab
	// popupGrabs is the popups that close on a press outside them, oldest
	// first.
	popupGrabs []*surface
	// viewers holds what each viewer's copy of the screen lacks.
	viewers map[*view]struct{}
	// dirty is set when a viewer's copy fell behind under the current lock.
	dirty bool
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
	screen := image.Rect(0, 0, width, height)
	return &Compositor{
		listener: l,
		screen:   screen,
		usable:   screen,
		start:    time.Now(),
		keymap:   k.NewMemfd("keymap", append([]byte(keymap), 0)),
		clients:  make(map[*client]struct{}),
		viewers:  make(map[*view]struct{}),
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
		c.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
			closed = c.closed
			if !closed {
				c.clients[cl] = struct{}{}
				c.connected = true
				broadcast()
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

	// Disconnect the clients, whose serve loops then clean up, and end the
	// waits and the viewers.
	var clients []*client
	c.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		c.closed = true
		for cl := range c.clients {
			clients = append(clients, cl)
		}
		broadcast()
	})
	for _, cl := range clients {
		_ = cl.conn.Close()
	}
	_ = c.keymap.Close()
}

// WaitClient waits until a client has connected, or until Close or ctx ends.
func (c *Compositor) WaitClient(ctx context.Context) error {
	for {
		var connected, closed bool
		var waitCh <-chan struct{}
		c.bcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			connected, closed, waitCh = c.connected, c.closed, getWaitCh()
		})
		switch {
		case connected:
			return nil
		case closed:
			return errors.New("compositor closed")
		}
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-waitCh:
		}
	}
}

// Run sends the whole screen to send, then the changes since each frame,
// until Close or ctx ends.
func (c *Compositor) Run(ctx context.Context, send func(*s4wave_vm.WasiVmDisplayFrame) error) error {
	// Register the viewer with the whole screen to draw.
	v := &view{damage: []image.Rectangle{c.screen}}
	c.bcast.HoldLock(func(func(), func() <-chan struct{}) {
		c.viewers[v] = struct{}{}
	})
	defer c.bcast.HoldLock(func(func(), func() <-chan struct{}) {
		delete(c.viewers, v)
	})

	// Send the changes each time the viewer falls behind.
	for {
		var frame *s4wave_vm.WasiVmDisplayFrame
		var closed bool
		var waitCh <-chan struct{}
		c.bcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			if !v.empty() {
				frame = c.readFrame(v)
			}
			closed, waitCh = c.closed, getWaitCh()
		})
		if closed {
			return nil
		}
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

// update runs fn under the lock, then wakes the viewers when fn damaged the
// screen.
func (c *Compositor) update(fn func()) {
	c.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		fn()
		if c.dirty {
			c.dirty = false
			broadcast()
		}
	})
}

// damage marks the screen area r for every viewer to draw again. The caller
// holds the lock through update, which wakes the viewers.
func (c *Compositor) damage(r image.Rectangle) {
	r = r.Intersect(c.screen)
	if r.Empty() {
		return
	}
	for v := range c.viewers {
		v.add(r)
	}
	c.dirty = true
}

// shift has every viewer move the screen area src by by, as far as both ends
// stay on the screen, and returns the area it moved to. The caller marks the
// areas the move uncovers, and holds the lock through update.
func (c *Compositor) shift(src image.Rectangle, by image.Point) image.Rectangle {
	// Keep the part whose source and destination are on the screen.
	src = src.Intersect(c.screen).Intersect(c.screen.Sub(by))
	if src.Empty() || by == (image.Point{}) {
		return image.Rectangle{}
	}

	// Queue the move for every viewer.
	for v := range c.viewers {
		v.copy(src, by)
	}
	c.dirty = true
	return src.Add(by)
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

// remove forgets a disconnected client and unmaps its windows and layer
// surfaces. The caller holds the lock.
func (c *Compositor) remove(cl *client) {
	// Drop the client and the focus it held, then hide its surfaces.
	delete(c.clients, cl)
	if c.hover != nil && c.hover.c == cl {
		c.hover = nil
	}
	if c.focus != nil && c.focus.c == cl {
		c.focus = nil
	}
	for _, w := range c.roots() {
		if w.c == cl {
			c.unmap(w)
		}
	}
	c.popupGrabs = slices.DeleteFunc(c.popupGrabs, func(p *surface) bool { return p.c == cl })
}
