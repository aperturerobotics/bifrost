package pconn

import (
	"context"
	"hash/crc32"
	"net"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/sirupsen/logrus"
	"github.com/xtaci/kcp-go"
	"github.com/xtaci/smux"
)

// Link represents a KCP-based connection/link.
type Link struct {
	// ctx is the context for this link
	ctx context.Context
	// ctxCancel cancels the context for this link.
	ctxCancel context.CancelFunc
	// addr is the bound remote address
	addr net.Addr
	// localAddr is the local address
	localAddr net.Addr
	// neg is the negotiated identity data
	neg *identity.Result
	// sharedSecret is the shared secret
	sharedSecret [32]byte
	// peerID is the remote peer id
	peerID peer.ID
	// mux is the reliable stream multiplexer
	mux *smux.Session
	// kpc is the kcp packet conn interface
	kpc *kcpPacketConn
	// sess is the kcp session
	sess *kcp.UDPSession
	// uuid is the link uuid
	uuid uint64
	// transportUUID is the transport uuid
	transportUUID uint64
}

// NewLink builds a new link.
func NewLink(
	ctx context.Context,
	localAddr, remoteAddr net.Addr,
	transportUUID uint64,
	neg *identity.Result,
	sharedSecret [32]byte,
	writer func(b []byte, addr net.Addr) (n int, err error),
	initiator bool,
) *Link {
	nctx, nctxCancel := context.WithCancel(ctx)
	pid, _ := peer.IDFromPublicKey(neg.Peer)
	l := &Link{
		ctx:           nctx,
		ctxCancel:     nctxCancel,
		sharedSecret:  sharedSecret,
		localAddr:     localAddr,
		addr:          remoteAddr,
		neg:           neg,
		peerID:        pid,
		uuid:          newLinkUUID(localAddr, remoteAddr, pid),
		transportUUID: transportUUID,
	}

	// dummy raddr
	l.kpc = newKcpPacketConn(
		l.ctx,
		l.localAddr,
		l.addr,
		func(b []byte, addr net.Addr) (n int, err error) {
			b = append(b, byte(PacketType_PacketType_KCP_SMUX))
			n, err = writer(b, l.addr)
			if n > 0 {
				n--
			}
			return
		},
		l.Close,
	)
	l.sess, _ = kcp.NewConn(
		// computeConvID(sharedSecret[:]),
		dummyKcpRemoteAddr.String(),
		l.buildBlockCrypt(),
		0, 0,
		l.kpc,
	)
	l.sess.SetMtu(1350)

	if initiator {
		l.mux, _ = smux.Server(l.sess, smux.DefaultConfig())
	} else {
		l.mux, _ = smux.Client(l.sess, smux.DefaultConfig())
	}

	go l.acceptStreamPump()
	return l
}

// GetUUID returns the link unique id.
func (l *Link) GetUUID() uint64 {
	return l.uuid
}

// computeConvID computes the conversation id using the shared secret
func computeConvID(sharedSecret []byte) uint32 {
	return crc32.ChecksumIEEE(sharedSecret)
}

// acceptStreamPump goroutine accepts incoming streams.
func (l *Link) acceptStreamPump() {
	for {
		s, err := l.mux.AcceptStream()
		if err != nil {
			logrus.WithError(err).Error("stopped accepting stream")
			_ = l.Close()
			return
		}

		logrus.
			WithField("stream-id", s.ID()).
			WithField("remote-peer", l.GetRemotePeer().Pretty()).
			Info("accepted stream")

		strm := &smuxStream{Stream: s}
		// TODO: handle incoming stream
		_ = strm
	}
}

// buildBlockCrypt returns the block crypto for this link.
func (l *Link) buildBlockCrypt() (c kcp.BlockCrypt) {
	c, _ = kcp.NewSalsa20BlockCrypt(l.sharedSecret[:])
	return
}

// newLinkUUID builds the UUID for a link
func newLinkUUID(localAddr, remoteAddr net.Addr, peerID peer.ID) uint64 {
	return scrc.Crc64(
		[]byte("udp"),
		[]byte(localAddr.String()),
		[]byte(remoteAddr.String()),
		[]byte(peerID),
	)
}

// GetTransportUUID returns the unique ID of the transport.
func (l *Link) GetTransportUUID() uint64 {
	return l.transportUUID
}

// GetRemotePeer returns the identity of the remote peer if encrypted.
func (l *Link) GetRemotePeer() peer.ID {
	return l.peerID
}

// OpenStream opens a stream on the link, with the given parameters.
func (l *Link) OpenStream(opts stream.OpenOpts) (stream.Stream, error) {
	s, err := l.mux.OpenStream()
	if err != nil {
		return nil, err
	}

	strm := &smuxStream{Stream: s}
	return strm, nil
}

// HandlePacket handles a packet.
func (l *Link) HandlePacket(packetType PacketType, data []byte) {
	switch packetType {
	case PacketType_PacketType_RAW:
		// TODO: route raw packet to stream
	case PacketType_PacketType_KCP_SMUX:
		l.kpc.pushPacket(data)
	}
}

// Close closes the connection.
func (l *Link) Close() error {
	logrus.Warnf("closing conn: %v", l.GetRemotePeer().Pretty())
	_ = l.mux.Close()
	_ = l.sess.Close()
	l.ctxCancel()
	return nil
}
