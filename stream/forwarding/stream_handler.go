package stream_forwarding

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/stream/proxy"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr-net"
	"github.com/sirupsen/logrus"
)

// MountedStreamHandler implements the mounted stream handler.
type MountedStreamHandler struct {
	// le is the logger entry
	le *logrus.Entry
	// dialMa is the multiaddr to dial
	dialMa ma.Multiaddr
}

// NewMountedStreamHandler constructs the mounted stream handler.
func NewMountedStreamHandler(le *logrus.Entry, dialMa ma.Multiaddr) (*MountedStreamHandler, error) {
	le = le.WithField("dial-multiaddr", dialMa.String())
	return &MountedStreamHandler{le: le, dialMa: dialMa}, nil
}

// HandleMountedStream handles an incoming mounted stream.
// Any returned error indicates the stream should be closed.
// This function should return as soon as possible, and start
// additional goroutines to manage the lifecycle of the stream.
func (m *MountedStreamHandler) HandleMountedStream(
	ctx context.Context,
	strm link.MountedStream,
) error {
	go func() {
		// Attempt to dial the target.
		// TODO: use context here
		m.le.Debug("dialing to forward stream")
		s := strm.GetStream()
		conn, err := manet.Dial(m.dialMa)
		if err != nil {
			s.Close()
			m.le.WithError(err).Warn("unable to dial to forward stream")
			return
		}

		m.le.Debug("connection opened")
		proxy.ProxyStreams(conn, s)
	}()
	return nil
}

var _ link.MountedStreamHandler = ((*MountedStreamHandler)(nil))
