package blockcompress

import (
	"io"

	cmp "github.com/klauspost/compress/zstd"
)

// ZStdStream implements a zstd compression backed stream.
type ZStdStream struct {
	io.ReadWriteCloser
	w *cmp.Encoder
	r *cmp.Decoder
}

// NewZStdStream constructs a new cmp compression stream.
func NewZStdStream(conn io.ReadWriteCloser) *ZStdStream {
	c := &ZStdStream{ReadWriteCloser: conn}
	c.w, _ = cmp.NewWriter(conn)
	c.r, _ = cmp.NewReader(conn)
	return c
}

// Read implements io.ReadWriteCloser.
func (c *ZStdStream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

// Write implements io.ReadWriteCloser.
func (c *ZStdStream) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	if err == nil {
		err = c.w.Flush()
	}
	return n, err
}

// Close implements io.ReadWriteCloser
func (c *ZStdStream) Close() error {
	_ = c.w.Close()
	return c.ReadWriteCloser.Close()
}

// _ is a type assertion
var _ io.ReadWriteCloser = ((*ZStdStream)(nil))
