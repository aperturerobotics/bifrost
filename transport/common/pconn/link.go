package pconn

import (
	"context"
	"encoding/binary"
	"hash/crc32"
	"net"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/handshake/identity"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/paralin/kcp-go-lite"
	"github.com/sirupsen/logrus"
	"github.com/xtaci/smux"
)

// Link represents a KCP-based connection/link.
type Link struct {
	// ctx is the context for this link
	ctx context.Context
	// ctxCancel cancels the context for this link.
	ctxCancel context.CancelFunc
	// le is the log entry
	le *logrus.Entry
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
	// sess is the kcp session
	sess *kcp.UDPSession
	// uuid is the link uuid
	uuid uint64
	// transportUUID is the transport uuid
	transportUUID uint64
	// remoteTransportUUID is the remote transport uuid
	remoteTransportUUID uint64
	// localPeerID is the local peer ID
	localPeerID peer.ID
	// closed is the closed callback
	closed func()
	// closedOnce guards closed
	closedOnce sync.Once
	// acceptStreamCh contains incoming streams
	acceptStreamCh chan *acceptedStream

	// rawStreamsMtx guards rawStreams
	rawStreamsMtx sync.Mutex
	// lastRawStream contains the last used raw stream.
	lastRawStream *rawStream
	// lastRawStreamID is the ID of the *rawStream in lastRawStream field
	lastRawStreamID uint32
	// rawStreams contains all raw streams by ID
	rawStreams map[uint32]*rawStream
	// nextRawStreamID is the next raw stream id to use for local stream identification
	nextRawStreamID uint32
	// rawStreamEstablishQueueInc is the set of rawStream that have not yet been
	// established (incoming, capped remotely)
	rawStreamEstablishQueueInc []*rawStream
	// inflightRawStreamEstablishOut is the number of in-flight *rawStream establishes
	inflightRawStreamEstablishOut uint32
	// rawStreamEstablishQueueOut is the set of rawStream that have not yet been
	// established (outgoing, no cap)
	rawStreamEstablishQueueOut []*rawStream
	// mtu is the maximum transmission unit
	mtu uint32

	// coordStreamCh holds the coordination stream
	coordStreamCh chan stream.Stream

	// writer is the writer function
	writer func(b []byte, addr net.Addr) (n int, err error)
}

// acceptedStream temporarily holds an incoming stream.
type acceptedStream struct {
	stream     stream.Stream
	streamOpts stream.OpenOpts
}

// NewLink builds a new link.
func NewLink(
	ctx context.Context,
	le *logrus.Entry,
	opts *Opts,
	localPeerID peer.ID,
	localAddr, remoteAddr net.Addr,
	transportUUID, remoteTransportUUID uint64,
	neg *identity.Result,
	writer func(b []byte, addr net.Addr) (n int, err error),
	initiator bool,
	closed func(),
) (*Link, error) {
	sharedSecret := neg.Secret
	mtu := opts.GetMtu()
	if mtu == 0 {
		mtu = 1350
	}

	nctx, nctxCancel := context.WithCancel(ctx)
	pid, _ := peer.IDFromPublicKey(neg.Peer)
	l := &Link{
		ctx:       nctx,
		ctxCancel: nctxCancel,

		neg:                 neg,
		le:                  le,
		mtu:                 mtu,
		addr:                remoteAddr,
		uuid:                newLinkUUID(localAddr, remoteAddr, pid),
		peerID:              pid,
		closed:              closed,
		writer:              writer,
		localAddr:           localAddr,
		rawStreams:          make(map[uint32]*rawStream),
		coordStreamCh:       make(chan stream.Stream, 1),
		localPeerID:         localPeerID,
		sharedSecret:        sharedSecret,
		transportUUID:       transportUUID,
		remoteTransportUUID: remoteTransportUUID,
		nextRawStreamID:     1,
		acceptStreamCh:      make(chan *acceptedStream),
	}

	// build conv id from shared secret
	convid := binary.LittleEndian.Uint32(sharedSecret[:4])
	dataShards := opts.GetDataShards()
	parityShards := opts.GetParityShards()
	bc, err := BuildBlockCrypt(opts.GetBlockCrypt(), neg.Secret[:])
	if err != nil {
		return nil, err
	}

	l.sess = kcp.NewUDPSession(
		func(b []byte) (n int, err error) {
			b = append(b, byte(PacketType_PacketType_KCP_SMUX))
			n, err = writer(b, l.addr)
			if n > 0 {
				n--
			}
			return
		},
		convid,
		int(dataShards),
		int(parityShards),
		bc,
	)

	l.sess.SetStreamMode(true)
	// l.sess.SetStreamMode(false)
	l.sess.SetMtu(int(mtu))

	kcpMode := opts.GetKcpMode()
	switch kcpMode {
	case KCPMode_KCPMode_UNKNOWN:
		fallthrough
	case KCPMode_KCPMode_NORMAL:
		l.sess.SetNoDelay(0, 100, 0, 0)
	case KCPMode_KCPMode_FAST:
		l.sess.SetNoDelay(0, 40, 2, 1)
	case KCPMode_KCPMode_FAST2:
		l.sess.SetNoDelay(1, 20, 2, 1)
	case KCPMode_KCPMode_FAST3:
		l.sess.SetNoDelay(1, 10, 2, 1)
	case KCPMode_KCPMode_SLOW1:
		l.sess.SetNoDelay(0, 200, 0, 0)
	}

	if kcpMode == KCPMode_KCPMode_SLOW1 ||
		kcpMode == KCPMode_KCPMode_NORMAL ||
		kcpMode == KCPMode_KCPMode_FAST {
		l.sess.SetWriteDelay(true)
		l.sess.SetACKNoDelay(false)
		l.sess.SetStreamMode(true)
		// Bandwidth-in-bits-per-second * Round-trip-latency-in-seconds = TCP window size in bytes
		// 10000*300
		l.sess.SetWindowSize(3000000, 3000000)
	} else {
		l.sess.SetWriteDelay(false)
		l.sess.SetWindowSize(1024*12, 1024*12)
		l.sess.SetACKNoDelay(true)
	}

	conf := smux.DefaultConfig()
	conf.KeepAliveInterval = time.Second * 5
	conf.KeepAliveTimeout = time.Second * 13
	conf.MaxReceiveBuffer = 4194304
	if initiator {
		l.mux, _ = smux.Server(l.sess, conf)
	} else {
		l.mux, _ = smux.Client(l.sess, conf)
	}

	go l.smuxAcceptPump(initiator)

	return l, nil
}

// GetUUID returns the link unique id.
func (l *Link) GetUUID() uint64 {
	return l.uuid
}

// AcceptStream accepts a stream from the link.
func (l *Link) AcceptStream() (stream.Stream, stream.OpenOpts, error) {
	var astrm *acceptedStream
	for astrm == nil {
		l.rawStreamsMtx.Lock()
		inc := l.drainIncomingEstablishQueue()
		l.rawStreamsMtx.Unlock()
		if inc != nil {
			return inc, stream.OpenOpts{}, nil
		}

		select {
		case <-l.ctx.Done():
			return nil, stream.OpenOpts{}, l.ctx.Err()
		case astrm = <-l.acceptStreamCh:
		}
	}

	s := astrm.stream
	opts := astrm.streamOpts
	l.le.
		WithField("stream-reliable", opts.Reliable).
		WithField("stream-encrypted", opts.Encrypted).
		WithField("remote-peer", l.GetRemotePeer().Pretty()).
		Info("accepted stream")
	return s, opts, nil
}

// computeConvID computes the conversation id using the shared secret
func computeConvID(sharedSecret []byte) uint32 {
	return crc32.ChecksumIEEE(sharedSecret)
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

// GetRemoteTransportUUID returns the unique ID of the remote transport.
// Reported by the remote peer. May be zero or unreliable value.
func (l *Link) GetRemoteTransportUUID() uint64 {
	return l.remoteTransportUUID
}

// GetRemotePeer returns the identity of the remote peer if encrypted.
func (l *Link) GetRemotePeer() peer.ID {
	return l.peerID
}

// GetLocalPeer returns the identity of the local peer.
func (l *Link) GetLocalPeer() peer.ID {
	return l.localPeerID
}

// OpenStream opens a stream on the link, with the given parameters.
func (l *Link) OpenStream(opts stream.OpenOpts) (stream.Stream, error) {
	// ensure coord stream exists
	select {
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	case strm := <-l.coordStreamCh:
		l.coordStreamCh <- strm
	}

	if opts.Encrypted || opts.Reliable {
		strm, err := l.mux.OpenStream()
		return strm, err
	}

	l.rawStreamsMtx.Lock()
	defer l.rawStreamsMtx.Unlock()
	return l.openRawStream()
}

// HandlePacket handles a packet.
func (l *Link) HandlePacket(packetType PacketType, data []byte) {
	// l.le.WithField("packet-type", packetType.String()).Debugf("handling packet: %#v", data)
	switch packetType {
	case PacketType_PacketType_RAW:
		l.handleRawPacket(data)
	case PacketType_PacketType_KCP_SMUX:
		l.sess.RxPacket(data)
	case PacketType_PacketType_CLOSE_LINK:
		l.le.Debug("received close_link packet")
		l.writer = nil
		l.Close()
	}
}

// handleRawPacket handles an incoming raw packet.
func (l *Link) handleRawPacket(data []byte) {
	// expect varint stream ID as suffix
	// reversed, take last 4 bytes
	// l.le.Infof("handleRawPacket: %v", data)
	fi := len(data) - 5
	if fi < 0 {
		fi = 0
	}
	streamID, varintBytes := decodeRawStreamIDVarint(data[fi:])
	if varintBytes == 0 {
		l.le.Warn("dropped raw packet with invalid varint trailer")
		return
	}

	data = data[:len(data)-varintBytes]
	if len(data) == 0 {
		l.le.Warn("dropped raw packet with empty body")
		return
	}

	if streamID > 4294967290 {
		l.le.
			WithField("stream-id", streamID).
			Warn("dropped raw packet with invalid stream id")
		return
	}

	sid := uint32(streamID)
	l.rawStreamsMtx.Lock()
	strm, ok := l.rawStreams[sid]
	l.rawStreamsMtx.Unlock()

	if ok {
		strm.PushPacket(data)
	} else {
		l.le.
			WithField("stream-id", sid).
			Warn("dropped raw packet with unknown stream id")
	}
	// xmitBuf.Put(data[:cap(data)])
}

// Close closes the connection.
func (l *Link) Close() error {
	// TODO race on l.writer
	if l.writer != nil {
		_, _ = l.writer(
			[]byte{byte(PacketType_PacketType_CLOSE_LINK)},
			l.addr,
		)
		l.writer = nil
	}
	if closed := l.closed; closed != nil {
		l.closedOnce.Do(closed)
	}
	l.ctxCancel()
	if l.mux != nil {
		_ = l.mux.Close()
	}
	if l.sess != nil {
		_ = l.sess.Close()
	}
	return nil
}

// _ is a type assertion
var _ link.Link = ((*Link)(nil))
