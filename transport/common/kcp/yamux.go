package kcp

import (
	"context"

	"github.com/aperturerobotics/bifrost/stream"
	yamux "github.com/libp2p/go-yamux/v4"
)

// yamuxWrapper wraps yamux.
type yamuxWrapper struct {
	*yamux.Session
}

// OpenStream opens a stream.
func (m *yamuxWrapper) OpenStream(ctx context.Context) (stream.Stream, error) {
	return m.Session.OpenStream(ctx)
}

// AcceptStream accepts a stream.
func (m *yamuxWrapper) AcceptStream() (stream.Stream, error) {
	return m.Session.AcceptStream()
}

// Close closes the muxer.
func (m *yamuxWrapper) Close() error {
	return m.Session.Close()
}

// _ is a type assertion
var _ streamMuxer = ((*yamuxWrapper)(nil))
