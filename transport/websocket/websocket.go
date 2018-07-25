//+build !js

package websocket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/util/scrc"
	"github.com/blang/semver"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// Version is the version of the websocket implementation.
var Version = semver.MustParse("0.0.1")

// handshakeTimeout is the time after which a handshake expires
var handshakeTimeout = time.Second * 8

// upgrader is the websocket upgrader
var upgrader = &websocket.Upgrader{
	// Allow any origin
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Transport is a Websocket based transport.
// This implements the non-browser end.
type Transport struct {
	ctx context.Context

	// le is the logger
	le *logrus.Entry
	// uuid is the unique id
	uuid uint64
	// privKey is the local priv key
	privKey crypto.PrivKey

	// listenErrCh is the listen error channel
	listenErrCh chan error

	// handshakesMtx guards the handshakes map
	handshakesMtx sync.Mutex
	// handshakes is the set of ongoing handshakes
	handshakes map[string]*inflightHandshake

	// linksMtx guards the links map
	linksMtx sync.Mutex
	// links is the set of active links
	links map[string]*Link

	// server is the http server
	server *http.Server
	// bootDialAddrs are addresses to dial on boot
	bootDialAddrs []string
}

// New builds a new packet-conn based transport, listening on the addr.
func New(
	le *logrus.Entry,
	listenStr string,
	bootDialAddrs []string,
	pKey crypto.PrivKey,
) *Transport {
	uuid := scrc.Crc64([]byte(listenStr))

	t := &Transport{
		le:      le,
		privKey: pKey,
		uuid:    uuid,

		handshakes: make(map[string]*inflightHandshake),
		links:      make(map[string]*Link),

		listenErrCh:   make(chan error, 1),
		bootDialAddrs: bootDialAddrs,
	}

	if listenStr != "" {
		mux := http.NewServeMux()
		mux.Handle("/ws/bifrost-0.1", t)
		t.server = &http.Server{Addr: listenStr, Handler: mux}
	}

	return t
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
		u.le.WithField("addr", url).Debug("pushing new handshaker")
		_, err := u.pushHandshaker(ctx, url, nil)
		return err
	}

	return nil
}

// ServeHTTP serves the bifrost-0.1 protocol.
func (u *Transport) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	le := u.le.WithField("remote-addr", req.RemoteAddr)
	conn, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		le.WithError(err).Warn("unable to upgrade ws conn")
		return
	}

	_, err = u.pushHandshaker(req.Context(), req.RemoteAddr, conn)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		conn.Close()
		return
	}

	// TODO: do something smarter than holding the conn open
	<-req.Context().Done()
}

// Execute processes the transport.
// Fatal errors are returned.
func (u *Transport) Execute(ctx context.Context) error {
	u.ctx = ctx

	if u.server != nil {
		go func() {
			u.listenErrCh <- u.server.ListenAndServe()
		}()
	}

	for _, d := range u.bootDialAddrs {
		go u.Dial(ctx, d)
	}

	// TODO: when returning, close all links
	select {
	case <-ctx.Done():
		if u.server != nil {
			return u.server.Shutdown(ctx)
		}

		return ctx.Err()
	case rerr := <-u.listenErrCh:
		return rerr
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

// Close closes the connection.
func (u *Transport) Close() error {
	return u.server.Shutdown(u.ctx)
}

// _ is a type assertion.
var _ transport.Transport = ((*Transport)(nil))
