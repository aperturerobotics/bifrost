package kernel

import (
	"context"
	"net"
	"path"

	"github.com/pkg/errors"
)

// ListenUnix binds a Unix stream socket for host code at the absolute path in
// the guest's filesystem, creating its directories, as a guest's bind and
// listen do. A socket node left there by a closed listener is replaced. Each
// guest connect arrives from Accept.
func (k *Kernel) ListenUnix(socketPath string) (*UnixListener, error) {
	// Find the directory and a name that is free or a socket nobody
	// listens on.
	resolved := path.Clean("/" + socketPath)
	dir, name := splitPath(resolved)
	k.mu.Lock()
	defer k.mu.Unlock()
	parent := k.vfs.mkDirAll(dir)
	if parent.mode&sIFMT != sIFDIR {
		return nil, errors.Errorf("listen on %s: parent is not a directory", resolved)
	}
	k.vfs.refreshDirChildren(parent)
	if old := parent.children[name]; old != nil {
		if old.mode&sIFMT != sIFSOCK || old.socket != nil {
			return nil, errors.Errorf("listen on %s: address in use", resolved)
		}
		delete(parent.children, name)
	}

	// Bind and listen, waking the host dials.
	sock := &unixSocket{
		ino:      k.vfs.allocIno(),
		path:     resolved,
		listener: &unixListener{waitCh: make(chan struct{})},
	}
	sock.node = k.vfs.mkNode(parent, name, sIFSOCK|0o777, 0)
	sock.node.socket = sock
	k.listens.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		broadcast()
	})
	ctx, cancel := context.WithCancel(context.Background())
	return &UnixListener{k: k, sock: sock, ctx: ctx, cancel: cancel}, nil
}

// UnixListener is host code's listening Unix socket in the guest's
// filesystem.
type UnixListener struct {
	// k is the kernel whose lock guards the socket.
	k *Kernel
	// sock is the listening socket.
	sock *unixSocket
	// ctx ends when the listener closes, which ends a waiting accept.
	ctx    context.Context
	cancel context.CancelFunc
}

// Accept waits for a guest to connect and returns the host's end.
func (l *UnixListener) Accept() (net.Conn, error) {
	return l.AcceptUnix()
}

// AcceptUnix waits for a guest to connect and returns the host's end. It
// fails with net.ErrClosed once the listener closes.
func (l *UnixListener) AcceptUnix() (*UnixConn, error) {
	// Wait for a connection.
	q := l.sock.listener
	l.k.mu.Lock()
	for len(q.queue) == 0 {
		waitCh := q.waitCh
		l.k.mu.Unlock()
		select {
		case <-l.ctx.Done():
			return nil, net.ErrClosed
		case <-waitCh:
		}
		l.k.mu.Lock()
	}

	// Take it.
	of := q.queue[0]
	q.queue = q.queue[1:]
	l.k.mu.Unlock()
	return l.k.newUnixConn(of, l.sock.path), nil
}

// Close stops listening: later connects are refused, the connections not yet
// accepted close, and a waiting Accept returns.
func (l *UnixListener) Close() error {
	// End the waits, then close the socket under the kernel lock.
	l.cancel()
	l.k.mu.Lock()
	l.sock.close()
	l.k.mu.Unlock()
	return nil
}

// Addr returns the socket's path.
func (l *UnixListener) Addr() net.Addr {
	return &net.UnixAddr{Name: l.sock.path, Net: "unix"}
}

// _ is a type assertion
var _ net.Listener = (*UnixListener)(nil)
