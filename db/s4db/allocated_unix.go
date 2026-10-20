//go:build darwin || linux

package s4db

import (
	"golang.org/x/sys/unix"
)

// allocated returns the bytes in [lo, hi) the file system holds data for,
// walking the data and hole regions.
func (f *osFile) allocated(lo, hi int64) (int64, bool) {
	fd := int(f.Fd()) // #nosec G115 -- descriptors fit in int.
	var n int64
	for off := lo; off < hi; {
		// Find the next data region, stopping at the end of the file.
		d, err := unix.Seek(fd, off, unix.SEEK_DATA)
		if err == unix.ENXIO || (err == nil && d >= hi) {
			break
		}
		if err != nil {
			return 0, false
		}

		// Count up to the hole that ends it.
		h, err := unix.Seek(fd, d, unix.SEEK_HOLE)
		if err != nil {
			return 0, false
		}
		n += min(h, hi) - d
		off = h
	}
	return n, true
}
