package kernel

import (
	"context"
	"io"
	"net"
	"path"
	"time"

	"github.com/pkg/errors"
)

var (
	// ErrNoListener is returned by DialUnix when no socket listens at the
	// path.
	ErrNoListener = errors.New("no socket listens at the path")

	// errNoDeadline is returned by the deadline methods of a host connection.
	errNoDeadline = errors.New("unix connection deadlines are unsupported")
)

// DialUnix connects host code to the Unix stream socket listening at the
// absolute path in the guest's filesystem, as a guest's connect does. The
// guest accepts the connection like any other.
func (k *Kernel) DialUnix(socketPath string) (net.Conn, error) {
	// Queue a connection to the listener.
	k.mu.Lock()
	of := k.newSocketFile(0)
	errno := k.connectUnix(of, path.Clean("/"+socketPath))
	k.mu.Unlock()
	if errno != 0 {
		return nil, errors.Wrap(ErrNoListener, socketPath)
	}

	// Wrap the host's end.
	ctx, cancel := context.WithCancel(context.Background())
	return &unixConn{of: of, path: socketPath, ctx: ctx, cancel: cancel}, nil
}

// unixConn is the host's end of a connection to a guest Unix socket. Its open
// file is never installed in a process, so only its pipes are shared.
type unixConn struct {
	// of holds the pipes the connection reads and writes.
	of *OpenFile
	// path is the address of the guest's socket.
	path string
	// ctx ends when the connection closes, which ends a waiting read or
	// write.
	ctx    context.Context
	cancel context.CancelFunc
}

// Read reads from the guest, waiting for data. It returns io.EOF once the
// guest closes its end.
func (c *unixConn) Read(b []byte) (int, error) {
	n, eof := c.of.rx.read(c.ctx, b)
	switch {
	case n != 0:
		return n, nil
	case eof:
		return 0, io.EOF
	case c.ctx.Err() != nil:
		return 0, net.ErrClosed
	}
	return 0, nil
}

// Write writes to the guest, waiting for room. It returns io.ErrClosedPipe
// once the guest closes its end.
func (c *unixConn) Write(b []byte) (int, error) {
	n, ok := c.of.tx.write(c.ctx, b)
	switch {
	case !ok:
		return 0, io.ErrClosedPipe
	case n < len(b):
		return n, net.ErrClosed
	}
	return n, nil
}

// Close closes both directions, which the guest reads as end of file, and
// ends a waiting read or write.
func (c *unixConn) Close() error {
	c.cancel()
	c.of.release()
	return nil
}

// LocalAddr returns the host's end, which has no address.
func (c *unixConn) LocalAddr() net.Addr {
	return &net.UnixAddr{Net: "unix"}
}

// RemoteAddr returns the guest socket's path.
func (c *unixConn) RemoteAddr() net.Addr {
	return &net.UnixAddr{Name: c.path, Net: "unix"}
}

// SetDeadline is unsupported; Close ends a waiting read or write.
func (c *unixConn) SetDeadline(time.Time) error {
	return errNoDeadline
}

// SetReadDeadline is unsupported.
func (c *unixConn) SetReadDeadline(time.Time) error {
	return errNoDeadline
}

// SetWriteDeadline is unsupported.
func (c *unixConn) SetWriteDeadline(time.Time) error {
	return errNoDeadline
}

// _ is a type assertion
var _ net.Conn = (*unixConn)(nil)
