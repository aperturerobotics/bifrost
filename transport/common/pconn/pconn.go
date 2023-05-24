package pconn

import (
	"context"
	"net"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	transport_quic "github.com/aperturerobotics/bifrost/transport/common/quic"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/quic-go/quic-go"
	"github.com/sirupsen/logrus"
)

// Transport implements a bifrost transport with a Quic-based packet conn.
// Transport UUIDs are deterministic and based on the LocalAddr() of the pconn.
type Transport struct {
	// Transport is the underlying quic transport
	*transport_quic.Transport
	// ctx is the root context
	ctx context.Context
	// le is the logger
	le *logrus.Entry
	// peerID is the local peer id
	peerID peer.ID
	// privKey is the local private key
	privKey crypto.PrivKey
	// pc is the underlying packet conn.
	pc net.PacketConn
	// handler is the transport handler
	handler transport.TransportHandler
	// opts are extra options
	opts *Opts
	// addrParser parses an address from a string
	// if nil, the dialer will not function
	addrParser func(addr string) (net.Addr, error)

	// quicConfig is the quic configuration
	quicConfig *quic.Config
	// quicTpt is the quic transport
	quicTpt *quic.Transport
}

// NewTransport constructs a new packet-conn based transport.
func NewTransport(
	ctx context.Context,
	le *logrus.Entry,
	privKey crypto.PrivKey,
	tc transport.TransportHandler,
	opts *Opts,
	// if uuid is 0, generates a uuid based on the local address
	uuid uint64,
	// pc is the packet conn
	pc net.PacketConn,
	// addrParser parses addresses to net.Addr for dialing
	// can be nil
	addrParser func(addr string) (net.Addr, error),
) (*Transport, error) {
	if opts == nil {
		opts = &Opts{}
	}

	peerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	tpt := &Transport{
		ctx:        ctx,
		le:         le,
		pc:         pc,
		handler:    tc,
		opts:       opts,
		peerID:     peerID,
		privKey:    privKey,
		addrParser: addrParser,
	}

	var dialFn transport_quic.DialFunc
	if addrParser != nil {
		dialFn = func(ctx context.Context, addr string) (quic.Connection, net.Addr, error) {
			// parse the addr to a net.Addr
			na, err := addrParser(addr)
			if err != nil {
				return nil, na, err
			}

			// dial via quic
			conn, _, err := transport_quic.DialSessionViaTransport(
				ctx,
				le,
				opts.GetQuic(),
				tpt.quicTpt,
				tpt.Transport.GetIdentity(),
				na,
				"",
			)
			if err != nil {
				return nil, na, err
			}
			return conn, na, nil
		}
	}

	// Build quic transport
	tpt.Transport, err = transport_quic.NewTransport(
		ctx,
		le,
		uuid,
		pc.LocalAddr(),
		privKey,
		tc,
		opts.GetQuic(),
		dialFn,
	)
	if err != nil {
		return nil, err
	}

	tpt.quicConfig = transport_quic.BuildQuicConfig(le, opts.GetQuic())
	tpt.quicTpt = &quic.Transport{Conn: pc}

	return tpt, nil
}

// GetPeerID returns the peer ID.
func (t *Transport) GetPeerID() peer.ID {
	return t.peerID
}

// Execute executes the transport as configured, returning any fatal error.
func (t *Transport) Execute(ctx context.Context) error {
	t.le.
		WithField("local-addr", t.LocalAddr().String()).
		Info("starting to listen with quic + tls")
	// Configure TLS to allow any incoming remote peer.
	tlsConf := transport_quic.BuildIncomingTlsConf(t.Transport.GetIdentity(), "")
	ln, err := t.quicTpt.Listen(tlsConf, t.quicConfig)
	if err != nil {
		return err
	}
	defer t.pc.Close()
	defer ln.Close()

	// accept new connections
	for {
		sess, err := ln.Accept(ctx)
		if err != nil {
			return err
		}

		_, err = t.HandleSession(ctx, sess)
		if err != nil {
			t.le.WithError(err).Warn("cannot build link for session")
			_ = sess.CloseWithError(500, "cannot build link")
			continue
		}
	}
}

// Close closes the transport, returning any errors closing.
func (t *Transport) Close() error {
	return nil
}

// _ is a type assertion
var _ transport.Transport = ((*Transport)(nil))
