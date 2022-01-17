package blockcompress

import (
	"io"

	cmp "github.com/pierrec/lz4"
)

// Lz4Stream implements a lz4 compression backed stream.
type Lz4Stream struct {
	io.ReadWriteCloser
	w *cmp.Writer
	r *cmp.Reader
}

// NewLz4Stream constructs a new cmp compression stream.
func NewLz4Stream(conn io.ReadWriteCloser) *Lz4Stream {
	c := &Lz4Stream{ReadWriteCloser: conn}
	c.w = cmp.NewWriter(conn)
	c.r = cmp.NewReader(conn)
	return c
}

// Read implements io.ReadWriteCloser.
func (c *Lz4Stream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

// Write implements io.ReadWriteCloser.
func (c *Lz4Stream) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	if err == nil {
		err = c.w.Flush()
	}
	return n, err
}

// Close implements io.ReadWriteCloser
func (c *Lz4Stream) Close() error {
	_ = c.w.Close()
	return c.ReadWriteCloser.Close()
}

// _ is a type assertion
var _ io.ReadWriteCloser = ((*Lz4Stream)(nil))
