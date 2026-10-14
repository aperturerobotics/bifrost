//go:build !js

package wayland

import (
	"encoding/binary"
	"strings"
	"testing"

	"github.com/s4wave/spacewave/plugin/wasivm/kernel"
	"github.com/s4wave/spacewave/plugin/wasivm/kernel/engine/native"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// testClient speaks the Wayland wire protocol to the compositor as a guest
// client would, through a host connection to its socket.
type testClient struct {
	// t is the test.
	t *testing.T
	// conn is the connection.
	conn *kernel.UnixConn
	// buf is the bytes read and not yet parsed.
	buf []byte
	// files is the files received and not yet taken.
	files []*kernel.File
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

// TestCompositor checks a client binds the globals, maps a toplevel from a
// shared memory buffer that the viewer then shows, receives the keymap file
// and gets pointer input.
func TestCompositor(t *testing.T) {
	// Start a kernel.
	ctx := t.Context()
	rt := native.New(ctx)
	defer func() { _ = rt.Close(ctx) }()
	k, err := kernel.New(ctx, rt, &kernel.Config{VirtualTerminal: true, Stdin: strings.NewReader("")})
	if err != nil {
		t.Fatal(err)
	}

	// Serve a compositor on it.
	comp, err := Listen(k, 64, 48)
	if err != nil {
		t.Fatal(err)
	}
	defer comp.Close()
	go func() { _ = comp.Serve(ctx) }()

	// Collect the viewer's frames.
	frames := make(chan *s4wave_vm.WasiVmDisplayFrame, 16)
	go func() {
		_ = comp.Run(ctx, func(f *s4wave_vm.WasiVmDisplayFrame) error {
			frames <- f
			return nil
		})
	}()

	// Connect and list the globals.
	conn, err := k.DialUnix(ctx, SocketPath)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	c := &testClient{t: t, conn: conn}
	c.request(newEvent(1, displayGetRegistry).uint(2))
	c.request(newEvent(1, displaySync).uint(3))

	// Note each global's name until the sync's callback.
	names := make(map[string]uint32)
	for {
		id, op, m := c.next()
		if id == 3 && op == callbackDone {
			break
		}
		if id == 2 && op == registryGlobal {
			name, iface := m.uint(), m.string()
			names[iface] = name
		}
	}

	// Bind the globals the client uses.
	bind := func(iface string, version, id uint32) {
		c.request(newEvent(2, registryBind).uint(names[iface]).string(iface).uint(version).uint(id))
	}
	bind("wl_compositor", 4, 4)
	bind("wl_shm", 1, 5)
	bind("xdg_wm_base", 1, 6)
	bind("wl_seat", 5, 7)

	// Make a toplevel and wait for its first configure.
	c.request(newEvent(4, compositorCreateSurface).uint(8))
	c.request(newEvent(6, wmBaseGetXdgSurface).uint(9).uint(8))
	c.request(newEvent(9, xdgSurfaceGetToplevel).uint(10))
	c.request(newEvent(8, surfaceCommit))
	m := c.await(10, toplevelConfigure)
	if w, h := m.int(), m.int(); w != 64 || h != 48 {
		t.Fatalf("toplevel configured at %dx%d, want 64x48", w, h)
	}
	c.await(9, xdgSurfaceConfigure)

	// Draw a 16x8 buffer of one color, blue green red padding in memory.
	pixels := make([]byte, 16*8*4)
	for i := 0; i < len(pixels); i += 4 {
		copy(pixels[i:], []byte{0x30, 0x20, 0x10, 0})
	}

	// Share it in a pool, then attach, damage and commit it with a frame
	// callback.
	pool := k.NewMemfd("pool", pixels)
	defer pool.Close()
	c.request(newEvent(5, shmCreatePool).uint(11).int(int32(len(pixels))), pool)
	c.request(newEvent(11, shmPoolCreateBuffer).uint(12).int(0).int(16).int(8).int(64).uint(shmFormatXRGB8888))
	c.request(newEvent(8, surfaceAttach).uint(12).int(0).int(0))
	c.request(newEvent(8, surfaceDamage).int(0).int(0).int(16).int(8))
	c.request(newEvent(8, surfaceFrame).uint(13))
	c.request(newEvent(8, surfaceCommit))

	// The compositor releases the buffer and calls back.
	c.await(12, bufferRelease)
	c.await(13, callbackDone)

	// The viewer sees the buffer at the corner and black beside it.
	for {
		f := <-frames
		if f.GetWidth() != 64 || f.GetHeight() != 48 {
			continue
		}
		if got := f.GetRgba()[:4]; string(got) == "\x10\x20\x30\xff" {
			if beside := f.GetRgba()[4*20 : 4*21]; string(beside) != "\x00\x00\x00\xff" {
				t.Fatalf("pixel beside the buffer = %x, want black", beside)
			}
			break
		}
	}

	// The keyboard gets the keymap file and enters the toplevel.
	c.request(newEvent(7, seatGetKeyboard).uint(14))
	m = c.await(14, keyboardKeymap)
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
	m = c.await(14, keyboardEnter)
	if _, surface := m.uint(), m.uint(); surface != 8 {
		t.Fatalf("keyboard entered %d, want 8", surface)
	}

	// The pointer enters the toplevel and follows the viewer's input.
	c.request(newEvent(7, seatGetPointer).uint(15))
	m = c.await(15, pointerEnter)
	if _, surface := m.uint(), m.uint(); surface != 8 {
		t.Fatalf("pointer entered %d, want 8", surface)
	}

	// Move the pointer and press the left button.
	comp.Input(&s4wave_vm.WasiVmDisplayInput{Body: &s4wave_vm.WasiVmDisplayInput_Pointer{
		Pointer: &s4wave_vm.WasiVmPointerEvent{X: 5, Y: 6, Buttons: 1},
	}})
	m = c.await(15, pointerMotion)
	if _, x, y := m.uint(), m.int(), m.int(); x != 5*256 || y != 6*256 {
		t.Fatalf("motion to %d,%d, want 5,6 in 24.8", x, y)
	}
	m = c.await(15, pointerButton)
	if _, _, button, state := m.uint(), m.uint(), m.uint(), m.uint(); button != 0x110 || state != 1 {
		t.Fatalf("button %#x state %d, want 0x110 pressed", button, state)
	}
}
