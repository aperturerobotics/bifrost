package wayland

import (
	"image"
	"testing"
)

// TestRegionCovers checks covers lets the newest rectangle over each point
// decide whether the region holds it.
func TestRegionCovers(t *testing.T) {
	add := func(x0, y0, x1, y1 int) regionOp { return regionOp{rect: image.Rect(x0, y0, x1, y1), add: true} }
	sub := func(x0, y0, x1, y1 int) regionOp { return regionOp{rect: image.Rect(x0, y0, x1, y1)} }
	rect := image.Rect(0, 0, 10, 10)
	for _, tt := range []struct {
		name string
		ops  []regionOp
		want bool
	}{
		{"empty", nil, false},
		{"whole", []regionOp{add(0, 0, 10, 10)}, true},
		{"larger", []regionOp{add(-5, -5, 20, 20)}, true},
		{"two halves", []regionOp{add(0, 0, 5, 10), add(5, 0, 10, 10)}, true},
		{"hole", []regionOp{add(0, 0, 10, 10), sub(4, 4, 5, 5)}, false},
		{"hole filled", []regionOp{add(0, 0, 10, 10), sub(4, 4, 5, 5), add(4, 4, 5, 5)}, true},
		{"hole outside", []regionOp{add(0, 0, 20, 20), sub(12, 12, 14, 14)}, true},
		{"subtracted first", []regionOp{sub(0, 0, 10, 10), add(0, 0, 10, 10)}, true},
		{"short", []regionOp{add(0, 0, 10, 9)}, false},
	} {
		if got := (&region{ops: tt.ops}).covers(rect); got != tt.want {
			t.Errorf("%s: covers = %v, want %v", tt.name, got, tt.want)
		}
	}
}

// TestSubtract checks the parts of a rectangle outside another cover exactly
// the difference.
func TestSubtract(t *testing.T) {
	a := image.Rect(0, 0, 10, 10)
	for _, b := range []image.Rectangle{
		image.Rect(2, 3, 5, 7),
		image.Rect(-5, -5, 4, 20),
		image.Rect(20, 20, 30, 30),
		image.Rect(0, 0, 10, 10),
	} {
		parts := subtract(a, b)
		for y := a.Min.Y; y < a.Max.Y; y++ {
			for x := a.Min.X; x < a.Max.X; x++ {
				pt, n := image.Pt(x, y), 0
				for _, p := range parts {
					if pt.In(p) {
						n++
					}
				}
				if want := map[bool]int{true: 0, false: 1}[pt.In(b)]; n != want {
					t.Fatalf("subtract(%v, %v): %v is in %d parts, want %d", a, b, pt, n, want)
				}
			}
		}
	}
}
