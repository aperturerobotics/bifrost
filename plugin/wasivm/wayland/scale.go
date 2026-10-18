package wayland

import "image"

// maxScale bounds the output's scale and a surface's buffer scale.
const maxScale = 3

// maxImageSize bounds each side of a surface's image in pixels: a window as
// large as its space, drawn at the largest scale the space shows well.
const maxImageSize = 2 * maxWindowSize

// The compositor places surfaces in screen units, the viewers' CSS pixels,
// and draws its frames at the output's scale: scale pixels a side per unit.
// A client draws its buffers at the output's scale once its surfaces enter
// the output, and the compositor samples a buffer of another scale to fit.

// scaleRect returns r with each coordinate multiplied by n.
func scaleRect(r image.Rectangle, n int) image.Rectangle {
	return image.Rectangle{Min: r.Min.Mul(n), Max: r.Max.Mul(n)}
}

// unscaleRect returns the smallest area of screen units covering the pixel
// area r, drawn at n pixels a unit, with r's corners not negative.
func unscaleRect(r image.Rectangle, n int) image.Rectangle {
	return image.Rect(r.Min.X/n, r.Min.Y/n, (r.Max.X+n-1)/n, (r.Max.Y+n-1)/n)
}

// setScale draws the output at n pixels a unit, within one and maxScale. The
// clients' outputs report the new mode and scale, and each viewer draws every
// space again at the new size. The caller holds the lock.
func (c *Compositor) setScale(n int) {
	// Set the scale.
	n = max(min(n, maxScale), 1)
	if n == c.scale {
		return
	}
	c.scale = n

	// Describe the output again, then redraw the spaces.
	for cl := range c.clients {
		for _, o := range cl.outputs {
			cl.describeMode(o)
			if o.version >= 2 {
				cl.send(newEvent(o.id, outputDone))
			}
		}
	}
	c.redrawSpace(c.desktop)
	for _, w := range c.apart {
		c.redrawSpace(w.xdg.toplevel.own)
	}
}

// redrawSpace has each viewer of sp drop what it lacked and draw the whole
// space again. The caller holds the lock.
func (c *Compositor) redrawSpace(sp *space) {
	for v := range sp.viewers {
		v.copies, v.damage = v.copies[:0], v.damage[:0]
	}
	c.damage(sp, sp.area)
}
