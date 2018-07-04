package pconn

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/handshake/identity/s2s"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// handshakeTimeout is the time after which a handshake expires
var handshakeTimeout = time.Second * 8

// Transport is a net.PacketConn based transport.
// The remote address string is used as an identifying key for sessions.
// It uses KCP to upgrade remote connections to reliable streams.
type Transport struct {
	ctx context.Context
	// le is the logger
	le *logrus.Entry
	// pc is the underlying packet conn.
	pc net.PacketConn
	// uuid is the unique id
	uuid uint64
	// privKey is the local priv key
	privKey crypto.PrivKey

	// handshakesMtx guards the handshakes map
	handshakesMtx sync.Mutex
	// handshakes is the set of ongoing handshakes
	handshakes map[string]*inflightHandshake

	// linksMtx guards the links map
	linksMtx sync.Mutex
	// links is the set of active links
	links map[string]*Link
	// lastLink was the last link to receive a packet
	lastLink *Link
	// lastLinkAddr was the last addr to receive a packet from
	lastLinkAddr string
}

// New builds a new packet-conn and KCP based transport, listening on the addr.
func New(le *logrus.Entry, pc net.PacketConn, pKey crypto.PrivKey) *Transport {
	uuid := scrc.Crc64([]byte(pc.LocalAddr().String()))
	return &Transport{
		le:      le,
		pc:      pc,
		privKey: pKey,
		uuid:    uuid,

		handshakes: make(map[string]*inflightHandshake),
		links:      make(map[string]*Link),
	}
}

// handleCompleteHandshake handles a completed handshake.
func (u *Transport) handleCompleteHandshake(
	addr net.Addr,
	result *identity.Result,
	initiator bool,
) {
	ctx := u.ctx
	as := addr.String()
	pid, _ := peer.IDFromPublicKey(result.Peer)
	le := u.le.
		WithField("remote-id", pid.Pretty()).
		WithField("remote-addr", as)
	le.Info("handshake complete")

	u.linksMtx.Lock()
	defer u.linksMtx.Unlock()

	// TODO; re-configure link for new secret rather than closing it.
	// TODO: find any peers with this ID and userp
	if l, ok := u.links[as]; ok {
		le.
			Debug("userping old session with peer")
		l.Close()
	}

	le.Debug("registering link")
	u.links[as] = NewLink(
		ctx,
		u.pc.LocalAddr(),
		addr,
		u.GetUUID(),
		result,
		result.Secret,
		u.pc.WriteTo,
		initiator,
	)
	// HandleLink()
}

// GetUUID returns a host-unique ID for this transport.
func (u *Transport) GetUUID() uint64 {
	return u.uuid
}

// Dial instructs the transport to attempt to handshake with a peer.
// The function may return immediately.
// The handshake will be canceled if ctx is canceled.
func (u *Transport) Dial(ctx context.Context, addr net.Addr) error {
	as := addr.String()

	u.handshakesMtx.Lock()
	defer u.handshakesMtx.Unlock()

	if _, ok := u.handshakes[as]; !ok {
		u.le.WithField("addr", as).Debug("pushing new handshaker")
		_, err := u.pushHandshaker(ctx, addr, true)
		return err
	}

	return nil
}

// Execute processes the transport, emitting events to the handler.
// Fatal errors are returned.
func (u *Transport) Execute(ctx context.Context, handler transport.Handler) error {
	u.ctx = ctx
	laddr := *u.pc.LocalAddr().(*net.UDPAddr)
	laddr.Port = 0
	var buf []byte
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := u.pc.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
			return err
		}

		if buf == nil {
			buf = xmitBuf.Get().([]byte)
			buf = buf[:cap(buf)]
		}

		n, addr, err := u.pc.ReadFrom(buf)
		if err != nil {
			if e, ok := err.(net.Error); !ok || (!e.Timeout() || !e.Temporary()) {
				return err
			}

			continue
		}

		if n < 2 {
			continue
		}

		if err := u.handlePacket(ctx, buf[:n], addr); err != nil {
			u.le.
				WithError(err).
				WithField("addr", addr.String()).
				Debug("dropped packet")
		} else {
			// u.le.WithField("length", n).WithField("addr", addr.String()).Debugf("handled packet: %#v", buf[:n])
			buf = nil
		}
	}
}

// handlePacket handles an incoming packet.
func (u *Transport) handlePacket(ctx context.Context, buf []byte, addr net.Addr) error {
	as := addr.String()
	packetType := PacketType(buf[len(buf)-1])
	if err := packetType.Validate(); err != nil {
		return err
	}
	buf = buf[:len(buf)-1]

	var err error
	if packetType == PacketType_PacketType_HANDSHAKE {
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

		hs.hs.Handle(buf)
		xmitBuf.Put(buf)
		return nil
	}

	u.linksMtx.Lock()
	var link *Link
	if u.lastLinkAddr == as && u.lastLink != nil {
		link = u.lastLink
	} else {
		var ok bool
		link, ok = u.links[as]
		if ok {
			u.lastLinkAddr = as
			u.lastLink = link
		}
	}
	u.linksMtx.Unlock()

	if link == nil {
		return errors.Errorf("unknown remote link: %s", as)
	}

	link.HandlePacket(packetType, buf)
	return nil
}

// LocalAddr returns the local address.
func (u *Transport) LocalAddr() net.Addr {
	return u.pc.LocalAddr()
}

// Close closes the connection.
func (u *Transport) Close() error {
	return u.pc.Close()
}

// pushHandshaker builds a new handshaker for the address.
// it is expected that handshakesMtx is locked before calling pushHandshaker
func (u *Transport) pushHandshaker(ctx context.Context, addr net.Addr, inititiator bool) (*inflightHandshake, error) {
	as := addr.String()
	nctx, nctxCancel := context.WithTimeout(ctx, handshakeTimeout)
	hs := &inflightHandshake{ctxCancel: nctxCancel, addr: addr}
	var err error
	hs.hs, err = s2s.NewHandshaker(
		u.privKey,
		nil,
		func(data []byte) error {
			data = append(data, byte(PacketType_PacketType_HANDSHAKE))
			_, err := u.pc.WriteTo(data, addr)
			return err
		},
		nil,
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
func (u *Transport) processHandshake(ctx context.Context, hs *inflightHandshake, initiator bool) {
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

	u.handleCompleteHandshake(hs.addr, res, initiator)
}

// GetLinks returns the links currently active.
func (u *Transport) GetLinks() (lnks []link.Link) {
	u.linksMtx.Lock()
	defer u.linksMtx.Unlock()

	lnks = make([]link.Link, 0, len(u.links))
	for _, lnk := range u.links {
		lnks = append(lnks, lnk)
	}

	return
}

// _ is a type assertion.
var _ transport.Transport = ((*Transport)(nil))
