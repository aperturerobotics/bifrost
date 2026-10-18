package wayland

import (
	"image"
	"testing"

	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
)

// TestScale checks a viewer's scale reaches the clients' outputs and the
// frames: an image of either scale covers the surface's size in screen
// units, and the viewer's pointer pixels arrive in screen units.
func TestScale(t *testing.T) {
	// Report a scale of two, which the output describes.
	tc := newTestCompositor(t, 160, 120)
	c := tc.connect()
	c.roundtrip()
	tc.comp.Input(&s4wave_vm.WasiVmDisplayInput{Body: &s4wave_vm.WasiVmDisplayInput_Scale{
		Scale: &s4wave_vm.WasiVmDisplayScale{Scale: 2},
	}})
	if n := c.await(idOutput, outputScale).int(); n != 2 {
		t.Fatalf("output scale %d, want 2", n)
	}

	// A window drawn at scale one shows twice as large, entering the output.
	w := c.window(16, 16, bufRed)
	c.await(w.surface, surfaceEnter)
	last, past := image.Pt(2*cascadeStep+31, 2*cascadeStep+31), image.Pt(2*cascadeStep+32, 2*cascadeStep+32)
	tc.awaitPixel(last, red)
	if size := tc.screen.Rect.Size(); size != image.Pt(320, 240) {
		t.Fatalf("frame size %v, want 320x240", size)
	}

	// Drawn at scale two, a buffer twice the size covers the same area.
	c.request(newEvent(w.surface, surfaceSetBufferScale).int(2))
	bid := c.buffer(32, 32, [4]byte{bufBlue[0], bufBlue[1], bufBlue[2], 0}, shmFormatXRGB8888)
	c.request(newEvent(w.surface, surfaceAttach).uint(bid).int(0).int(0))
	c.request(newEvent(w.surface, surfaceDamageBuffer).int(0).int(0).int(32).int(32))
	c.request(newEvent(w.surface, surfaceCommit))
	tc.awaitPixel(last, blue)
	tc.awaitPixel(past, background)

	// A press at a pixel arrives at its screen unit on the window.
	pointer := c.newID()
	c.request(newEvent(idSeat, seatGetPointer).uint(pointer))
	c.roundtrip()
	tc.pointer(2*(cascadeStep+5)+1, 2*(cascadeStep+6), 1)
	m := c.await(pointer, pointerEnter)
	if _, s, x, y := m.uint(), m.uint(), m.int(), m.int(); s != w.surface || x != 5*256 || y != 6*256 {
		t.Fatalf("pointer entered %d at %d,%d, want %d at 5,6 in 24.8", s, x, y, w.surface)
	}
}
