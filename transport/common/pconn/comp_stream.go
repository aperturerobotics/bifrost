package pconn

import (
	"github.com/golang/snappy"
	"github.com/paralin/kcp-go-lite"
)

type compStream struct {
	*kcp.UDPSession
	w *snappy.Writer
	r *snappy.Reader
}

func (c *compStream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *compStream) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	err = c.w.Flush()
	return n, err
}

func newCompStream(conn *kcp.UDPSession) *compStream {
	c := &compStream{UDPSession: conn}
	c.w = snappy.NewBufferedWriter(conn)
	c.r = snappy.NewReader(conn)
	return c
}
