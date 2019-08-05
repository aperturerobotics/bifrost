package pconn

import (
	"github.com/paralin/kcp-go-lite"
	"github.com/pierrec/lz4"
)

type lz4Stream struct {
	*kcp.UDPSession
	w *lz4.Writer
	r *lz4.Reader
}

func (c *lz4Stream) Read(p []byte) (n int, err error) {
	return c.r.Read(p)
}

func (c *lz4Stream) Write(p []byte) (n int, err error) {
	n, err = c.w.Write(p)
	if err == nil {
		err = c.w.Flush()
	}
	return n, err
}

func newLz4Stream(conn *kcp.UDPSession) *lz4Stream {
	c := &lz4Stream{UDPSession: conn}
	c.w = lz4.NewWriter(conn)
	c.r = lz4.NewReader(conn)
	return c
}
