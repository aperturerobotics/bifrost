package wayland

import (
	"bytes"
	"image"
	"image/draw"
	"math/rand/v2"
	"testing"
)

// TestView checks a viewer's copy of the screen matches the screen after
// each frame, through random paints and moves, with frames taken at random.
func TestView(t *testing.T) {
	// Start the screen and the viewer's copy alike.
	screen := image.Rect(0, 0, 24, 24)
	truth, copied := image.NewRGBA(screen), image.NewRGBA(screen)
	v := &view{}
	rng := rand.New(rand.NewPCG(1, 2))
	color := byte(0)
	paint := func(r image.Rectangle) {
		color++
		for y := r.Min.Y; y < r.Max.Y; y++ {
			for x := r.Min.X; x < r.Max.X; x++ {
				truth.Pix[truth.PixOffset(x, y)] = color
			}
		}
		v.add(r)
	}
	randRect := func() image.Rectangle {
		return image.Rect(rng.IntN(24), rng.IntN(24), rng.IntN(24), rng.IntN(24))
	}

	// Paint, move and take frames at random.
	for step := range 20000 {
		switch rng.IntN(4) {
		case 0:
			// Paint an area.
			paint(randRect())
		case 1, 2:
			// Move an area, often the last move's destination, and paint
			// what it uncovers.
			src := randRect()
			if n := len(v.copies); n != 0 && rng.IntN(2) == 0 {
				src = v.copies[n-1].dst()
			}
			by := image.Pt(rng.IntN(9)-4, rng.IntN(9)-4)
			src = src.Intersect(screen).Intersect(screen.Sub(by))
			if src.Empty() || by == (image.Point{}) {
				continue
			}
			from := image.NewRGBA(src)
			draw.Draw(from, src, truth, src.Min, draw.Src)
			draw.Draw(truth, src.Add(by), from, src.Min, draw.Src)
			v.copy(src, by)
			for _, r := range subtract(src, src.Add(by)) {
				paint(r)
			}
		case 3:
			// Take a frame and apply it as a viewer does.
			for _, m := range v.copies {
				from := image.NewRGBA(m.src)
				draw.Draw(from, m.src, copied, m.src.Min, draw.Src)
				draw.Draw(copied, m.dst(), from, m.src.Min, draw.Src)
			}
			for _, r := range v.damage {
				draw.Draw(copied, r, truth, r.Min, draw.Src)
			}
			v.copies, v.damage = v.copies[:0], v.damage[:0]
			if !bytes.Equal(truth.Pix, copied.Pix) {
				t.Fatalf("step %d: the viewer's copy differs from the screen", step)
			}
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
