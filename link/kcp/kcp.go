package kcp

import (
	"net"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/libp2p/go-libp2p-crypto"
	kp "github.com/xtaci/kcp-go"
)

// Link is a KCP-upgraded link.
type Link struct {
	// Link is the underlying link.
	link.Link
	// session is the kcp session
	sess *kp.UDPSession
	// pubKey is the public key
	pubKey crypto.PubKey
	// peerID is the remote peer id
	peerID peer.ID
}

// NewLink builds a new KCP-based encrypted link.
func NewLink(
	lnk link.Link,
	secret []byte,
	pubKey crypto.PubKey,
) (*Link, error) {
	bc, err := kp.NewSalsa20BlockCrypt(secret)
	if err != nil {
		return nil, err
	}

	var pc net.PacketConn
	if pcc, ok := lnk.(net.PacketConn); ok {
		pc = pcc
	} else {
		pc = link.NewPacketConn(lnk)
	}

	// use a dummy remote address
	dataShards := 10
	parityShard := 3
	sess, err := kp.NewConn("", bc, dataShards, parityShard, pc)
	if err != nil {
		return nil, err
	}

	if mtu := lnk.GetMTU(); mtu != 0 {
		sess.SetMtu(mtu)
	}

	peerID, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	return &Link{
		Link:   lnk,
		sess:   sess,
		pubKey: pubKey,
		peerID: peerID,
	}, nil
}

// Read reads a packet from the connection, copying the payload into b.
// It returns the number of bytes copied into b. ReadFrom can be made to
// time out and return an Error with Timeout() == true after a fixed time
// limit; see SetDeadline and SetReadDeadline.
// Returns io.EOF when the connection is closed.
func (l *Link) Read(b []byte) (n int, err error) {
	return l.sess.Read(b)
}

// Write writes a packet with payload b. an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline. On
// packet-oriented connections, write timeouts are rare. The size of the
// packet must be less than the MTU.
func (l *Link) Write(b []byte) (n int, err error) {
	return l.sess.Write(b)
}

// GetMTU returns the maximum-transmission-unit (max packet size).
// If 0, there is no packet size limited (indicate stream-oriented).
func (l *Link) GetMTU() int {
	return 0
}

// GetIsReliable returns if this link is ordered and reliable. A link is
// reliable if it provides both message ordering and message delivery
// guarantees.
func (l *Link) GetIsReliable() bool {
	return true
}

// GetRemotePeer returns the identity of the remote peer if encrypted.
// Returns a zero value if the connection is not pre-identified.
func (l *Link) GetRemotePeer() peer.ID {
	return l.peerID
}

// SetDeadline sets the read and write deadlines associated with the
// connection. It is equivalent to calling both SetReadDeadline and
// SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations fail with a
// timeout (see type Error) instead of blocking. The deadline applies to all
// future and pending I/O, not just the immediately following call to
// ReadFrom or WriteTo. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending the deadline
// after successful ReadFrom or WriteTo calls.
//
// A zero value for t means I/O operations will not time out.
func (l *Link) SetDeadline(t time.Time) error {
	return l.sess.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future ReadFrom calls and any
// currently-blocked ReadFrom call. A zero value for t means ReadFrom will
// not time out.
func (l *Link) SetReadDeadline(t time.Time) error {
	return l.sess.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future WriteTo calls and any
// currently-blocked WriteTo call. Even if write times out, it may return n
// > 0, indicating that some of the data was successfully written. A zero
// value for t means WriteTo will not time out.
func (l *Link) SetWriteDeadline(t time.Time) error {
	return l.sess.SetWriteDeadline(t)
}

// Close closes the connection.
// Any blocked ReadFrom or WriteTo operations will be unblocked and return errors.
func (l *Link) Close() error {
	defer l.Link.Close()
	return l.sess.Close()
}
