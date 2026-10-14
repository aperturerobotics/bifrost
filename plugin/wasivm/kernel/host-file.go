package kernel

import (
	"io"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

// File is host code's reference to an open file passed over a Unix socket.
// It keeps the file open until Close, like a guest descriptor.
type File struct {
	// k is the kernel whose lock guards the file.
	k *Kernel
	// of is the open file, or nil once closed.
	of *OpenFile
}

// Close drops the reference.
func (f *File) Close() error {
	f.k.mu.Lock()
	defer f.k.mu.Unlock()
	if f.of != nil {
		f.of.unref()
		f.of = nil
	}
	return nil
}

// ReadAt reads the file at off as the guest's processes see it, including
// the writes in their writable shared mappings that no msync has written
// back. It returns io.EOF when it reads less than len(b).
func (f *File) ReadAt(b []byte, off int64) (int, error) {
	// Read under the kernel lock, which keeps the mappings' memory alive.
	if off < 0 {
		return 0, errors.New("read: negative offset")
	}
	f.k.mu.Lock()
	if f.of == nil {
		f.k.mu.Unlock()
		return 0, os.ErrClosed
	}
	data, errno := f.k.readMapped(f.of, off, len(b))
	f.k.mu.Unlock()
	if errno != 0 {
		return 0, errors.New("read: errno " + strconv.Itoa(-int(int32(errno))))
	}

	// Report a short read.
	n := copy(b, data)
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}
