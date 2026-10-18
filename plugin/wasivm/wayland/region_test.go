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
