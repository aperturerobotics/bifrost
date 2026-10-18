package dex_solicit

import "net"

// writeObserverConn reports an attempted write before it blocks on a real pipe.
type writeObserverConn struct {
	// Conn is the unbuffered pipe connection.
	net.Conn
	// writing receives a notification when a write reaches the pipe.
	writing chan struct{}
}

// Write reports the attempt without changing the pipe's backpressure.
func (c *writeObserverConn) Write(data []byte) (int, error) {
	// Let the test revoke while the response write is in progress.
	select {
	case c.writing <- struct{}{}:
	default:
	}

	// Preserve the real transport's blocking write.
	return c.Conn.Write(data)
}

// _ is a type assertion.
var _ net.Conn = (*writeObserverConn)(nil)
