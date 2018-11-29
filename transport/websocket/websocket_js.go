//+build js

package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// Version is the version of the websocket implementation.
var Version = semver.MustParse("0.0.1")

// handshakeTimeout is the time after which a handshake expires
var handshakeTimeout = time.Second * 4

// Transport is a Websocket based transport.
type Transport struct {
	ctx context.Context

	// le is the logger
	le *logrus.Entry
	// uuid is the unique id
	uuid uint64
	// privKey is the local priv key
	privKey crypto.PrivKey
	// peerID is the local peer id
	peerID peer.ID
	// handler is the transport handler
	handler transport.TransportHandler

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

// New builds a new websocket based transport.
// In the browser, this can only dial out.
func New(
	le *logrus.Entry,
	_ string,
	pKey crypto.PrivKey,
	handler transport.TransportHandler,
) *Transport {
	uuid := scrc.Crc64([]byte("websocket/js"))
	peerID, _ := peer.IDFromPrivateKey(pKey)
	return &Transport{
		le:      le,
		privKey: pKey,
		uuid:    uuid,
		handler: handler,
		peerID:  peerID,

		handshakes: make(map[string]*inflightHandshake),
		links:      make(map[string]*Link),
	}
}

// GetUUID returns a host-unique ID for this transport.
func (u *Transport) GetUUID() uint64 {
	return u.uuid
}

// Dial instructs the transport to attempt to handshake with a peer.
// The function may return immediately.
// The handshake will be canceled if ctx is canceled.
func (u *Transport) Dial(ctx context.Context, url string) error {
	u.handshakesMtx.Lock()
	defer u.handshakesMtx.Unlock()

	if _, ok := u.handshakes[url]; !ok {
		u.le.WithField("url", url).Debug("pushing new handshaker [dial]")
		_, err := u.pushHandshaker(ctx, url)
		return err
	}

	return nil
}

// Execute processes the transport.
// Fatal errors are returned.
func (u *Transport) Execute(ctx context.Context) error {
	u.ctx = ctx
	return nil
}

// GetPeerID returns the peer ID.
func (u *Transport) GetPeerID() peer.ID {
	return u.peerID
}

// Close closes the transport.
func (u *Transport) Close() error {
	u.handshakesMtx.Lock()
	for k, h := range u.handshakes {
		h.ctxCancel()
		delete(u.handshakes, k)
	}
	u.handshakesMtx.Unlock()

	u.linksMtx.Lock()
	for k, l := range u.links {
		_ = l.Close()
		delete(u.links, k)
	}
	u.linksMtx.Unlock()

	return nil
}

// _ is a type assertion.
var _ transport.Transport = ((*Transport)(nil))
