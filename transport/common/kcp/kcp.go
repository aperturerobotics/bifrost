package kcp

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// handshakeTimeout is the time after which a handshake expires
var handshakeTimeout = time.Second * 8

// defaultMtu is the default mtu to use
var defaultMtu = 1300

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
	// peerID is the node peer id
	peerID peer.ID
	// handler is the transport handler
	handler transport.TransportHandler
	// opts are extra options
	opts *Opts

	// readErrCh indicates a read error
	readErrCh chan error

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

	// addrParser parses an address from a string
	// if nil, the dialer will not function
	addrParser func(addr string) (net.Addr, error)
}

// New builds a new packet-conn based transport, listening on the addr.
func New(
	le *logrus.Entry,
	uuid uint64,
	pc net.PacketConn,
	pKey crypto.PrivKey,
	addrParser func(addr string) (net.Addr, error),
	tc transport.TransportHandler,
	opts *Opts,
) *Transport {
	if opts == nil {
		opts = &Opts{}
	}
	pid, _ := peer.IDFromPrivateKey(pKey)
	return &Transport{
		le:      le.WithField("laddr", pc.LocalAddr().String()),
		pc:      pc,
		privKey: pKey,
		peerID:  pid,
		uuid:    uuid,
		handler: tc,
		opts:    opts,

		handshakes: make(map[string]*inflightHandshake),
		links:      make(map[string]*Link),

		readErrCh:  make(chan error, 1),
		addrParser: addrParser,
	}
}

// GetUUID returns a host-unique ID for this transport.
func (u *Transport) GetUUID() uint64 {
	return u.uuid
}

// DialPeer dials a peer given an address. The yielded link should be
// emitted to the transport handler. DialPeer should return nil if the link
// was established. DialPeer will then not be called again for the same peer
// ID and address tuple until the yielded link is lost.
func (u *Transport) DialPeer(ctx context.Context, peerID peer.ID, as string) (bool, error) {
	if u.addrParser == nil {
		return true, nil
	}

	addr, err := u.addrParser(as)
	if err != nil {
		return true, err
	}

	// abort if we already have a peer with the same id connected
	u.linksMtx.Lock()
	for _, lnk := range u.links {
		if lnk.peerID == peerID {
			u.linksMtx.Unlock()
			return false, nil
		}
	}
	u.linksMtx.Unlock()

	var hs *inflightHandshake
	var ok bool
	u.handshakesMtx.Lock()
	hs, ok = u.handshakes[as]
	if !ok {
		u.le.WithField("addr", as).Debug("pushing new handshaker [dial]")
		hs, err = u.pushHandshaker(ctx, addr, true)
		if err != nil {
			u.handshakesMtx.Unlock()
			return true, err
		}
	}
	u.handshakesMtx.Unlock()

	errCh := make(chan error, 1)
	hs.pushCompleteCb(func(err error) {
		errCh <- err
	})
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case err := <-errCh:
		return false, err
	}
}

// Execute processes the transport, emitting events to the handler.
// Fatal errors are returned.
func (u *Transport) Execute(ctx context.Context) error {
	u.ctx = ctx
	go func() {
		_ = u.readPump(ctx)
	}()

	u.le.Debug("listening")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case rerr := <-u.readErrCh:
		return rerr
	}
}

// readPump reads data from the listener
func (u *Transport) readPump(ctx context.Context) (readErr error) {
	defer func() {
		u.readErrCh <- readErr
	}()

	mtu := u.opts.GetMtu()
	if mtu == 0 {
		mtu = uint32(defaultMtu)
	}
	buf := make([]byte, mtu*2)

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
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				return err
			}

			continue
		}

		if n == 0 {
			continue
		}

		if err := u.handlePacket(ctx, buf[:n], addr); err != nil {
			u.le.
				WithError(err).
				WithField("addr", addr.String()).
				Debugf("dropped packet len(%d)", n)
		}
		buf = buf[:cap(buf)]
	}
}

// LocalAddr returns the local address.
func (u *Transport) LocalAddr() net.Addr {
	return u.pc.LocalAddr()
}

// handlePacket handles an incoming packet.
// buf must be copied before it returns
func (u *Transport) handlePacket(ctx context.Context, buf []byte, addr net.Addr) error {
	as := addr.String()
	packetType := PacketType(buf[len(buf)-1])
	if err := packetType.Validate(); err != nil {
		return err
	}
	buf = buf[:len(buf)-1]

	var err error
	if packetType == PacketType_PacketType_HANDSHAKE {
		ale := u.le.WithField("addr", as).WithField("data-len", len(buf))
		u.handshakesMtx.Lock()
		hs := u.handshakes[as]
		if hs == nil {
			ale.Debug("pushing new handshaker [accept]")
			hs, err = u.pushHandshaker(ctx, addr, false)
		} else {
			ale.Debug("used existing handshaker")
		}
		u.handshakesMtx.Unlock()
		if err != nil {
			return errors.Wrap(err, "build handshaker")
		}

		hs.pushPacket(buf)
		return nil
	}

	isCloseMarker := packetType == PacketType_PacketType_CLOSE_LINK
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
		if !isCloseMarker {
			u.handshakesMtx.Lock()
			hs, ok := u.handshakes[as]
			u.handshakesMtx.Unlock()
			if ok {
				hs.pushPacket(buf)
				return nil
			} else {
				_, _ = u.pc.WriteTo([]byte{byte(PacketType_PacketType_CLOSE_LINK)}, addr)
				return errors.Errorf("unknown remote link: %s", as)
			}
		} else {
			return nil
		}
	}

	link.HandlePacket(packetType, buf)
	return nil
}

// handleLinkLost is called when a link is lost.
func (u *Transport) handleLinkLost(addr string, lnk *Link) {
	u.linksMtx.Lock()
	existing := u.links[addr]
	rel := existing == lnk
	if rel {
		delete(u.links, addr)
	}
	if u.lastLink == lnk {
		u.lastLink = nil
		u.lastLinkAddr = ""
	}
	u.linksMtx.Unlock()

	if u.handler != nil && rel {
		u.handler.HandleLinkLost(lnk)
	}
}

// GetPeerID returns the node peer id.
func (u *Transport) GetPeerID() peer.ID {
	return u.peerID
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Transport) HandleDirective(ctx context.Context, di directive.Instance) (directive.Resolver, error) {
	// TODO
	return nil, nil
}

// Close closes the transport.
func (u *Transport) Close() error {
	return u.pc.Close()
}

// _ is a type assertion.
var _ transport.Transport = ((*Transport)(nil))
