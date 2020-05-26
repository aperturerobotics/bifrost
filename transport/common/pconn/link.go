package pconn

import (
	"context"
	"net"
	"sync"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/bifrost/util/scrc"
	p2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/lucas-clemente/quic-go"
	"github.com/sirupsen/logrus"
)

// Link represents a Quic-based connection/link.
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
	// sess is the quic session
	sess quic.Session
	// peerID is the remote peer id
	peerID peer.ID
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
}

// NewLink builds a new link.
func NewLink(
	ctx context.Context,
	le *logrus.Entry,
	opts *Opts,
	localTransportUUID uint64,
	localPeerID peer.ID,
	localAddr net.Addr,
	sess quic.Session,
	closed func(),
) (*Link, error) {
	// The tls.Config used to establish this connection already verified the certificate chain.
	// Since we don't have any way of knowing which tls.Config was used though,
	// we have to re-determine the peer's identity here.
	// Therefore, this is expected to never fail.
	remotePubKey, err := p2ptls.PubKeyFromCertChain(sess.ConnectionState().PeerCertificates)
	if err != nil {
		return nil, err
	}
	remotePeerID, err := peer.IDFromPublicKey(remotePubKey)
	if err != nil {
		return nil, err
	}
	remoteAddr := sess.RemoteAddr()
	nctx, nctxCancel := context.WithCancel(ctx)
	uuid := newLinkUUID(localAddr, remoteAddr, remotePeerID)
	remoteTransportUUID := newTransportUUID(remoteAddr, remotePeerID)
	return &Link{
		ctx:       nctx,
		ctxCancel: nctxCancel,

		le:                  le.WithField("link-uuid", uuid),
		addr:                remoteAddr,
		uuid:                uuid,
		sess:                sess,
		peerID:              remotePeerID,
		closed:              closed,
		localAddr:           localAddr,
		localPeerID:         localPeerID,
		transportUUID:       localTransportUUID,
		remoteTransportUUID: remoteTransportUUID,
	}, nil
}

// GetUUID returns the link unique id.
func (l *Link) GetUUID() uint64 {
	return l.uuid
}

// AcceptStream accepts a stream from the link.
func (l *Link) AcceptStream() (stream.Stream, stream.OpenOpts, error) {
	qstream, err := l.sess.AcceptStream(l.ctx)
	if err != nil {
		return nil, stream.OpenOpts{}, err
	}

	opts := stream.OpenOpts{
		Reliable:  true,
		Encrypted: true,
	}
	return qstream, opts, nil
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
	return l.sess.OpenStreamSync(l.ctx)
}

// Close closes the connection.
func (l *Link) Close() error {
	l.closedOnce.Do(func() {
		if closed := l.closed; closed != nil {
			closed()
		}
		l.ctxCancel()
		/*
			if l.sess != nil {
				_ = l.sess.Close()
			}
		*/
	})
	return nil
}

// _ is a type assertion
var _ link.Link = ((*Link)(nil))
