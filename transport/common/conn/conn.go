package conn

import (
	"context"
	"io"
	"net"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	transport_quic "github.com/aperturerobotics/bifrost/transport/common/quic"
	"github.com/aperturerobotics/bifrost/util/rwc"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
)

// defaultMtu is the default max packet size to use
const defaultMtu = 65000 // 65Kb

// Transport implements a Bifrost transport with reliable conns.
//
// An example is a TCP connection: the OS provides an ordered stream of data as
// the interface for the Go program to use.
type Transport struct {
	// Transport is the underlying quic transport
	*transport_quic.Transport
	// ctx is the root context
	ctx context.Context
	// le is the logger
	le *logrus.Entry
	// handler is the transport handler
	handler transport.TransportHandler
	// opts are extra options
	opts *Opts
	// mtu is the max packet size to use
	mtu uint32
	// bufSize is the number of packets to keep in memory
	bufSize uint32
}

// AddrDialFunc dials an address.
type AddrDialFunc func(ctx context.Context, addr string) (io.ReadWriteCloser, net.Addr, error)

// NewTransport constructs a new conn-backed transport.
//
// addrDialer is an optional function to enable dialing out.
func NewTransport(
	ctx context.Context,
	le *logrus.Entry,
	privKey crypto.PrivKey,
	tc transport.TransportHandler,
	opts *Opts,
	// if uuid is 0, generates a default uuid based on peer id only
	uuid uint64,
	// laddr is the local address, if nil, generates one based on uuid
	laddr net.Addr,
	// addrDialer dials an address from a string.
	// if nil, the dialer will not function
	addrDialer AddrDialFunc,
) (*Transport, error) {
	if opts == nil {
		opts = &Opts{}
	}

	mtu := opts.GetMtu()
	if mtu <= 0 {
		mtu = defaultMtu
	}

	bufSize := opts.GetBufSize()
	if bufSize <= 0 {
		bufSize = 10
	}

	var dialFn transport_quic.DialFunc
	if addrDialer != nil {
		dialFn = func(dctx context.Context, addr string) (net.PacketConn, net.Addr, error) {
			c, na, err := addrDialer(dctx, addr)
			if err != nil {
				return nil, nil, err
			}
			pc := rwc.NewPacketConn(ctx, c, laddr, na, mtu, int(bufSize))
			return pc, na, err
		}
	}

	tpt := &Transport{
		ctx:     ctx,
		le:      le,
		handler: tc,
		opts:    opts,
		mtu:     mtu,
		bufSize: bufSize,
	}
	var err error
	tpt.Transport, err = transport_quic.NewTransport(
		ctx,
		le,
		uuid,
		laddr,
		privKey,
		tc,
		opts.GetQuic(),
		dialFn,
	)
	if err != nil {
		return nil, err
	}
	return tpt, nil
}

// HandleConn handles an incoming or outgoing connection.
//
// dial indicates if this is the originator (outgoing) conn or not
// ctx is used for the negotiation phase only
// if peerID is empty, allows any peer ID on the other end
// raddr can be nil if peerID is NOT empty
func (t *Transport) HandleConn(
	ctx context.Context,
	dial bool,
	c io.ReadWriteCloser,
	raddr net.Addr,
	peerID peer.ID,
) (*Link, error) {
	if raddr == nil {
		if len(peerID) == 0 {
			return nil, transport_quic.ErrRemoteUnspecified
		}
		raddr = peer.NewNetAddr(peerID)
	}

	pc := rwc.NewPacketConn(
		t.ctx,
		c,
		t.Transport.LocalAddr(),
		raddr,
		t.mtu,
		int(t.bufSize),
	)

	return t.Transport.HandleConn(ctx, dial, pc, raddr, peerID)
}

// _ is a type assertion
var _ transport.Transport = ((*Transport)(nil))
