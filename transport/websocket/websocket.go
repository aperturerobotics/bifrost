package websocket

import (
	"context"
	"net"
	"net/http"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	transport_quic "github.com/aperturerobotics/bifrost/transport/common/quic"
	"github.com/aperturerobotics/bifrost/util/saddr"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	websocket "nhooyr.io/websocket"
)

// TransportType is the transport type identifier for this transport.
const TransportType = "ws"

// ControllerID is the WebSocket controller ID.
const ControllerID = "bifrost/websocket"

// Version is the version of the implementation.
var Version = semver.MustParse("0.0.1")

// WebSocket implements a WebSocket transport.
type WebSocket struct {
	// Transport implements the quic-backed transport type
	*transport_quic.Transport
	// ctx is the context
	ctx context.Context
	// le is the logger
	le *logrus.Entry
	// conf is the websocket config
	conf *Config
	// restrictPeerID restricts incoming conns to the peer ID given
	// usually empty
	restrictPeerID peer.ID
}

// NewWebSocket builds a new WebSocket transport.
//
// ServeHTTP is implemented and can be used with a standard HTTP mux.
// Optionally listens on an address.
func NewWebSocket(
	ctx context.Context,
	le *logrus.Entry,
	conf *Config,
	pKey crypto.PrivKey,
	c transport.TransportHandler,
) (*WebSocket, error) {
	restrictPeerID, err := conf.ParseRestrictPeerID()
	if err != nil {
		return nil, err
	}

	peerID, err := peer.IDFromPrivateKey(pKey)
	if err != nil {
		return nil, err
	}

	var dialFn transport_quic.DialFunc = func(dctx context.Context, addr string) (net.PacketConn, net.Addr, error) {
		conn, _, err := websocket.Dial(dctx, addr, &websocket.DialOptions{
			// Negotiate the bifrost quic sub-protocol ID.
			Subprotocols: []string{transport_quic.Alpn},
		})
		if err != nil {
			return nil, nil, err
		}
		laddr := peer.NewNetAddr(peerID)
		raddr := saddr.NewStringAddr("ws", addr)
		pc := NewPacketConn(ctx, conn, laddr, raddr)
		return pc, raddr, nil
	}

	quicOpts := conf.GetQuic()
	if quicOpts == nil {
		quicOpts = &transport_quic.Opts{}
	} else {
		quicOpts = proto.Clone(quicOpts).(*transport_quic.Opts)
	}

	// set websocket-specific quic opts
	quicOpts.DisableDatagrams = true
	quicOpts.DisableKeepAlive = false
	quicOpts.DisablePathMtuDiscovery = true

	// quicOpts.Verbose = true
	// quicOpts.MaxIdleTimeoutDur = "30s"

	tconn, err := transport_quic.NewTransport(
		ctx,
		le,
		0,
		nil,
		pKey,
		c,
		quicOpts,
		dialFn,
	)
	if err != nil {
		return nil, err
	}
	return &WebSocket{
		Transport:      tconn,
		ctx:            ctx,
		le:             le,
		conf:           conf,
		restrictPeerID: restrictPeerID,
	}, nil
}

// MatchTransportType checks if the given transport type ID matches this transport.
// If returns true, the transport controller will call DialPeer with that tptaddr.
// E.x.: "udp-quic" or "ws"
func (w *WebSocket) MatchTransportType(transportType string) bool {
	return transportType == TransportType
}

// Execute executes the transport as configured, returning any fatal error.
func (w *WebSocket) Execute(ctx context.Context) error {
	// note: w.Transport.Execute is unnecessary (no-op for quic)
	listenAddr := w.conf.GetListenAddr()
	if listenAddr == "" {
		return nil
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- w.ListenHTTP(ctx, listenAddr)
	}()

	select {
	case <-w.ctx.Done():
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}

	return nil
}

// ListenHTTP listens for incoming HTTP connections on an address.
func (w *WebSocket) ListenHTTP(ctx context.Context, addr string) error {
	w.le.Debugf("listening for http/ws on address: %s", addr)
	server := &http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Addr:    addr,
		Handler: w,
	}
	err := http.ListenAndServe(addr, w)
	if serr := server.Shutdown(ctx); serr != nil && serr != context.Canceled {
		w.le.WithError(serr).Warn("graceful shutdown failed")
	}
	_ = server.Close()
	return err
}

// ServeHTTP serves the websocket upgraded HTTP endpoint.
func (w *WebSocket) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c, err := websocket.Accept(rw, req, &websocket.AcceptOptions{
		// Negotiate the bifrost quic sub-protocol ID.
		Subprotocols: []string{transport_quic.Alpn},
		// We are trusting the Quic handshake, not the Websocket handshake.
		InsecureSkipVerify: true,
		// We compress on the quic layer instead.
		CompressionMode: websocket.CompressionDisabled,
	})
	if err != nil {
		w.le.WithError(err).Debug("failed to handle http request")
		return
	}

	raddr := saddr.NewStringAddr("ws", req.RemoteAddr)
	pc := NewPacketConn(req.Context(), c, w.Transport.LocalAddr(), raddr)
	lnk, err := w.Transport.HandleConn(w.ctx, false, pc, raddr, "")
	if err != nil {
		w.le.WithError(err).Warn("unable to handle websocket conn")
		c.Close(websocket.StatusInternalError, "unable to handle connection")
		return
	}

	// hold the HTTP request open until something closes.
	select {
	case <-w.ctx.Done():
	case <-lnk.GetContext().Done():
	case <-req.Context().Done():
	}
	_ = pc.Close()
	_ = lnk.Close()
}

// _ is a type assertion.
var (
	_ transport.Transport       = ((*WebSocket)(nil))
	_ transport.TransportDialer = ((*WebSocket)(nil))
	_ http.Handler              = ((*WebSocket)(nil))
)
