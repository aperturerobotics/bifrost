//+build js

package websocket

import (
	"context"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/directive"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

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

	// addDirectiveCh indicates a new incoming directive
	addDirectiveCh chan directive.Directive

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
func New(le *logrus.Entry, pKey crypto.PrivKey) *Transport {
	uuid := scrc.Crc64([]byte("websocket/js"))
	return &Transport{
		le:      le,
		privKey: pKey,
		uuid:    uuid,

		handshakes: make(map[string]*inflightHandshake),
		links:      make(map[string]*Link),

		addDirectiveCh: make(chan directive.Directive, 5),
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
		u.le.WithField("url", url).Debug("pushing new handshaker")
		_, err := u.pushHandshaker(ctx, url)
		return err
	}

	return nil
}

// Execute processes the transport, emitting events to the handler.
// Fatal errors are returned.
func (u *Transport) Execute(ctx context.Context, handler transport.Handler) error {
	u.ctx = ctx

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d := <-u.addDirectiveCh:
			_ = d
		}
	}
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

// AddDirective adds a new directive to the transport.
func (u *Transport) AddDirective(d directive.Directive) {
	select {
	case <-u.ctx.Done():
	case u.addDirectiveCh <- d:
	}
}

// _ is a type assertion.
var _ transport.Transport = ((*Transport)(nil))
