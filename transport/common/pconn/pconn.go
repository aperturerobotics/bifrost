package pconn

import (
	"context"
	"crypto/tls"
	"net"
	"sync"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/libp2p/go-libp2p-core/crypto"
	p2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/lucas-clemente/quic-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Transport implements a bifrost transport with a Quic-based packet conn.
// Transport UUIDs are deterministic and based on the LocalAddr() of the pconn.
type Transport struct {
	// le is the logger
	le *logrus.Entry
	// peerID is the local peer id
	peerID peer.ID
	// privKey is the local private key
	privKey crypto.PrivKey
	// pc is the underlying packet conn.
	pc net.PacketConn
	// uuid is the unique id
	uuid uint64
	// laddr is the local address
	laddr net.Addr
	// handler is the transport handler
	handler transport.TransportHandler
	// opts are extra options
	opts Opts
	// addrParser parses an address from a string
	// if nil, the dialer will not function
	addrParser func(addr string) (net.Addr, error)
	// identity is the p2ptls identity
	identity *p2ptls.Identity
	// ctxCh contains the ctx
	ctxCh chan context.Context

	// linksMtx guards links and dialers
	linksMtx sync.Mutex
	// links is the links map
	// TODO: we can have multiple links for one remote peer on a transport
	// TODO: key this by address instead
	// maps peer.ID to Link
	links map[string]*Link
	// dialers is the dialers map
	// maps address to dialer
	dialers map[string]*dialer
}

// New constructs a new packet-conn based transport.
func New(
	le *logrus.Entry,
	pc net.PacketConn,
	privKey crypto.PrivKey,
	addrParser func(addr string) (net.Addr, error),
	tc transport.TransportHandler,
	opts *Opts,
) (*Transport, error) {
	if opts == nil {
		opts = &Opts{}
	}

	peerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	identity, err := p2ptls.NewIdentity(privKey)
	if err != nil {
		return nil, err
	}

	laddr := pc.LocalAddr()
	return &Transport{
		le:         le.WithField("laddr", laddr.String()),
		pc:         pc,
		handler:    tc,
		identity:   identity,
		opts:       *opts,
		peerID:     peerID,
		laddr:      laddr,
		privKey:    privKey,
		uuid:       newTransportUUID(laddr, peerID),
		links:      make(map[string]*Link),
		dialers:    make(map[string]*dialer),
		addrParser: addrParser,
		ctxCh:      make(chan context.Context, 1),
	}, nil
}

// newTransportUUID builds the UUID for a transport
func newTransportUUID(localAddr net.Addr, peerID peer.ID) uint64 {
	return scrc.Crc64(
		[]byte("bifrost/pconn/"),
		[]byte(localAddr.String()),
		[]byte("/"),
		[]byte(peerID.Pretty()),
	)
}

// defaultQuicConfig constructs the default quic config.
func defaultQuicConfig() *quic.Config {
	return &quic.Config{
		MaxIncomingStreams:                    1000,
		MaxIncomingUniStreams:                 -1,              // disable unidirectional streams
		MaxReceiveStreamFlowControlWindow:     3 * (1 << 20),   // 3 MB
		MaxReceiveConnectionFlowControlWindow: 4.5 * (1 << 20), // 4.5 MB
		AcceptToken: func(clientAddr net.Addr, _ *quic.Token) bool {
			// TODO(#6): require source address validation when under load
			return true
		},
		KeepAlive: true,
	}
}

// LocalAddr returns the local address.
func (t *Transport) LocalAddr() net.Addr {
	return t.laddr
}

// DialPeer dials a peer given an address. The yielded link should be
// emitted to the transport handler. DialPeer should return nil if the link
// was established. DialPeer will then not be called again for the same peer
// ID and address tuple until the yielded link is lost.
func (t *Transport) DialPeer(ctx context.Context, peerID peer.ID, as string) (bool, error) {
	if t.addrParser == nil {
		return true, nil
	}

	addr, err := t.addrParser(as)
	if err != nil {
		return true, err
	}
	if adrs := addr.String(); adrs != as {
		return false, errors.Errorf("addr parser returned %s when input was %s", adrs, as)
	}

	var rctx context.Context
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case rctx = <-t.ctxCh:
		t.ctxCh <- rctx
	}

	// abort if we already have a peer with the same id connected
	var dl *dialer
	t.linksMtx.Lock()
	defer t.linksMtx.Unlock()

	if edl, dialerOk := t.dialers[as]; dialerOk {
		if edl.peerID != peerID {
			// TODO: possibly override the prior
			return false, nil
		}
		dl = edl
	} else {
		if _, elnk := t.links[as]; elnk {
			return false, nil
		}
	}
	if dl == nil {
		dl, err = newDialer(rctx, t, peerID, addr, as)
		if err == nil {
			t.dialers[as] = dl
			go func() {
				lnk, _ := dl.execute()
				if lnk == nil {
					t.linksMtx.Lock()
					if odl, odlOk := t.dialers[as]; odlOk && odl == dl {
						delete(t.dialers, as)
					}
					t.linksMtx.Unlock()
				}
			}()
		}
	}
	if err != nil {
		return false, err
	}
	return true, nil

	// Add a reference to the dialer.
	/*
		nlnk, err := dl.waitForComplete(ctx)
		return nlnk != nil, err
	*/
}

// Execute executes the transport as configured, returning any fatal error.
func (t *Transport) Execute(ctx context.Context) error {
	t.ctxCh <- ctx
	defer func() {
		<-t.ctxCh
	}()
	// Listen
	var tlsConf tls.Config
	tlsConf.GetConfigForClient = func(_ *tls.ClientHelloInfo) (*tls.Config, error) {
		// return a tls.Config that verifies the peer's certificate chain.
		// Note that since we have no way of associating an incoming QUIC connection with
		// the peer ID calculated here, we don't actually receive the peer's public key
		// from the key chan.
		conf, _ := t.identity.ConfigForAny()
		return conf, nil
	}
	quicConfig := defaultQuicConfig()
	t.le.Debug("starting to listen with quic + tls")
	ln, err := quic.Listen(t.pc, &tlsConf, quicConfig)
	if err != nil {
		return err
	}
	defer ln.Close()

	// accept new connections
	for {
		sess, err := ln.Accept(ctx)
		if err != nil {
			return err
		}

		_, err = t.handleSession(ctx, sess)
		if err != nil {
			t.le.WithError(err).Warn("cannot build link for session")
			_ = sess.CloseWithError(500, "cannot build link")
			continue
		}
	}
}

// GetUUID returns a host-unique ID for this transport.
func (t *Transport) GetUUID() uint64 {
	return t.uuid
}

// GetPeerID returns the peer ID.
func (t *Transport) GetPeerID() peer.ID {
	return t.peerID
}

// Close closes the transport, returning any errors closing.
func (t *Transport) Close() error {
	return nil
}

// handleLinkLost is called when a link is lost.
func (u *Transport) handleLinkLost(addrStr string, lnk *Link) {
	u.linksMtx.Lock()
	existing := u.links[addrStr]
	rel := existing == lnk
	if rel {
		delete(u.links, addrStr)
	}
	u.linksMtx.Unlock()

	if u.handler != nil && rel {
		u.handler.HandleLinkLost(lnk)
	}
}

// handleSession handles a new session.
func (t *Transport) handleSession(ctx context.Context, sess quic.Session) (*Link, error) {
	var lnk *Link
	var err error
	as := sess.RemoteAddr().String()
	lnk, err = NewLink(
		ctx,
		t.le,
		&t.opts,
		t.GetUUID(),
		t.peerID,
		t.pc.LocalAddr(),
		sess,
		func() {
			if lnk != nil {
				go t.handleLinkLost(as, lnk)
			}
		},
	)
	if err != nil {
		return nil, err
	}

	t.linksMtx.Lock()
	if dialer, dialerOk := t.dialers[as]; dialerOk {
		// TODO fully cancel dialer
		// TODO push link to dialer (to resolve it)
		dialer.ctxCancel()
		delete(t.dialers, as)
	}
	if elnk, elnkOk := t.links[as]; elnkOk {
		rpeer := elnk.peerID
		t.le.
			WithField("remote-addr", as).
			WithField("remote-peer", rpeer.Pretty()).
			Warn("userping existing session with peer")
		go elnk.Close()
	}
	t.links[as] = lnk
	go t.handler.HandleLinkEstablished(lnk)
	t.linksMtx.Unlock()
	return lnk, nil
}

// _ is a type assertion
var _ transport.Transport = ((*Transport)(nil))
