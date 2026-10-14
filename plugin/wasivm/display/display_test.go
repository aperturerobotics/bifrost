package display

import (
	"testing"

	"github.com/jezek/xgb/xproto"
)

// TestClip checks a region is cut to the screen.
func TestClip(t *testing.T) {
	for _, tc := range []struct {
		r, want xproto.Rectangle
	}{
		{xproto.Rectangle{X: 10, Y: 10, Width: 20, Height: 20}, xproto.Rectangle{X: 10, Y: 10, Width: 20, Height: 20}},
		{xproto.Rectangle{X: -5, Y: 90, Width: 20, Height: 20}, xproto.Rectangle{X: 0, Y: 90, Width: 15, Height: 10}},
		{xproto.Rectangle{X: 100, Y: 0, Width: 20, Height: 20}, xproto.Rectangle{}},
	} {
		if got := clip(tc.r, 100, 100); got != tc.want {
			t.Errorf("clip(%v) = %v, want %v", tc.r, got, tc.want)
		}
	}
}
