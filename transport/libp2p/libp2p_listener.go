package libp2p

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/aperturerobotics/controllerbus/directive"
	lt "github.com/libp2p/go-libp2p-transport"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

// Listener listens on a libp2p transport.
type Listener struct {
	// le is the logger
	le *logrus.Entry
	// tpt is the transport to listen on
	tpt lt.Transport
	// ma is the multiaddr to listen with
	ma ma.Multiaddr
	// uuid is the host-local unique id
	uuid uint64
}

// NewListener constructs a new listener controller.
func NewListener(
	le *logrus.Entry,
	tpt lt.Transport,
	listenAddr ma.Multiaddr,
) *Listener {
	lstr := listenAddr.String()
	uuid := scrc.Crc64([]byte(lstr))
	return &Listener{
		le:   le.WithField("listen-multiaddr", lstr),
		tpt:  tpt,
		ma:   listenAddr,
		uuid: uuid,
	}
}

// GetUUID returns a host-unique ID for this transport.
func (l *Listener) GetUUID() uint64 {
	return l.uuid
}

// GetLinks returns the list of links this transport has active.
func (l *Listener) GetLinks() []link.Link {
	// TODO
	return nil
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (l *Listener) Execute(ctx context.Context) error {
	l.le.Info("listening")
	list, err := l.tpt.Listen(l.ma)
	if err != nil {
		l.le.WithError(err).Error("error listening")
		return err
	}
	defer list.Close()

	// accept loop
	errCh := make(chan error, 1)
	go func() {
		for {
			conn, err := list.Accept()
			if err != nil {
				errCh <- err
				return
			}

			l.handleConn(conn)
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (l *Listener) HandleDirective(inst directive.Instance) (directive.Resolver, error) {
	// TODO
	return nil, nil
}

// handleConn handles a new connection
func (l *Listener) handleConn(c lt.Conn) {
	le := l.le.
		WithField("remote-multiaddr", c.RemoteMultiaddr().String()).
		WithField("remote-peer", c.RemotePeer().Pretty())
	le.Info("connection accepted")
	// TODO
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (l *Listener) Close() error {
	return nil
}

// _ is a type assertion
var _ transport.Transport = ((*Listener)(nil))
