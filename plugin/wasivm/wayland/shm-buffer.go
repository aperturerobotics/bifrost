package wayland

import (
	"image"
	"io"

	"github.com/pkg/errors"
)

// shmBuffer is a wl_buffer: a pixel array in a shared memory pool. Both of
// its formats store each pixel as the bytes blue, green, red and alpha or
// padding.
type shmBuffer struct {
	// pool is the pool holding the pixels.
	pool *shmPool
	// id is the buffer's object id.
	id uint32
	// offset is the first pixel's offset in the pool.
	offset int64
	// width and height are the buffer's size in pixels.
	width, height int
	// stride is the distance between rows in bytes.
	stride int
	// destroyed is set once the client destroys the buffer.
	destroyed bool
}

// request handles wl_buffer.destroy.
func (b *shmBuffer) request(op uint16, _ *message) error {
	if op == bufferDestroy {
		b.pool.c.destroy(b.id)
		b.destroy()
	}
	return nil
}

// destroy releases the buffer's hold on its pool.
func (b *shmBuffer) destroy() {
	b.destroyed = true
	b.pool.unref()
}

// read returns the area r of the buffer as opaque RGBA pixels. A pool file
// shorter than its declared size reads as zeros past its end.
func (b *shmBuffer) read(r image.Rectangle) ([]byte, error) {
	// Read the band of rows covering r.
	if r.Empty() {
		return nil, nil
	}
	start := b.offset + int64(r.Min.Y*b.stride+4*r.Min.X)
	band := make([]byte, (r.Dy()-1)*b.stride+4*r.Dx())
	if _, err := b.pool.file.ReadAt(band, start); err != nil && !errors.Is(err, io.EOF) {
		return nil, err
	}

	// Reorder each pixel's bytes and drop its alpha.
	pix := make([]byte, 4*r.Dx()*r.Dy())
	for y := range r.Dy() {
		row := band[y*b.stride:]
		out := pix[4*y*r.Dx():]
		for x := range r.Dx() {
			out[4*x], out[4*x+1], out[4*x+2], out[4*x+3] = row[4*x+2], row[4*x+1], row[4*x], 0xff
		}
	}
	return pix, nil
}

// _ is a type assertion.
var _ object = (*shmBuffer)(nil)
