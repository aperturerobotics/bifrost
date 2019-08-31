package blockcompress

import (
	"github.com/golang/snappy"
	"io"
)

// SnappyStream implements a snappy compression backed stream.
type SnappyStream struct {
	io.ReadWriteCloser
	w *snappy.Writer
	r *snappy.Reader
}

// NewSnappyStream constructs a new snappy compression stream.
func NewSnappyStream(conn io.ReadWriteCloser) *SnappyStream {
	c := &SnappyStream{ReadWriteCloser: conn}
	c.w = snappy.NewBufferedWriter(conn)
	c.r = snappy.NewReader(conn)
	return c
}

// Read implements io.ReadWriter.
func (c *SnappyStream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

// Write implements io.ReadWriter.
func (c *SnappyStream) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	if err == nil {
		err = c.w.Flush()
	}
	return n, err
}

// Close implements io.ReadWriteCloser
func (c *SnappyStream) Close() error {
	_ = c.w.Close()
	return c.ReadWriteCloser.Close()
}

// _ is a type assertion
var _ io.ReadWriteCloser = ((*SnappyStream)(nil))
