package pconn

import (
	"github.com/golang/snappy"
	"github.com/paralin/kcp-go-lite"
)

type snappyStream struct {
	*kcp.UDPSession
	w *snappy.Writer
	r *snappy.Reader
}

func (c *snappyStream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *snappyStream) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	err = c.w.Flush()
	return n, err
}

func newSnappyStream(conn *kcp.UDPSession) *snappyStream {
	c := &snappyStream{UDPSession: conn}
	c.w = snappy.NewBufferedWriter(conn)
	c.r = snappy.NewReader(conn)
	return c
}
