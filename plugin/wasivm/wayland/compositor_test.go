package wayland

import (
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel"
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
	idOutput        = 10
	idXdgOutputs    = 11
	idDecorations   = 12
	firstFreeID     = 20
)

// testCompositor is a compositor on a fresh kernel, with a viewer of the
// desktop.
type testCompositor struct {
	// testView is the desktop's viewer.
	*testView
	// k is the kernel.
	k *kernel.Kernel
	// comp is the compositor.
	comp *Compositor
}

// testView is a viewer that draws the scene updates it receives on a copy of
// a space, as the browser viewer does.
type testView struct {
	// t is the test.
	t *testing.T
	// updates is the viewer's scene updates.
	updates chan *s4wave_vm.WasiVmScene
	// scene is the screen, the layers, the roots and the grab the updates so
	// far set.
	scene *s4wave_vm.WasiVmScene
	// textures is each layer's texture by layer id.
	textures map[uint32]*image.RGBA
	// screen is the space's RGBA pixels as the scene so far draws it.
	screen *image.RGBA
	// patched counts the texture pixels the updates so far carried.
	patched int
	// done receives the error that ended the stream.
	done chan error
}

// newTestView starts a viewer of the stream run.
func newTestView(t *testing.T, run func(context.Context, func(*s4wave_vm.WasiVmScene) error) error) *testView {
	tv := &testView{
		t:        t,
		updates:  make(chan *s4wave_vm.WasiVmScene, 64),
		scene:    &s4wave_vm.WasiVmScene{},
		textures: make(map[uint32]*image.RGBA),
		screen:   image.NewRGBA(image.Rectangle{}),
		done:     make(chan error, 1),
	}
	go func() {
		tv.done <- run(t.Context(), func(u *s4wave_vm.WasiVmScene) error {
			tv.updates <- u
			return nil
		})
	}()
	return tv
}

// newTestCompositor starts a compositor with a width by height screen and a
// viewer.
func newTestCompositor(t *testing.T, width, height int) *testCompositor {
	// Start a kernel.
	ctx := t.Context()
	k, err := kernel.New(ctx, testRuntime(t, ctx), &kernel.Config{VirtualTerminal: true, Stdin: strings.NewReader("")})
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

	// View the desktop.
	return &testCompositor{testView: newTestView(t, comp.Run), k: k, comp: comp}
}

// awaitPixel applies the viewer's updates until the pixel at pt is want, in
// RGBA.
func (tc *testView) awaitPixel(pt image.Point, want [4]byte) {
	tc.t.Helper()
	deadline := time.After(10 * time.Second)
	for {
		if pt.In(tc.screen.Rect) && string(tc.screen.Pix[tc.screen.PixOffset(pt.X, pt.Y):][:4]) == string(want[:]) {
			return
		}
		select {
		case u := <-tc.updates:
			tc.apply(u)
		case <-deadline:
			tc.t.Fatalf("pixel at %v in %v = %x, want %x", pt, tc.screen.Rect, tc.screen.At(pt.X, pt.Y), want)
		}
	}
}

// awaitScene applies the viewer's updates until ok holds for the scene so
// far, failing with what when it never does.
func (tc *testView) awaitScene(what string, ok func(*s4wave_vm.WasiVmScene) bool) {
	tc.t.Helper()
	deadline := time.After(10 * time.Second)
	for !ok(tc.scene) {
		select {
		case u := <-tc.updates:
			tc.apply(u)
		case <-deadline:
			tc.t.Fatalf("the scene never showed %s: %v", what, tc.scene)
		}
	}
}

// rootAt returns the position of the root id in the scene so far.
func rootAt(sc *s4wave_vm.WasiVmScene, id uint32) (image.Point, bool) {
	for _, r := range sc.GetRoots().GetRoots() {
		if r.GetId() == id {
			return image.Pt(int(r.GetX()), int(r.GetY())), true
		}
	}
	return image.Point{}, false
}

// apply applies the update u as the browser viewer does: the screen, the
// layers, the roots, the grab and the frames it sets, then its patches, and draws the
// scene.
func (tc *testView) apply(u *s4wave_vm.WasiVmScene) {
	// Take the screen, the layers and the roots, dropping the textures of
	// layers gone and giving a resized layer a new texture.
	if u.GetScreen() != nil {
		tc.scene.Screen = u.GetScreen()
	}
	if u.GetRoots() != nil {
		tc.scene.Roots = u.GetRoots()
	}
	if u.GetGrab() != nil {
		tc.scene.Grab = u.GetGrab()
	}
	if u.GetFrames() != nil {
		tc.scene.Frames = u.GetFrames()
	}
	if u.GetLayers() != nil {
		tc.scene.Layers = u.GetLayers()
		textures := make(map[uint32]*image.RGBA)
		for _, l := range u.GetLayers().GetLayers() {
			size := image.Pt(int(l.GetTextureWidth()), int(l.GetTextureHeight()))
			if t := tc.textures[l.GetId()]; t != nil && t.Rect.Size() == size {
				textures[l.GetId()] = t
			} else {
				textures[l.GetId()] = image.NewRGBA(image.Rectangle{Max: size})
			}
		}
		tc.textures = textures
	}

	// Draw the patches into the textures.
	for _, p := range u.GetPatches() {
		t := tc.textures[p.GetLayer()]
		if t == nil {
			tc.t.Fatalf("a patch for layer %d, which the scene lacks", p.GetLayer())
		}
		r := image.Rect(int(p.GetX()), int(p.GetY()), int(p.GetX()+p.GetWidth()), int(p.GetY()+p.GetHeight()))
		for y := range r.Dy() {
			copy(t.Pix[t.PixOffset(r.Min.X, r.Min.Y+y):], p.GetRgba()[4*y*r.Dx():][:4*r.Dx()])
		}
		tc.patched += r.Dx() * r.Dy()
	}
	tc.draw()
}

// draw composites the scene on the screen at its scale: the background, then
// each layer at its root's position plus its offset, sampling its texture at
// the pixel nearest each one drawn and blending its premultiplied pixels.
func (tc *testView) draw() {
	// Fill the background.
	sc := tc.scene.GetScreen()
	n := max(int(sc.GetScale()), 1)
	tc.screen = image.NewRGBA(image.Rect(0, 0, int(sc.GetWidth())*n, int(sc.GetHeight())*n))
	var bg [4]byte
	binary.BigEndian.PutUint32(bg[:], sc.GetBackground())
	for i := 0; i < len(tc.screen.Pix); i += 4 {
		copy(tc.screen.Pix[i:], bg[:])
	}

	// Draw each layer.
	roots := make(map[uint32]image.Point)
	for _, r := range tc.scene.GetRoots().GetRoots() {
		roots[r.GetId()] = image.Pt(int(r.GetX()), int(r.GetY()))
	}
	for _, l := range tc.scene.GetLayers().GetLayers() {
		t := tc.textures[l.GetId()]
		at := roots[l.GetRoot()].Add(image.Pt(int(l.GetX()), int(l.GetY()))).Mul(n)
		size := image.Pt(int(l.GetWidth()), int(l.GetHeight())).Mul(n)
		in := image.Rectangle{Min: at, Max: at.Add(size)}.Intersect(tc.screen.Rect)
		for y := in.Min.Y; y < in.Max.Y; y++ {
			for x := in.Min.X; x < in.Max.X; x++ {
				from := t.Pix[t.PixOffset((x-at.X)*t.Rect.Dx()/size.X, (y-at.Y)*t.Rect.Dy()/size.Y):][:4]
				to := tc.screen.Pix[tc.screen.PixOffset(x, y):][:4]
				if l.GetOpaque() {
					copy(to, from)
					continue
				}
				keep := 255 - uint32(from[3])
				for j := range 4 {
					to[j] = from[j] + byte((uint32(to[j])*keep+127)/255) //nolint:gosec // at most 255.
				}
			}
		}
	}
}

// topLayer returns the topmost layer the scene so far shows, or nil.
func (tc *testView) topLayer() *s4wave_vm.WasiVmSceneLayer {
	layers := tc.scene.GetLayers().GetLayers()
	if len(layers) == 0 {
		return nil
	}
	return layers[len(layers)-1]
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
		{"wl_output", 4, idOutput},
		{"zxdg_output_manager_v1", 3, idXdgOutputs},
		{"zxdg_decoration_manager_v1", 1, idDecorations},
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
	pixels := make([]byte, 4*width*height)
	for i := 0; i < len(pixels); i += 4 {
		copy(pixels[i:], bgra[:])
	}
	return c.bufferOf(width, height, 4*width, pixels, format)
}

// bufferOf makes a width by height buffer of pixels, rows stride bytes apart
// of blue, green, red and alpha bytes, in a pool of its own.
func (c *testClient) bufferOf(width, height, stride int, pixels []byte, format uint32) uint32 {
	// Fill a shared memory file.
	pool := c.k.NewMemfd("pool", pixels)
	c.t.Cleanup(func() { _ = pool.Close() })

	// Share it, then make the buffer.
	pid, bid := c.newID(), c.newID()
	c.request(newEvent(idShm, shmCreatePool).uint(pid).int(int32(len(pixels))), pool) //nolint:gosec // small buffers.
	c.request(newEvent(pid, shmPoolCreateBuffer).uint(bid).int(0).
		int(int32(width)).int(int32(height)).int(int32(stride)).uint(format)) //nolint:gosec // small buffers.
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
	c.t.Helper()
	w := c.newWindow()
	c.mapWindow(w, width, height, bgr, 0)
	return w
}

// decoratedWindow maps a window as window does whose client asks for
// decorations in mode, or leaves the mode to the compositor when mode is 0.
// It returns the window and the mode the compositor configures.
func (c *testClient) decoratedWindow(width, height int, bgr [3]byte, mode uint32) (testWindow, uint32) {
	// Make the window and its decoration before the first commit.
	c.t.Helper()
	w := c.newWindow()
	decoration := c.newID()
	c.request(newEvent(idDecorations, decorationManagerGetToplevelDecoration).uint(decoration).uint(w.toplevel))
	if mode != 0 {
		c.request(newEvent(decoration, decorationSetMode).uint(mode))
	}
	return w, c.mapWindow(w, width, height, bgr, decoration)
}

// newWindow makes a toplevel with no buffer.
func (c *testClient) newWindow() testWindow {
	// Make the wl_surface, its xdg_surface and the toplevel.
	w := testWindow{surface: c.newID(), xdg: c.newID(), toplevel: c.newID()}
	c.request(newEvent(idCompositor, compositorCreateSurface).uint(w.surface))
	c.request(newEvent(idWmBase, wmBaseGetXdgSurface).uint(w.xdg).uint(w.surface))
	c.request(newEvent(w.xdg, xdgSurfaceGetToplevel).uint(w.toplevel))
	return w
}

// mapWindow maps the window w at width by height in one opaque color after
// its first configure, which leaves the size to the client. When decoration
// is not 0 it returns the mode that decoration's configure gives.
func (c *testClient) mapWindow(w testWindow, width, height int, bgr [3]byte, decoration uint32) uint32 {
	// Commit, and wait for the first configure.
	c.t.Helper()
	c.request(newEvent(w.surface, surfaceCommit))
	var mode uint32
	if decoration != 0 {
		mode = c.await(decoration, decorationConfigure).uint()
	}
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
	return mode
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
	// Map a window showing a pattern.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	w := c.window(32, 32, bufBlue)
	pix := pattern(32, 32, 0)
	c.request(newEvent(w.surface, surfaceAttach).uint(c.bufferOf(32, 32, 4*32, pix, shmFormatXRGB8888)).int(0).int(0))
	c.request(newEvent(w.surface, surfaceDamage).int(0).int(0).int(32).int(32))
	c.request(newEvent(w.surface, surfaceCommit))
	tc.awaitPattern(image.Pt(cascadeStep, cascadeStep), 32, 32, pix)

	// Drag it by 20,10 and release, noting the pixels sent so far.
	patched := tc.patched
	tc.grabMove(c, w, image.Pt(cascadeStep+8, cascadeStep+8))
	tc.pointer(cascadeStep+28, cascadeStep+18, 1)
	tc.pointer(cascadeStep+28, cascadeStep+18, 0)

	// The window follows, leaving the background behind. The viewer moves
	// the window's texture rather than receiving its pixels again.
	tc.awaitPattern(image.Pt(cascadeStep+20, cascadeStep+10), 32, 32, pix)
	tc.awaitPixel(image.Pt(cascadeStep, cascadeStep), background)
	if tc.patched != patched {
		t.Fatalf("the move sent %d pixels, want none", tc.patched-patched)
	}
}

// TestMoveGrabLate checks a window whose client asks to move it after the
// pointer has already moved catches up with the pointer, as the request
// takes a round trip through the client.
func TestMoveGrabLate(t *testing.T) {
	// Map a window showing a pattern.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	w := c.window(32, 32, bufBlue)
	pix := pattern(32, 32, 0)
	c.request(newEvent(w.surface, surfaceAttach).uint(c.bufferOf(32, 32, 4*32, pix, shmFormatXRGB8888)).int(0).int(0))
	c.request(newEvent(w.surface, surfaceDamage).int(0).int(0).int(32).int(32))
	c.request(newEvent(w.surface, surfaceCommit))
	tc.awaitPattern(image.Pt(cascadeStep, cascadeStep), 32, 32, pix)

	// Press on it and drag by 20,10 before asking to move it.
	pointer := c.newID()
	c.request(newEvent(idSeat, seatGetPointer).uint(pointer))
	c.roundtrip()
	tc.pointer(cascadeStep+8, cascadeStep+8, 1)
	serial := c.await(pointer, pointerButton).uint()
	tc.pointer(cascadeStep+28, cascadeStep+18, 1)
	c.request(newEvent(w.toplevel, toplevelMove).uint(idSeat).uint(serial))
	c.roundtrip()

	// The window jumps under the pointer without waiting for it to move,
	// then keeps the offset of the press.
	tc.awaitPattern(image.Pt(cascadeStep+20, cascadeStep+10), 32, 32, pix)
	tc.pointer(cascadeStep+38, cascadeStep+28, 1)
	tc.pointer(cascadeStep+38, cascadeStep+28, 0)
	tc.awaitPattern(image.Pt(cascadeStep+30, cascadeStep+20), 32, 32, pix)
}

// TestMoveUnderPanel checks a window moved under a panel stays below it.
func TestMoveUnderPanel(t *testing.T) {
	// Map a red panel along the bottom.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	const panel = 16
	c.layerSurface(layerTop, anchorBottom|anchorLeft|anchorRight, image.Pt(0, panel), 0, interactivityNone, bufRed)

	// Map a window showing a pattern.
	w := c.window(32, 32, bufBlue)
	pix := pattern(32, 32, 0)
	c.request(newEvent(w.surface, surfaceAttach).uint(c.bufferOf(32, 32, 4*32, pix, shmFormatXRGB8888)).int(0).int(0))
	c.request(newEvent(w.surface, surfaceDamage).int(0).int(0).int(32).int(32))
	c.request(newEvent(w.surface, surfaceCommit))
	tc.awaitPattern(image.Pt(cascadeStep, cascadeStep), 32, 32, pix)

	// Drag it down until its lower half is under the panel, then release.
	start := image.Pt(cascadeStep+8, cascadeStep+8)
	tc.grabMove(c, w, start)
	to := image.Pt(cascadeStep+10, 120-panel-16)
	tc.pointer(start.X+to.X-cascadeStep, start.Y+to.Y-cascadeStep, 1)
	tc.pointer(start.X+to.X-cascadeStep, start.Y+to.Y-cascadeStep, 0)

	// The window's upper half shows above the panel, and the panel covers
	// the rest.
	tc.awaitPattern(to, 32, 16, pix[:4*32*16])
	tc.awaitPixel(image.Pt(to.X+8, 120-panel+4), red)
	tc.awaitPixel(image.Pt(cascadeStep, cascadeStep), background)
}

// TestMoveRule checks the scene's move places the window where the
// compositor does for each pointer position, including one that would lift
// the window's top under a panel, so a viewer can draw the move from its own
// pointer.
func TestMoveRule(t *testing.T) {
	// Map a panel along the top and a window below it.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	const panel = 16
	c.layerSurface(layerTop, anchorTop|anchorLeft|anchorRight, image.Pt(0, panel), panel, interactivityNone, bufRed)
	w := c.window(32, 32, bufBlue)
	at := image.Pt(cascadeStep, panel+cascadeStep)
	tc.awaitPixel(at, blue)

	// Grab it, and wait for the scene to show the move.
	press := at.Add(image.Pt(8, 8))
	tc.grabMove(c, w, press)
	var move *s4wave_vm.WasiVmSceneMove
	tc.awaitScene("the move", func(sc *s4wave_vm.WasiVmScene) bool {
		move = sc.GetGrab().GetMove()
		return move != nil
	})
	if got := image.Pt(int(move.GetPressX()), int(move.GetPressY())); got != press {
		t.Fatalf("the move's press is %v, want %v", got, press)
	}

	// Move the pointer down, then up past the panel. The window's root
	// lands where the rule puts it.
	for _, to := range []image.Point{{60, 70}, {90, 4}} {
		tc.pointer(to.X, to.Y, 1)
		want := image.Pt(
			int(move.GetX())+to.X-int(move.GetPressX()),
			max(int(move.GetY())+to.Y-int(move.GetPressY()), int(move.GetMinY())),
		)
		tc.awaitScene(fmt.Sprintf("the window at %v", want), func(sc *s4wave_vm.WasiVmScene) bool {
			got, ok := rootAt(sc, move.GetRoot())
			return ok && got == want
		})
	}
	if move.GetMinY() != panel {
		t.Fatalf("the move's least top is %d, want the panel's height %d", move.GetMinY(), panel)
	}

	// Release. The grab ends.
	tc.pointer(90, 4, 0)
	tc.awaitScene("the grab ended", func(sc *s4wave_vm.WasiVmScene) bool {
		return sc.GetGrab() != nil && sc.GetGrab().GetMove() == nil
	})
}

// TestOpaqueRegion checks a window with an alpha channel and an opaque
// region over all of it shows as an opaque layer, like a window without
// alpha.
func TestOpaqueRegion(t *testing.T) {
	// Map a window showing an opaque pattern in a format with alpha.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	w := c.window(32, 32, bufBlue)
	pix := pattern(32, 32, 0)
	for i := 3; i < len(pix); i += 4 {
		pix[i] = 0xff
	}

	// Mark all of it opaque.
	opaque := c.newID()
	c.request(newEvent(idCompositor, compositorCreateRegion).uint(opaque))
	c.request(newEvent(opaque, regionAdd).int(0).int(0).int(32).int(32))
	c.request(newEvent(w.surface, surfaceSetOpaqueRegion).uint(opaque))

	// Draw it. Its layer is opaque.
	c.request(newEvent(w.surface, surfaceAttach).uint(c.bufferOf(32, 32, 4*32, pix, shmFormatARGB8888)).int(0).int(0))
	c.request(newEvent(w.surface, surfaceDamage).int(0).int(0).int(32).int(32))
	c.request(newEvent(w.surface, surfaceCommit))
	tc.awaitPattern(image.Pt(cascadeStep, cascadeStep), 32, 32, pix)
	if l := tc.topLayer(); !l.GetOpaque() {
		t.Fatalf("the window's layer %v is not opaque", l)
	}
}

// grabMove presses on the window w at the screen point at and asks to move
// it, so pointer motion with the button held drags it.
func (tc *testCompositor) grabMove(c *testClient, w testWindow, at image.Point) {
	// Press on it, noting the press's serial.
	tc.t.Helper()
	pointer := c.newID()
	c.request(newEvent(idSeat, seatGetPointer).uint(pointer))
	c.roundtrip()
	tc.pointer(at.X, at.Y, 1)
	m := c.await(pointer, pointerButton)
	serial := m.uint()

	// Ask to move it.
	c.request(newEvent(w.toplevel, toplevelMove).uint(idSeat).uint(serial))
	c.roundtrip()
}

// TestUnmapNullBuffer checks a window a null buffer unmaps leaves the
// background where it was.
func TestUnmapNullBuffer(t *testing.T) {
	// Map a window, then commit no buffer.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	w := c.window(32, 32, bufBlue)
	tc.awaitPixel(image.Pt(cascadeStep+8, cascadeStep+8), blue)
	c.request(newEvent(w.surface, surfaceAttach).uint(0).int(0).int(0))
	c.request(newEvent(w.surface, surfaceCommit))

	// The background shows again.
	tc.awaitPixel(image.Pt(cascadeStep+8, cascadeStep+8), background)
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

// TestInputRegion checks the pointer passes through the parts of a
// subsurface outside its input region to the surface below.
func TestInputRegion(t *testing.T) {
	// Map a window with an 8x8 subsurface at 4,4 and take a pointer.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	w := c.window(32, 32, bufBlue)
	pointer := c.newID()
	c.request(newEvent(idSeat, seatGetPointer).uint(pointer))

	// Add the subsurface.
	sub, subsurface := c.newID(), c.newID()
	c.request(newEvent(idCompositor, compositorCreateSurface).uint(sub))
	c.request(newEvent(idSubcompositor, subcompositorGetSubsurface).uint(subsurface).uint(sub).uint(w.surface))
	c.request(newEvent(subsurface, subsurfaceSetPosition).int(4).int(4))
	bid := c.buffer(8, 8, [4]byte{0, 0, 0xff, 0xff}, shmFormatARGB8888)

	// Give it an empty input region, destroying the region before the
	// commit, which keeps the region set.
	empty := c.newID()
	c.request(newEvent(idCompositor, compositorCreateRegion).uint(empty))
	c.request(newEvent(sub, surfaceSetInputRegion).uint(empty))
	c.request(newEvent(empty, regionDestroy))
	c.request(newEvent(sub, surfaceAttach).uint(bid).int(0).int(0))
	c.request(newEvent(sub, surfaceCommit))
	c.request(newEvent(w.surface, surfaceCommit))
	c.roundtrip()

	// The pointer over it enters the window.
	tc.pointer(cascadeStep+6, cascadeStep+6, 0)
	m := c.await(pointer, pointerEnter)
	if _, s := m.uint(), m.uint(); s != w.surface {
		t.Fatalf("pointer entered %d, want the window %d", s, w.surface)
	}

	// Set a region of its left half. The pointer, at 2,2 on it, enters it
	// once the region applies.
	half := c.newID()
	c.request(newEvent(idCompositor, compositorCreateRegion).uint(half))
	c.request(newEvent(half, regionAdd).int(0).int(0).int(8).int(8))
	c.request(newEvent(half, regionSubtract).int(4).int(0).int(4).int(8))
	c.request(newEvent(sub, surfaceSetInputRegion).uint(half))
	c.request(newEvent(sub, surfaceCommit))
	m = c.await(pointer, pointerEnter)
	if _, s := m.uint(), m.uint(); s != sub {
		t.Fatalf("pointer entered %d, want the subsurface %d", s, sub)
	}

	// The right half passes the pointer through to the window.
	tc.pointer(cascadeStep+10, cascadeStep+6, 0)
	m = c.await(pointer, pointerEnter)
	if _, s := m.uint(), m.uint(); s != w.surface {
		t.Fatalf("pointer entered %d, want the window %d", s, w.surface)
	}
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

// TestXdgOutput checks the output and an xdg output name the screen, the xdg
// output describes its logical size, and the output's done event ends the
// description.
func TestXdgOutput(t *testing.T) {
	// The output names itself when bound.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	if name := c.await(idOutput, outputName).string(); name != "WASIVM-1" {
		t.Fatalf("output name %q, want WASIVM-1", name)
	}
	c.await(idOutput, outputDone)

	// Describe the output.
	x := c.newID()
	c.request(newEvent(idXdgOutputs, xdgOutputManagerGetXdgOutput).uint(x).uint(idOutput))

	// It reads as the whole screen, then the output's done ends it.
	m := c.await(x, xdgOutputLogicalSize)
	if w, h := m.int(), m.int(); w != 160 || h != 120 {
		t.Fatalf("logical size %dx%d, want 160x120", w, h)
	}
	if name := c.await(x, xdgOutputName).string(); name != "WASIVM-1" {
		t.Fatalf("name %q, want WASIVM-1", name)
	}
	c.await(idOutput, outputDone)
}

// TestWaitClient checks WaitClient returns once a client connects, and Close
// ends a viewer's Run.
func TestWaitClient(t *testing.T) {
	// No client has connected yet.
	tc := newTestCompositor(t, 160, 120)
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	if err := tc.comp.WaitClient(ctx); err == nil {
		t.Fatal("WaitClient returned before a client connected")
	}

	// A client connects.
	tc.connect()
	if err := tc.comp.WaitClient(t.Context()); err != nil {
		t.Fatal(err)
	}

	// Close ends a viewer.
	done := make(chan error, 1)
	go func() {
		done <- tc.comp.Run(t.Context(), func(*s4wave_vm.WasiVmScene) error { return nil })
	}()
	tc.comp.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Run continued after Close")
	}
}

// pattern returns the buffer pixels of a width by height image whose pixel at
// x,y has its own blue, green and red bytes, so a misplaced row or column
// shows as a wrong color.
func pattern(width, height, seed int) []byte {
	pixels := make([]byte, 4*width*height)
	for y := range height {
		for x := range width {
			px := pixels[4*(y*width+x):]
			px[0], px[1], px[2] = byte(x+seed), byte(y+2*seed), byte(x^y+3*seed) //nolint:gosec // wraps on purpose.
		}
	}
	return pixels
}

// awaitPattern applies the viewer's frames until the screen area at origin
// shows the RGBA of the buffer pixels of a width by height pattern.
func (tc *testView) awaitPattern(origin image.Point, width, height int, pixels []byte) {
	tc.t.Helper()
	for y := range height {
		for x := range width {
			px := pixels[4*(y*width+x):]
			tc.awaitPixel(origin.Add(image.Pt(x, y)), [4]byte{px[2], px[1], px[0], 0xff})
		}
	}
}

// TestPattern checks the screen shows each pixel of a window's buffer in
// place, through a full commit and a commit damaging part of a buffer whose
// rows are wider than the window.
func TestPattern(t *testing.T) {
	// Map a window.
	const width, height = 300, 70
	tc := newTestCompositor(t, 400, 200)
	c := tc.connect()
	w := c.window(width, height, bufRed)

	// Commit a pattern over all of it.
	first := pattern(width, height, 0)
	bid := c.bufferOf(width, height, 4*width, first, shmFormatXRGB8888)
	c.request(newEvent(w.surface, surfaceAttach).uint(bid).int(0).int(0))
	c.request(newEvent(w.surface, surfaceDamage).int(0).int(0).int(width).int(height))
	c.request(newEvent(w.surface, surfaceCommit))
	origin := image.Pt(cascadeStep, cascadeStep)
	tc.awaitPattern(origin, width, height, first)

	// Make a second pattern in a buffer whose rows have 40 bytes of padding.
	const stride = 4*width + 40
	second := pattern(width, height, 1)
	padded := make([]byte, stride*height)
	for y := range height {
		copy(padded[y*stride:], second[4*y*width:][:4*width])
	}
	padID := c.bufferOf(width, height, stride, padded, shmFormatXRGB8888)

	// Commit it, damaging only part of it.
	damage := image.Rect(17, 9, 263, 51)
	c.request(newEvent(w.surface, surfaceAttach).uint(padID).int(0).int(0))
	c.request(newEvent(w.surface, surfaceDamage).int(int32(damage.Min.X)).int(int32(damage.Min.Y)).int(int32(damage.Dx())).int(int32(damage.Dy()))) //nolint:gosec // small rectangles.
	c.request(newEvent(w.surface, surfaceCommit))
	c.await(padID, bufferRelease)

	// The damaged area shows the second pattern and the rest the first.
	want := slices.Clone(first)
	for y := damage.Min.Y; y < damage.Max.Y; y++ {
		i := 4 * (y*width + damage.Min.X)
		copy(want[i:i+4*damage.Dx()], second[i:])
	}
	tc.awaitPattern(origin, width, height, want)
}

// TestRestoreRedraw checks a window leaving the maximized state redraws the
// screen it covered, before its client draws at the restored size.
func TestRestoreRedraw(t *testing.T) {
	// Map a window and maximize it.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	w := c.window(16, 16, bufRed)
	c.request(newEvent(w.toplevel, toplevelSetMaximized))
	c.awaitSize(w.toplevel, image.Pt(160, 120))

	// Draw it at the screen's size.
	bid := c.buffer(160, 120, [4]byte{bufRed[0], bufRed[1], bufRed[2], 0}, shmFormatXRGB8888)
	c.request(newEvent(w.surface, surfaceAttach).uint(bid).int(0).int(0))
	c.request(newEvent(w.surface, surfaceDamage).int(0).int(0).int(160).int(120))
	c.request(newEvent(w.surface, surfaceCommit))
	tc.awaitPixel(image.Pt(4, 4), red)

	// Restored, the corner it left shows the background at once.
	c.request(newEvent(w.toplevel, toplevelUnsetMaximized))
	tc.awaitPixel(image.Pt(4, 4), background)
}
