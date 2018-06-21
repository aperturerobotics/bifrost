package udp

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/handshake/identity/s2s"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// UDP implements a UDP transport.
// It is unordered, unreliable, and unencrypted.
type UDP struct {
	// le is the logger
	le *logrus.Entry
	// pc is the packet conn
	pc net.PacketConn
	// privKey is the local priv key
	privKey crypto.PrivKey
	// dialCh is the dial request channel
	dialCh chan net.Addr

	// handshakesMtx guards the handshakes map
	handshakesMtx sync.Mutex
	// handshakes is the set of ongoing handshakes
	handshakes map[string]*inflightHandshake

	// linksMtx guards the links map
	linksMtx sync.Mutex
	// links is the set of active links
	links map[string]*Link
}

// inflightHandshake is an on-going handshake.
type inflightHandshake struct {
	ctxCancel context.CancelFunc
	hs        identity.Handshaker
	addr      net.Addr
}

// NewUDP builds a new UDP transport, listening on the addr.
func NewUDP(le *logrus.Entry, listenAddr string, pKey crypto.PrivKey) (*UDP, error) {
	pc, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	return &UDP{
		le:      le,
		pc:      pc,
		privKey: pKey,
		dialCh:  make(chan net.Addr, 10),

		handshakes: make(map[string]*inflightHandshake),
		links:      make(map[string]*Link),
	}, nil
}

// handleCompleteHandshake handles a completed handshake.
func (u *UDP) handleCompleteHandshake(addr net.Addr, result *identity.Result) {
	as := addr.String()
	pid, _ := peer.IDFromPublicKey(result.Peer)
	le := u.le.
		WithField("remote-id", pid.Pretty()).
		WithField("remote-addr", as).
		WithField("secret-sha", hashData(result.Secret[:]))
	le.Info("handshake complete")

	u.linksMtx.Lock()
	defer u.linksMtx.Unlock()

	// TODO; re-configure link for new secret rather than closing it.
	// TODO: find any peers with
	if l, ok := u.links[as]; ok {
		le.
			WithField("old-secret", hashData(l.sharedSecret[:])).
			Debug("userping old session with peer")
		go l.Close() // new goroutine to avoid deadlock
	}

	u.links[as] = NewLink(u.pc, addr.(*net.UDPAddr), result, result.Secret)
	// HandleLink()
}

// Dial instructs the transport to attempt to handshake with a peer.
// The function may return immediately.
// The handshake will be canceled if ctx is canceled.
func (u *UDP) Dial(ctx context.Context, addr net.Addr) error {
	as := addr.String()

	u.handshakesMtx.Lock()
	defer u.handshakesMtx.Unlock()

	if _, ok := u.handshakes[as]; !ok {
		_, err := u.pushHandshaker(ctx, addr, true)
		return err
	}

	return nil
}

// Execute processes the transport, emitting events to the handler.
// Fatal errors are returned.
func (u *UDP) Execute(ctx context.Context, handler transport.Handler) error {
	buf := make([]byte, 1400)

	laddr := *u.pc.LocalAddr().(*net.UDPAddr)
	laddr.Port = 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := u.pc.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
			return err
		}

		n, addr, err := u.pc.ReadFrom(buf)
		if err != nil {
			if e, ok := err.(net.Error); !ok || (!e.Timeout() || !e.Temporary()) {
				return err
			}

			continue
		}

		as := addr.String()
		u.handshakesMtx.Lock()
		hs := u.handshakes[as]
		ale := u.le.WithField("addr", as)
		if hs == nil {
			ale.Debug("pushing new handshaker")
			hs, err = u.pushHandshaker(ctx, addr, false)
		} else {
			ale.Debug("used existing handshaker")
		}
		u.handshakesMtx.Unlock()
		if err != nil {
			return errors.Wrap(err, "build handshaker")
		}

		// re-use buf
		hs.hs.Handle(buf[:n])
	}
}

// LocalAddr returns the local address.
func (u *UDP) LocalAddr() net.Addr {
	return u.pc.LocalAddr()
}

// Close closes the connection.
func (u *UDP) Close() error {
	return u.pc.Close()
}

// pushHandshaker builds a new handshaker for the address.
// it is expected that handshakesMtx is locked before calling pushHandshaker
func (u *UDP) pushHandshaker(ctx context.Context, addr net.Addr, inititiator bool) (*inflightHandshake, error) {
	as := addr.String()
	nctx, nctxCancel := context.WithCancel(ctx)
	hs := &inflightHandshake{ctxCancel: nctxCancel, addr: addr}
	var err error
	hs.hs, err = s2s.NewHandshaker(
		u.privKey,
		nil,
		func(data []byte) error {
			_, err := u.pc.WriteTo(data, addr)
			return err
		},
		nil,
	)
	if err != nil {
		nctxCancel()
		return nil, err
	}

	u.handshakes[as] = hs
	go u.processHandshake(nctx, hs, inititiator)
	return hs, nil
}

// processHandshake processes an in-flight handshake.
func (u *UDP) processHandshake(ctx context.Context, hs *inflightHandshake, initiator bool) {
	as := hs.addr.String()
	ule := u.le.WithField("addr", as)

	defer func() {
		hs.hs.Close()
		u.handshakesMtx.Lock()
		ohs := u.handshakes[as]
		if ohs == hs {
			delete(u.handshakes, as)
		}
		u.handshakesMtx.Unlock()
	}()

	res, err := hs.hs.Execute(ctx, initiator)
	if err != nil {
		if err == context.Canceled {
			return
		}

		ule.WithError(err).Warn("error handshaking")
		return
	}

	u.handleCompleteHandshake(hs.addr, res)
}

// _ is a type assertion.
var _ transport.Transport = ((*UDP)(nil))
