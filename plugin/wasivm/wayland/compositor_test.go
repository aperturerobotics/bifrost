//go:build !js

package wayland

import (
	"encoding/binary"
	"image"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine/native"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// The object ids a connected testClient binds its globals to. The ids it
// allocates later start at firstFreeID.
const (
	idRegistry      = 2
	idCompositor    = 4
	idShm           = 5
	idWmBase        = 6
	idSeat          = 7
	idSubcompositor = 8
	idLayerShell    = 9
	firstFreeID     = 20
)

// testCompositor is a compositor on a fresh kernel, with a viewer whose
// frames build a copy of the screen.
type testCompositor struct {
	// t is the test.
	t *testing.T
	// k is the kernel.
	k *kernel.Kernel
	// comp is the compositor.
	comp *Compositor
	// frames is the viewer's frames.
	frames chan *s4wave_vm.WasiVmDisplayFrame
	// screen is the screen's RGBA pixels as the frames so far draw it.
	screen *image.RGBA
}

// newTestCompositor starts a compositor with a width by height screen and a
// viewer.
func newTestCompositor(t *testing.T, width, height int) *testCompositor {
	// Start a kernel.
	ctx := t.Context()
	rt := native.New(ctx)
	t.Cleanup(func() { _ = rt.Close(ctx) })
	k, err := kernel.New(ctx, rt, &kernel.Config{VirtualTerminal: true, Stdin: strings.NewReader("")})
	if err != nil {
		t.Fatal(err)
	}

	// Serve a compositor on it.
	comp, err := Listen(k, width, height)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(comp.Close)
	go func() { _ = comp.Serve(ctx) }()

	// Collect the viewer's frames.
	tc := &testCompositor{
		t:      t,
		k:      k,
		comp:   comp,
		frames: make(chan *s4wave_vm.WasiVmDisplayFrame, 64),
		screen: image.NewRGBA(image.Rect(0, 0, width, height)),
	}
	go func() {
		_ = comp.Run(ctx, func(f *s4wave_vm.WasiVmDisplayFrame) error {
			tc.frames <- f
			return nil
		})
	}()
	return tc
}

// awaitPixel applies the viewer's frames until the screen pixel at pt is
// want, in RGBA.
func (tc *testCompositor) awaitPixel(pt image.Point, want [4]byte) {
	tc.t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		if got := tc.screen.Pix[tc.screen.PixOffset(pt.X, pt.Y):][:4]; string(got) == string(want[:]) {
			return
		}
		select {
		case f := <-tc.frames:
			r := image.Rect(int(f.GetX()), int(f.GetY()), int(f.GetX()+f.GetWidth()), int(f.GetY()+f.GetHeight()))
			for y := range r.Dy() {
				copy(tc.screen.Pix[tc.screen.PixOffset(r.Min.X, r.Min.Y+y):], f.GetRgba()[4*y*r.Dx():][:4*r.Dx()])
			}
		case <-deadline:
			got := tc.screen.Pix[tc.screen.PixOffset(pt.X, pt.Y):][:4]
			tc.t.Fatalf("pixel at %v = %x, want %x", pt, got, want)
		}
	}
}

// pointer sends a viewer's pointer event.
func (tc *testCompositor) pointer(x, y int, buttons uint32) {
	tc.comp.Input(&s4wave_vm.WasiVmDisplayInput{Body: &s4wave_vm.WasiVmDisplayInput_Pointer{
		Pointer: &s4wave_vm.WasiVmPointerEvent{X: int32(x), Y: int32(y), Buttons: buttons}, //nolint:gosec // small test points.
	}})
}

// testClient speaks the Wayland wire protocol to the compositor as a guest
// client would, through a host connection to its socket.
type testClient struct {
	// t is the test.
	t *testing.T
	// k is the kernel, which holds the shared memory files.
	k *kernel.Kernel
	// conn is the connection.
	conn *kernel.UnixConn
	// buf is the bytes read and not yet parsed.
	buf []byte
	// files is the files received and not yet taken.
	files []*kernel.File
	// nextID is the next object id to allocate.
	nextID uint32
}

// connect connects a client and binds the globals it uses.
func (tc *testCompositor) connect() *testClient {
	// Connect and list the globals.
	tc.t.Helper()
	conn, err := tc.k.DialUnix(tc.t.Context(), SocketPath)
	if err != nil {
		tc.t.Fatal(err)
	}
	tc.t.Cleanup(func() { _ = conn.Close() })
	c := &testClient{t: tc.t, k: tc.k, conn: conn, nextID: firstFreeID}
	c.request(newEvent(1, displayGetRegistry).uint(idRegistry))
	sync := c.newID()
	c.request(newEvent(1, displaySync).uint(sync))

	// Note each global's name until the sync's callback.
	names := make(map[string]uint32)
	for {
		id, op, m := c.next()
		if id == sync && op == callbackDone {
			break
		}
		if id == idRegistry && op == registryGlobal {
			name, iface := m.uint(), m.string()
			names[iface] = name
		}
	}

	// Bind the globals.
	for _, g := range []struct {
		iface       string
		version, id uint32
	}{
		{"wl_compositor", 4, idCompositor},
		{"wl_shm", 1, idShm},
		{"xdg_wm_base", 1, idWmBase},
		{"wl_seat", 5, idSeat},
		{"wl_subcompositor", 1, idSubcompositor},
		{"zwlr_layer_shell_v1", 4, idLayerShell},
	} {
		c.request(newEvent(idRegistry, registryBind).uint(names[g.iface]).string(g.iface).uint(g.version).uint(g.id))
	}
	return c
}

// newID allocates an object id.
func (c *testClient) newID() uint32 {
	c.nextID++
	return c.nextID - 1
}

// request sends a request built as an event: the encodings match.
func (c *testClient) request(e *event, files ...*kernel.File) {
	c.t.Helper()
	if _, err := c.conn.WriteMsg(e.bytes(), files); err != nil {
		c.t.Fatal(err)
	}
}

// next reads the next event, returning its object, opcode and arguments.
func (c *testClient) next() (uint32, uint16, *message) {
	// Read until a whole event is buffered.
	c.t.Helper()
	for len(c.buf) < 8 || len(c.buf) < int(binary.LittleEndian.Uint16(c.buf[6:])) {
		chunk := make([]byte, 4096)
		n, files, err := c.conn.ReadMsg(chunk)
		if err != nil {
			c.t.Fatal(err)
		}
		c.buf = append(c.buf, chunk[:n]...)
		c.files = append(c.files, files...)
	}

	// Take it from the buffer.
	size := int(binary.LittleEndian.Uint16(c.buf[6:]))
	id, op := binary.LittleEndian.Uint32(c.buf), binary.LittleEndian.Uint16(c.buf[4:])
	m := &message{args: c.buf[8:size], files: &c.files}
	c.buf = c.buf[size:]
	return id, op, m
}

// await reads events until object id sends op, returning its arguments.
func (c *testClient) await(id uint32, op uint16) *message {
	c.t.Helper()
	for {
		if gotID, gotOp, m := c.next(); gotID == id && gotOp == op {
			return m
		} else if gotID == 1 && gotOp == displayError {
			c.t.Fatalf("protocol error on %d: code %d: %s", m.uint(), m.uint(), m.string())
		}
	}
}

// roundtrip waits until the compositor has handled every request sent.
func (c *testClient) roundtrip() {
	c.t.Helper()
	id := c.newID()
	c.request(newEvent(1, displaySync).uint(id))
	c.await(id, callbackDone)
}

// awaitStates reads events until toplevel id is configured, returning the
// size and the states.
func (c *testClient) awaitStates(id uint32) (image.Point, []uint32) {
	// Read the size and the array of states.
	c.t.Helper()
	m := c.await(id, toplevelConfigure)
	w, h, raw := m.int(), m.int(), m.bytes()

	// Decode the states.
	var states []uint32
	for i := 0; i+4 <= len(raw); i += 4 {
		states = append(states, binary.LittleEndian.Uint32(raw[i:]))
	}
	return image.Pt(int(w), int(h)), states
}

// buffer makes a width by height buffer of one pixel value, its bytes blue,
// green, red and alpha, in a pool of its own.
func (c *testClient) buffer(width, height int, bgra [4]byte, format uint32) uint32 {
	// Fill a shared memory file.
	pixels := make([]byte, 4*width*height)
	for i := 0; i < len(pixels); i += 4 {
		copy(pixels[i:], bgra[:])
	}
	pool := c.k.NewMemfd("pool", pixels)
	c.t.Cleanup(func() { _ = pool.Close() })

	// Share it, then make the buffer.
	pid, bid := c.newID(), c.newID()
	c.request(newEvent(idShm, shmCreatePool).uint(pid).int(int32(len(pixels))), pool) //nolint:gosec // small buffers.
	c.request(newEvent(pid, shmPoolCreateBuffer).uint(bid).int(0).
		int(int32(width)).int(int32(height)).int(int32(4 * width)).uint(format)) //nolint:gosec // small buffers.
	return bid
}

// testWindow is the object ids of a test client's window.
type testWindow struct {
	// surface, xdg and toplevel are the wl_surface, xdg_surface and
	// xdg_toplevel ids.
	surface, xdg, toplevel uint32
}

// window maps a width by height window of one opaque color, its bytes blue,
// green and red, after its first configure.
func (c *testClient) window(width, height int, bgr [3]byte) testWindow {
	// Make a toplevel and wait for its first configure, which leaves the
	// size to the client.
	c.t.Helper()
	w := testWindow{surface: c.newID(), xdg: c.newID(), toplevel: c.newID()}
	c.request(newEvent(idCompositor, compositorCreateSurface).uint(w.surface))
	c.request(newEvent(idWmBase, wmBaseGetXdgSurface).uint(w.xdg).uint(w.surface))
	c.request(newEvent(w.xdg, xdgSurfaceGetToplevel).uint(w.toplevel))
	c.request(newEvent(w.surface, surfaceCommit))
	if size, _ := c.awaitStates(w.toplevel); size != (image.Point{}) {
		c.t.Fatalf("toplevel configured at %v, want 0x0", size)
	}
	c.await(w.xdg, xdgSurfaceConfigure)

	// Map it with a buffer.
	bid := c.buffer(width, height, [4]byte{bgr[0], bgr[1], bgr[2], 0}, shmFormatXRGB8888)
	c.request(newEvent(w.surface, surfaceAttach).uint(bid).int(0).int(0))
	c.request(newEvent(w.surface, surfaceDamage).int(0).int(0).int(int32(width)).int(int32(height))) //nolint:gosec // small windows.
	c.request(newEvent(w.surface, surfaceCommit))
	c.await(bid, bufferRelease)
	return w
}

// layerSurface maps a layer surface of one opaque color in layer, held to
// anchor, with the exclusive zone and keyboard interactivity given. A zero
// side of size spans the screen between its anchors. It returns the
// wl_surface id.
func (c *testClient) layerSurface(layer, anchor uint32, size image.Point, zone int32, interactivity uint32, bgr [3]byte) uint32 {
	// Make the layer surface.
	c.t.Helper()
	sid, lid := c.newID(), c.newID()
	c.request(newEvent(idCompositor, compositorCreateSurface).uint(sid))
	c.request(newEvent(idLayerShell, layerShellGetLayerSurface).uint(lid).uint(sid).uint(0).uint(layer).string("test"))

	// Set its state and wait for its first configure.
	c.request(newEvent(lid, layerSurfaceSetAnchor).uint(anchor))
	c.request(newEvent(lid, layerSurfaceSetSize).uint(uint32(size.X)).uint(uint32(size.Y))) //nolint:gosec // small sizes.
	c.request(newEvent(lid, layerSurfaceSetExclusiveZone).int(zone))
	c.request(newEvent(lid, layerSurfaceSetKeyboardInteractivity).uint(interactivity))
	c.request(newEvent(sid, surfaceCommit))
	m := c.await(lid, layerSurfaceConfigure)
	_, w, h := m.uint(), int(m.uint()), int(m.uint())

	// Map it at the configured size.
	bid := c.buffer(w, h, [4]byte{bgr[0], bgr[1], bgr[2], 0}, shmFormatXRGB8888)
	c.request(newEvent(sid, surfaceAttach).uint(bid).int(0).int(0))
	c.request(newEvent(sid, surfaceCommit))
	c.await(bid, bufferRelease)
	return sid
}

// The colors the tests draw, in RGBA.
var (
	red   = [4]byte{0xff, 0, 0, 0xff}
	green = [4]byte{0, 0xff, 0, 0xff}
	blue  = [4]byte{0, 0, 0xff, 0xff}
)

// The colors the tests draw, as the blue, green and red bytes of a buffer.
var (
	bufRed   = [3]byte{0, 0, 0xff}
	bufGreen = [3]byte{0, 0xff, 0}
	bufBlue  = [3]byte{0xff, 0, 0}
)

// TestCompositor checks a client maps a window from a shared memory buffer
// that the viewer shows on the cascade, receives the keymap file and the
// keyboard focus, and gets pointer input in the window's coordinates.
func TestCompositor(t *testing.T) {
	// Map a window, which lands on the first cascade step.
	tc := newTestCompositor(t, 128, 96)
	c := tc.connect()
	w := c.window(16, 8, bufRed)
	tc.awaitPixel(image.Pt(cascadeStep, cascadeStep), red)
	tc.awaitPixel(image.Pt(cascadeStep-1, cascadeStep), background)

	// The keyboard gets the keymap file and enters the window.
	keyboard := c.newID()
	c.request(newEvent(idSeat, seatGetKeyboard).uint(keyboard))
	m := c.await(keyboard, keyboardKeymap)
	format, keymapFile, size := m.uint(), m.fd(), m.uint()
	if m.err != nil || format != keymapXKBV1 {
		t.Fatalf("keymap format %d, err %v", format, m.err)
	}

	// Read the keymap text from the file.
	text := make([]byte, size)
	_, _ = keymapFile.ReadAt(text, 0)
	_ = keymapFile.Close()
	if !strings.HasPrefix(string(text), "xkb_keymap {") {
		t.Fatalf("keymap starts %q", text[:min(len(text), 16)])
	}

	// Wait for the enter.
	m = c.await(keyboard, keyboardEnter)
	if _, s := m.uint(), m.uint(); s != w.surface {
		t.Fatalf("keyboard entered %d, want %d", s, w.surface)
	}

	// Pressing on the window enters the pointer at the window's point.
	pointer := c.newID()
	c.request(newEvent(idSeat, seatGetPointer).uint(pointer))
	c.roundtrip()
	tc.pointer(cascadeStep+5, cascadeStep+6, 1)
	m = c.await(pointer, pointerEnter)
	if _, s, x, y := m.uint(), m.uint(), m.int(), m.int(); s != w.surface || x != 5*256 || y != 6*256 {
		t.Fatalf("pointer entered %d at %d,%d, want %d at 5,6 in 24.8", s, x, y, w.surface)
	}

	// The motion and the press follow.
	m = c.await(pointer, pointerMotion)
	if _, x, y := m.uint(), m.int(), m.int(); x != 5*256 || y != 6*256 {
		t.Fatalf("motion to %d,%d, want 5,6 in 24.8", x, y)
	}
	m = c.await(pointer, pointerButton)
	if _, _, button, state := m.uint(), m.uint(), m.uint(), m.uint(); button != 0x110 || state != 1 {
		t.Fatalf("button %#x state %d, want 0x110 pressed", button, state)
	}
}

// TestStacking checks a new window maps above the last and takes the
// keyboard, and a press on a lower window raises and activates it.
func TestStacking(t *testing.T) {
	// Map two overlapping windows, the second on top and active.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	first := c.window(48, 48, bufRed)
	second := c.window(48, 48, bufGreen)
	overlap := image.Pt(2*cascadeStep+4, 2*cascadeStep+4)
	tc.awaitPixel(overlap, green)
	if _, states := c.awaitStates(second.toplevel); !slices.Contains(states, toplevelActivated) {
		t.Fatalf("second window configured with states %v, want activated", states)
	}

	// Press on the first window where it shows.
	tc.pointer(cascadeStep+4, cascadeStep+4, 1)
	tc.pointer(cascadeStep+4, cascadeStep+4, 0)

	// It rises over the second and becomes active.
	tc.awaitPixel(overlap, red)
	if _, states := c.awaitStates(first.toplevel); !slices.Contains(states, toplevelActivated) {
		t.Fatalf("first window configured with states %v, want activated", states)
	}
}

// TestMoveGrab checks a window moves with the pointer after its client asks
// to move it during a press.
func TestMoveGrab(t *testing.T) {
	// Map a window.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	w := c.window(32, 32, bufBlue)

	// Press on it, noting the press's serial.
	pointer := c.newID()
	c.request(newEvent(idSeat, seatGetPointer).uint(pointer))
	c.roundtrip()
	tc.pointer(cascadeStep+8, cascadeStep+8, 1)
	m := c.await(pointer, pointerButton)
	serial := m.uint()

	// Ask to move it, then drag the pointer by 20,10 and release.
	c.request(newEvent(w.toplevel, toplevelMove).uint(idSeat).uint(serial))
	c.roundtrip()
	tc.pointer(cascadeStep+28, cascadeStep+18, 1)
	tc.pointer(cascadeStep+28, cascadeStep+18, 0)

	// The window follows, leaving the background behind.
	tc.awaitPixel(image.Pt(cascadeStep+20+31, cascadeStep+10+31), blue)
	tc.awaitPixel(image.Pt(cascadeStep, cascadeStep), background)
}

// TestSubsurface checks a translucent subsurface blends over its parent at
// the position the parent's commit applies.
func TestSubsurface(t *testing.T) {
	// Map a blue window.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	w := c.window(32, 32, bufBlue)

	// Put a half transparent gray subsurface on it at 4,4.
	sub, subsurface := c.newID(), c.newID()
	c.request(newEvent(idCompositor, compositorCreateSurface).uint(sub))
	c.request(newEvent(idSubcompositor, subcompositorGetSubsurface).uint(subsurface).uint(sub).uint(w.surface))
	c.request(newEvent(subsurface, subsurfaceSetPosition).int(4).int(4))
	bid := c.buffer(8, 8, [4]byte{0x80, 0x80, 0x80, 0x80}, shmFormatARGB8888)
	c.request(newEvent(sub, surfaceAttach).uint(bid).int(0).int(0))
	c.request(newEvent(sub, surfaceCommit))
	c.request(newEvent(w.surface, surfaceCommit))

	// The gray adds to the blue scaled by the gray's transparency.
	tc.awaitPixel(image.Pt(cascadeStep+4, cascadeStep+4), [4]byte{0x80, 0x80, 0xff, 0xff})
	tc.awaitPixel(image.Pt(cascadeStep+3, cascadeStep+3), blue)
}

// TestLayerShell checks a panel's exclusive zone keeps the windows below it,
// a maximized window fills the area left, an overlay draws above the windows
// and an exclusive overlay takes the keyboard.
func TestLayerShell(t *testing.T) {
	// Map a 16 pixel panel across the top of the screen.
	const panel = 16
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	c.layerSurface(layerTop, anchorTop|anchorLeft|anchorRight, image.Pt(0, panel), panel, interactivityNone, bufRed)
	tc.awaitPixel(image.Pt(159, panel-1), red)

	// A window cascades from below the panel and takes the keyboard.
	keyboard := c.newID()
	c.request(newEvent(idSeat, seatGetKeyboard).uint(keyboard))
	w := c.window(16, 16, bufGreen)
	tc.awaitPixel(image.Pt(cascadeStep, panel+cascadeStep), green)
	m := c.await(keyboard, keyboardEnter)
	if _, s := m.uint(), m.uint(); s != w.surface {
		t.Fatalf("keyboard entered %d, want the window %d", s, w.surface)
	}

	// Maximized, it fills the screen below the panel.
	c.request(newEvent(w.toplevel, toplevelSetMaximized))
	if size, _ := c.awaitStates(w.toplevel); size != image.Pt(160, 120-panel) {
		t.Fatalf("maximized to %v, want 160x%d", size, 120-panel)
	}

	// An overlay over the whole screen draws above the window and takes the
	// keyboard from it.
	overlay := c.layerSurface(layerOverlay, anchorTop|anchorBottom|anchorLeft|anchorRight, image.Point{}, -1, interactivityExclusive, bufBlue)
	tc.awaitPixel(image.Pt(cascadeStep, panel+cascadeStep), blue)
	tc.awaitPixel(image.Pt(0, 0), blue)
	c.await(keyboard, keyboardLeave)
	m = c.await(keyboard, keyboardEnter)
	if _, s := m.uint(), m.uint(); s != overlay {
		t.Fatalf("keyboard entered %d, want the overlay %d", s, overlay)
	}
}
