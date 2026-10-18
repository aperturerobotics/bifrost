package wayland

import "image"

// maxScale bounds the output's scale and a surface's buffer scale.
const maxScale = 3

// maxImageSize bounds each side of a surface's image in pixels: a window as
// large as its space, drawn at the largest scale the space shows well.
const maxImageSize = 2 * maxWindowSize

// The compositor places surfaces in screen units, the viewers' CSS pixels,
// and the viewers draw the screen at the output's scale: scale pixels a side
// per unit. A client draws its buffers at the output's scale once its
// surfaces enter the output, and a viewer samples a texture of another scale
// to fit.

// scaleRect returns r with each coordinate multiplied by n.
func scaleRect(r image.Rectangle, n int) image.Rectangle {
	return image.Rectangle{Min: r.Min.Mul(n), Max: r.Max.Mul(n)}
}

// setScale draws the output at n pixels a unit, within one and maxScale. The
// clients' outputs report the new mode and scale. The caller holds the lock.
func (c *Compositor) setScale(n int) {
	// Set the scale.
	n = max(min(n, maxScale), 1)
	if n == c.scale {
		return
	}
	c.scale = n

	// Describe the output again. The viewers draw the screen at the new
	// scale with the textures they have until the clients draw again.
	for cl := range c.clients {
		for _, o := range cl.outputs {
			cl.describeMode(o)
			if o.version >= 2 {
				cl.send(newEvent(o.id, outputDone))
			}
		}
	}
	c.changed = true
}
