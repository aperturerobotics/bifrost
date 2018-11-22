package pconn

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/aperturerobotics/controllerbus/directive"
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
	// peerID is the node peer id
	peerID peer.ID
	// handler is the transport handler
	handler transport.TransportHandler
	// opts are extra options
	opts Opts

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
}

// New builds a new packet-conn based transport, listening on the addr.
func New(
	le *logrus.Entry,
	pc net.PacketConn,
	pKey crypto.PrivKey,
	tc transport.TransportHandler,
	opts *Opts,
) *Transport {
	uuid := scrc.Crc64([]byte(pc.LocalAddr().String()))
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
		opts:    *opts,

		handshakes: make(map[string]*inflightHandshake),
		links:      make(map[string]*Link),

		readErrCh: make(chan error, 1),
	}
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
		u.le.WithField("addr", as).Debug("pushing new handshaker [dial]")
		_, err := u.pushHandshaker(ctx, addr, true)
		return err
	}

	return nil
}

// Execute processes the transport, emitting events to the handler.
// Fatal errors are returned.
func (u *Transport) Execute(ctx context.Context) error {
	u.ctx = ctx
	go u.readPump(ctx)

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
			/*
				u.le.
					WithField("length", n).
					WithField("addr", addr.String()).
					Debugf("handled packet: %#v", buf[:n])
			*/
			buf = nil
		}
	}
}

// LocalAddr returns the local address.
func (u *Transport) LocalAddr() net.Addr {
	return u.pc.LocalAddr()
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
			ale.Debug("pushing new handshaker [accept]")
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

// handleLinkLost is called when a link is lost.
func (u *Transport) handleLinkLost(addr string, lnk *Link) {
	u.linksMtx.Lock()
	existing := u.links[addr]
	rel := existing == lnk
	if rel {
		delete(u.links, addr)
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

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Transport) HandleDirective(ctx context.Context, di directive.Instance) (directive.Resolver, error) {
	// TODO
	return nil, nil
}

// Close closes the connection.
func (u *Transport) Close() error {
	return u.pc.Close()
}

// _ is a type assertion.
var _ transport.Transport = ((*Transport)(nil))
