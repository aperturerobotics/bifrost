package blockcompress

import (
	snappy "github.com/klauspost/compress/s2"
	"io"
)

// S2Stream implements a s2 compression backed stream.
type S2Stream struct {
	io.ReadWriteCloser
	w *snappy.Writer
	r *snappy.Reader
}

// NewS2Stream constructs a new s2 compression stream.
func NewS2Stream(conn io.ReadWriteCloser) *S2Stream {
	c := &S2Stream{ReadWriteCloser: conn}
	c.w = snappy.NewWriter(conn) // snappy.WriterBetterCompression() - slower
	c.r = snappy.NewReader(conn)
	return c
}

// Read implements io.ReadWriteCloser.
func (c *S2Stream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

// Write implements io.ReadWriteCloser.
func (c *S2Stream) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	if err == nil {
		err = c.w.Flush()
	}
	return n, err
}

// Close implements io.ReadWriteCloser
func (c *S2Stream) Close() error {
	_ = c.w.Close()
	return c.ReadWriteCloser.Close()
}

// _ is a type assertion
var _ io.ReadWriteCloser = ((*S2Stream)(nil))
