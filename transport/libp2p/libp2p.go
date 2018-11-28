package libp2p

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-crypto"
	lt "github.com/libp2p/go-libp2p-transport"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

// BaseTransportID is the base transport identifier
const BaseTransportID = "libp2p"

// GetTransportID returns the transport identifier for a listen ma.
func GetTransportID(listenMultiaddr ma.Multiaddr) string {
	return BaseTransportID + "/" + listenMultiaddr.String()
}

// Version is the version of the udp implementation.
var Version = semver.MustParse("0.0.1")

// LibP2P listens on a libp2p transport.
// NOTE: This transport is not code-complete yet.
type LibP2P struct {
	// le is the logger
	le *logrus.Entry
	// tpt is the transport to listen on
	tpt lt.Transport
	// ma is the multiaddr to listen with
	ma ma.Multiaddr
	// uuid is the host-local unique id
	uuid uint64
	// PeerID is the peer id
	PeerID peer.ID
}

// NewLibP2P constructs a new listener controller.
func NewLibP2P(
	le *logrus.Entry,
	tpt lt.Transport,
	privKey crypto.PrivKey,
	listenAddr ma.Multiaddr,
) *LibP2P {
	lstr := listenAddr.String()
	uuid := scrc.Crc64([]byte(lstr))
	pid, _ := peer.IDFromPrivateKey(privKey)
	return &LibP2P{
		le:     le.WithField("listen-multiaddr", lstr),
		tpt:    tpt,
		ma:     listenAddr,
		uuid:   uuid,
		PeerID: pid,
	}
}

// GetPeerID returns the node id.
func (l *LibP2P) GetPeerID() peer.ID {
	return l.PeerID
}

// GetControllerInfo returns information about the controller.
func (l *LibP2P) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		"bifrost/transport/libp2p/"+l.ma.String()+"/"+Version.String(),
		Version,
		"libp2p listener",
	)
}

// GetUUID returns a host-unique ID for this transport.
func (l *LibP2P) GetUUID() uint64 {
	return l.uuid
}

// Execute executes the given controller.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (l *LibP2P) Execute(ctx context.Context) error {
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
func (c *LibP2P) HandleDirective(ctx context.Context, di directive.Instance) (directive.Resolver, error) {
	// TODO
	return nil, nil
}

// handleConn handles a new connection
func (l *LibP2P) handleConn(c lt.Conn) {
	le := l.le.
		WithField("remote-multiaddr", c.RemoteMultiaddr().String()).
		WithField("remote-peer", c.RemotePeer().Pretty())
	le.Info("connection accepted")
	remoteTransportUUID := scrc.Crc64([]byte(c.RemoteMultiaddr().String()))
	// TODO
	_ = remoteTransportUUID
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (l *LibP2P) Close() error {
	return nil
}

// _ is a type assertion
var _ transport.Transport = ((*LibP2P)(nil))

// _ is a type assertion
var _ controller.Controller = ((*LibP2P)(nil))
