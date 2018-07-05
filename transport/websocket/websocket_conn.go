//+build !js

package websocket

import (
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// wsConn wraps the websocket conn to make it net.Conn compatible
type wsConn struct {
	*websocket.Conn
}

func newWsConn(conn *websocket.Conn) *wsConn {
	return &wsConn{Conn: conn}
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (w *wsConn) Read(b []byte) (int, error) {
	for {
		mt, r, err := w.NextReader()
		if err != nil {
			return 0, err
		}

		if mt != websocket.BinaryMessage {
			continue
		}

		n, err := r.Read(b)
		if n == 0 && err == nil {
			continue
		}

		return n, err
	}
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (w *wsConn) Write(b []byte) (n int, err error) {
	err = w.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return
	}

	return len(b), nil
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (w *wsConn) SetDeadline(t time.Time) error {
	if err := w.SetReadDeadline(t); err != nil {
		return err
	}

	return w.SetWriteDeadline(t)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (w *wsConn) Close() error {
	return w.Conn.Close()
}

// _ is a type assertion
var _ net.Conn = ((*wsConn)(nil))
