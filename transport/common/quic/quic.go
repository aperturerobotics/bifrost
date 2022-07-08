package transport_quic

import (
	"context"
	"net"
	"sync"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/libp2p/go-libp2p-core/crypto"
	p2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
)

// Transport implements a bifrost transport with a Quic-based packet conn.
// Transport UUIDs are deterministic and based on the LocalAddr() of the pconn.
type Transport struct {
	// ctx is the root context
	ctx context.Context
	// le is the logger
	le *logrus.Entry
	// peerID is the local peer id
	peerID peer.ID
	// privKey is the local private key
	privKey crypto.PrivKey
	// uuid is the unique id
	uuid uint64
	// laddr is the local address
	laddr net.Addr
	// handler is the transport handler
	handler transport.TransportHandler
	// opts are extra options
	opts *Opts
	// identity is the p2ptls identity
	identity *p2ptls.Identity
	// dialFn is the dialer function
	// may be empty
	dialFn DialFunc

	// mtx guards below fields
	mtx sync.Mutex
	// links is the links map
	// maps address string to Link
	links map[string]*Link
	// dialers is the dialers map
	// maps address to dialer
	dialers map[string]*Dialer
	// sessionCounter is incremented when a link is created.
	sessionCounter uint32
}

// DialFunc is a function to dial a peer with a string address.
type DialFunc func(ctx context.Context, addr string) (net.PacketConn, net.Addr, error)

// NewTransport constructs a new quic-backed based transport.
func NewTransport(
	ctx context.Context,
	le *logrus.Entry,
	uuid uint64,
	laddr net.Addr,
	privKey crypto.PrivKey,
	tc transport.TransportHandler,
	opts *Opts,
	dialFn DialFunc,
) (*Transport, error) {
	if opts == nil {
		opts = &Opts{}
	}

	peerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	if uuid == 0 {
		uuid = NewTransportUUID("conn", peerID)
	}

	if laddr == nil {
		laddr = peer.NewNetAddr(peerID)
	}

	identity, err := p2ptls.NewIdentity(privKey)
	if err != nil {
		return nil, err
	}

	return &Transport{
		ctx:      ctx,
		le:       le,
		laddr:    laddr,
		uuid:     uuid,
		opts:     opts,
		handler:  tc,
		identity: identity,
		peerID:   peerID,
		privKey:  privKey,
		dialFn:   dialFn,
		links:    make(map[string]*Link),
		dialers:  make(map[string]*Dialer),
	}, nil
}

// GetUUID returns a host-unique ID for this transport.
func (t *Transport) GetUUID() uint64 {
	return t.uuid
}

// GetPeerID returns the peer ID.
func (t *Transport) GetPeerID() peer.ID {
	return t.peerID
}

// GetIdentity returns the p2ptls identity.
func (t *Transport) GetIdentity() *p2ptls.Identity {
	return t.identity
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
	if t.dialFn == nil {
		return false, ErrDialUnimplemented
	}

	// abort if we already have a peer with the same addr connected
	ok, err := CheckAlreadyConnected(t, as, peerID)
	if ok || err != nil {
		return ok, err
	}

	var dl *Dialer
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if edl, dialerOk := t.dialers[as]; dialerOk {
		if edl.peerID != peerID {
			// TODO: possibly override the prior
			return false, nil
		}
		dl = edl
	}
	if dl == nil {
		dl, err = NewDialer(t.ctx, t, peerID, as)
		if err == nil {
			t.dialers[as] = dl
			// start a separate goroutine to execute the dialer.
			go func() {
				_, dlErr := dl.Execute()
				if dlErr != nil {
					t.le.
						WithError(dlErr).
						WithField("remote-addr", as).
						Warn("error dialing peer")
				}
				t.mtx.Lock()
				if odl, odlOk := t.dialers[as]; odlOk && odl == dl {
					delete(t.dialers, as)
				}
				t.mtx.Unlock()
			}()
		}
	}
	if err != nil {
		return false, err
	}
	// TODO: wait for dialer before returning
	return true, nil
}

// CancelDialer cancels the dialer to a given address.
func (t *Transport) CancelDialer(as string) {
	t.mtx.Lock()
	d, ok := t.dialers[as]
	if ok {
		if d.ctxCancel != nil {
			d.ctxCancel()
		}
		delete(t.dialers, as)
	}
	t.mtx.Unlock()
}

// Execute executes the transport as configured, returning any fatal error.
func (t *Transport) Execute(ctx context.Context) error {
	// note: usually implemented in a higher level layer: pconn or conn
	return nil
}

// HandleConn handles an incoming or outgoing packet connection.
//
// dial indicates if this is the originator (outgoing) conn or not
// ctx is used for the negotiation phase only
// if peerID is empty, allows any peer ID on the other end
// raddr can be nil if peerID is NOT empty
func (t *Transport) HandleConn(ctx context.Context, dial bool, pc net.PacketConn, raddr net.Addr, peerID peer.ID) (*Link, error) {
	if raddr == nil {
		if len(peerID) == 0 {
			return nil, ErrRemoteUnspecified
		}
		raddr = peer.NewNetAddr(peerID)
	}

	var sess quic.Connection
	var err error

	t.le.Debugf("negotiating quic session with: %s", raddr.String())
	if dial {
		sess, _, err = DialSession(
			ctx,
			t.le,
			t.opts,
			pc,
			t.identity,
			raddr,
			peerID,
		)
	} else {
		sess, err = ListenSession(
			ctx,
			t.le,
			t.opts,
			pc,
			t.identity,
			peerID,
		)
	}
	if err != nil {
		return nil, err
	}

	lnk, err := t.HandleSession(ctx, sess)
	if err != nil {
		t.le.WithError(err).Warn("cannot build link for session")
		_ = sess.CloseWithError(500, "cannot build link")
		return nil, err
	}

	return lnk, nil
}

// HandleSession handles a new Quic session, creating & registering a link.
func (t *Transport) HandleSession(ctx context.Context, sess quic.Connection) (*Link, error) {
	t.mtx.Lock()
	sessID := t.sessionCounter
	t.sessionCounter++
	t.mtx.Unlock()
	var lnk *Link
	var err error
	as := sess.RemoteAddr().String()
	lnk, err = NewLink(
		t.ctx,
		t.le.WithField("session-id", sessID),
		t.opts,
		t.uuid,
		t.peerID,
		t.laddr,
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

	t.mtx.Lock()
	// Clear any ongoing dial attempt for this addr.
	raddr := lnk.RemoteAddr()
	if raddr != nil {
		rs := raddr.String()
		delete(t.dialers, rs)
	}
	if elnk, elnkOk := t.links[as]; elnkOk {
		rpeer := elnk.GetRemotePeer()
		t.le.
			WithField("remote-addr", as).
			WithField("remote-peer", rpeer.Pretty()).
			Warn("userping existing session with peer")
		go elnk.Close()
	}
	t.links[as] = lnk
	go t.handler.HandleLinkEstablished(lnk)
	t.mtx.Unlock()
	return lnk, nil
}

// LookupLinkWithAddr returns any link with the given remote addr.
func (t *Transport) LookupLinkWithAddr(as string) (*Link, bool) {
	t.mtx.Lock()
	lnk, ok := t.links[as]
	t.mtx.Unlock()
	return lnk, ok
}

// LookupLinkWithPeer returns any link with the given remote peer.
func (t *Transport) LookupLinkWithPeer(p peer.ID) (*Link, bool) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	for _, lnk := range t.links {
		if lnk.GetRemotePeer() == p {
			return lnk, true
		}
	}
	return nil, false
}

// Close closes the transport, returning any errors closing.
func (t *Transport) Close() error {
	return nil
}

// handleLinkLost is called when a link is lost.
func (u *Transport) handleLinkLost(addrStr string, lnk *Link) {
	u.mtx.Lock()
	existing := u.links[addrStr]
	rel := existing == lnk
	if rel {
		delete(u.links, addrStr)
	}
	u.mtx.Unlock()

	if u.handler != nil && rel {
		u.handler.HandleLinkLost(lnk)
	}
}

// _ is a type assertion
var _ transport.Transport = ((*Transport)(nil))
