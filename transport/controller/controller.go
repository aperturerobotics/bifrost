package transport_controller

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/blang/semver"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// streamEstablishTimeout is the max time to wait for a stream header.
var streamEstablishTimeout = time.Second * 5

// streamEstablishMaxPacketSize is the maximum stream establish header size
var streamEstablishMaxPacketSize = 100000

// Constructor constructs a transport with common parameters.
type Constructor func(
	ctx context.Context,
	le *logrus.Entry,
	pkey crypto.PrivKey,
	handler transport.TransportHandler,
) (transport.Transport, error)

// Controller implements a common transport controller.
// The controller looks up the Node, acquires its identity, constructs the
// transport, and manages the lifecycle of dialing and accepting links.
type Controller struct {
	// ctx is the controller context
	// set in the execute() function
	// ensure not used before execute sets it.
	ctx context.Context
	// le is the logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// ctor is the constructor
	ctor Constructor
	// peerIDConstraint constrains the node peer id constraint
	peerIDConstraint peer.ID
	// localPeerID contains the node peer ID
	localPeerID peer.ID

	// tptCh holds the transport like a bucket
	tptCh chan transport.Transport

	// linksMtx is the links mutex
	linksMtx sync.Mutex
	// links is the links set, keyed by link uuid
	links map[uint64]*establishedLink
	// linkWaiters is a set of callbacks waiting for connections with peers.
	linkWaiters map[peer.ID][]*linkWaiter
	// linkDialers tracks ongoing dial attempts
	linkDialers map[linkDialerKey]*linkDialer

	// transportID is the transport identifier.
	transportID string
	// transportVersion is the transport version
	transportVersion semver.Version

	// staticPeerMap maps a peer ID to a peermap.DialPeer
	// when EstablishLink matches a peer ID in this map,
	// the transport controller will dial the peer.
	staticPeerMap map[string]*dialer.DialerOpts
}

// NewController constructs a new transport controller.
func NewController(
	le *logrus.Entry,
	bus bus.Bus,
	nodePeerIDConstraint peer.ID,
	ctor Constructor,
	transportID string,
	transportVersion semver.Version,
	staticPeerMap map[string]*dialer.DialerOpts,
) *Controller {
	return &Controller{
		le:   le,
		bus:  bus,
		ctor: ctor,

		links:       make(map[uint64]*establishedLink),
		linkWaiters: make(map[peer.ID][]*linkWaiter),
		tptCh:       make(chan transport.Transport, 1),

		peerIDConstraint: nodePeerIDConstraint,
		localPeerID:      nodePeerIDConstraint,
		transportID:      transportID,
		transportVersion: transportVersion,
		linkDialers:      make(map[linkDialerKey]*linkDialer),

		staticPeerMap: staticPeerMap,
	}
}

// GetControllerID returns the controller ID.
func (c *Controller) GetControllerID() string {
	return strings.Join([]string{
		"bifrost",
		"transport",
		c.transportID,
		c.transportVersion.String(),
	}, "/")
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() controller.Info {
	return controller.NewInfo(
		c.GetControllerID(),
		c.transportVersion,
		"transport controller "+c.transportID+"@"+c.transportVersion.String(),
	)
}

// Execute executes the transport controller and the transport.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	c.ctx = ctx
	// Acquire a handle to the node.
	c.le.
		WithField("peer-id", c.peerIDConstraint.Pretty()).
		Info("looking up peer with ID")
	n, err := peer.GetPeerWithID(ctx, c.bus, c.peerIDConstraint)
	if err != nil {
		return err
	}

	// Get the priv key
	privKey := n.GetPrivKey()
	c.localPeerID, err = peer.IDFromPrivateKey(privKey)
	if err != nil {
		return err
	}

	// Construct the transport
	tpt, err := c.ctor(
		ctx,
		c.le,
		privKey,
		c,
	)
	if err != nil {
		return err
	}
	defer tpt.Close()

	c.tptCh <- tpt

	tptErr := make(chan error, 1)
	go func() {
		c.le.Debug("executing transport")
		if err := tpt.Execute(ctx); err != nil {
			tptErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-tptErr:
		return err
	}
}

// GetTransport returns the controlled transport.
// This may be nil until the transport is constructed.
func (c *Controller) GetTransport(ctx context.Context) (transport.Transport, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case tpt := <-c.tptCh:
		c.tptCh <- tpt
		return tpt, nil
	}
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(ctx context.Context, di directive.Instance) (directive.Resolver, error) {
	dir := di.GetDirective()
	switch d := dir.(type) {
	case link.OpenStreamWithPeer:
		return c.resolveOpenStreamWithPeer(ctx, di, d)
	case link.EstablishLinkWithPeer:
		return c.resolveEstablishLink(ctx, di, d)
	case transport.LookupTransport:
		return c.resolveLookupTransport(ctx, di, d)
	}

	return nil, nil
}

// HandleLinkEstablished is called by the transport when a link is established.
func (c *Controller) HandleLinkEstablished(lnk link.Link) {
	le := c.loggerForLink(lnk)
	c.linksMtx.Lock()
	defer c.linksMtx.Unlock()

	pidStr := lnk.GetRemotePeer().Pretty()
	for k, d := range c.linkDialers {
		if k.peerID == pidStr {
			delete(c.linkDialers, k)
			d.cancel()
		}
	}

	// quick sanity check
	if lnk.GetRemotePeer() == c.localPeerID {
		le.Warn("self-dial detected, closing link")
		go lnk.Close()
		return
	}

	luuid := lnk.GetUUID()
	el, elOk := c.links[luuid]
	if elOk {
		if el.Link == lnk {
			// duplicate HandleLinkEstablished call
			le.Debug("duplicate handle-link-established call")
			return
		}

		// close dupe
		le.Debug("closing existing link identical to incoming link")
		go c.flushEstablishedLink(el)
		delete(c.links, luuid)
	}

	// construct new established link
	el, err := newEstablishedLink(c.le, c.ctx, c.bus, lnk, c)
	if err != nil {
		c.le.WithError(err).Warn("unable to construct established link")
		return
	}
	c.links[luuid] = el
	le.Debug("link established")

	// flush any relevant link waiters
	c.resolveLinkWaiters(el.Link, true)
}

// HandleIncomingStream handles an incoming stream from a link. It negotiates
// the protocol for the stream, acquires a handler for the protocol, and hands
// the stream to the protocol handler, then returns. Uses the ctx for
// cancellation.
func (c *Controller) HandleIncomingStream(
	rctx context.Context,
	lnk link.Link,
	strm stream.Stream,
	strmOpts stream.OpenOpts,
) {
	// Assert EstablishLink to keep the stream open during the header exchange.
	_, elRef, err := c.bus.AddDirective(
		link.NewEstablishLinkWithPeer(lnk.GetRemotePeer()),
		nil,
	)
	if err == nil {
		defer elRef.Release()
	}

	readDeadline := time.Now().Add(streamEstablishTimeout)
	ctx, ctxCancel := context.WithDeadline(rctx, readDeadline)
	defer ctxCancel()
	strm.SetReadDeadline(readDeadline)

	// process stream establish header;
	streamEst, err := readStreamEstablishHeader(strm)
	if err != nil {
		c.le.WithError(err).Warn("unable to read stream establish header")
		strm.Close()
		return
	}
	strm.SetDeadline(time.Time{})

	// received stream establish header, now, create handlestream directive
	pid := protocol.ID(streamEst.GetProtocolId())
	if err := pid.Validate(); err != nil {
		c.le.
			WithError(err).
			WithField("stream-protocol-id", streamEst.GetProtocolId()).
			Warn("failed to validate protocol id")
		strm.Close()
		return
	}

	var mstrm link.MountedStream = newMountedStream(strm, strmOpts, pid, lnk)
	_ = mstrm

	// bus is the controller bus
	le := c.loggerForLink(lnk).WithField("protocol-id", pid)
	le.Debug("handling incoming stream")
	dir := link.NewHandleMountedStreamWithProtocolID(pid, c.localPeerID, mstrm.GetPeerID())
	dval, dref, err := bus.ExecOneOff(ctx, c.bus, dir, nil)
	if err != nil {
		le.WithError(err).Warn("error retrieving stream handler for stream")
		strm.Close()
		return
	}
	defer dref.Release()

	mhnd, ok := dval.(link.MountedStreamHandler)
	if !ok {
		c.le.
			WithError(err).
			WithField("protocol-id", pid).
			Warn("stream handler retrieved is not a link.MountedStreamHandler")
		strm.Close()
		return
	}

	if err := mhnd.HandleMountedStream(rctx, mstrm); err != nil {
		c.le.
			WithError(err).
			WithField("protocol-id", pid).
			Warn("stream handler returned an error")
		strm.Close()
		return
	}

	// stream is now handled by the handler.
}

// HandleLinkLost is called when a link is lost.
func (c *Controller) HandleLinkLost(lnk link.Link) {
	c.linksMtx.Lock()
	defer c.linksMtx.Unlock()

	// fast path: clear by uuid
	luuid := lnk.GetUUID()
	if el, elOk := c.links[luuid]; elOk {
		delete(c.links, luuid)
		c.flushEstablishedLink(el)
		return
	}

	// slow path: equality check to be sure
	for k, l := range c.links {
		if l.Link == lnk {
			delete(c.links, k)
			c.flushEstablishedLink(l)
			break
		}
	}
}

// PushDialer pushes a new dialer.
// Waits for the transport to be constructed.
// If the transport is not a TransportDialer, returns nil.
// Returns after the dialer is pushed.
func (c *Controller) PushDialer(
	ctx context.Context,
	peerID peer.ID,
	opts *dialer.DialerOpts,
) error {
	var tptDialer transport.TransportDialer
	select {
	case <-ctx.Done():
		return ctx.Err()
	case tpt := <-c.tptCh:
		c.tptCh <- tpt
		var ok bool
		tptDialer, ok = tpt.(transport.TransportDialer)
		if !ok {
			return nil
		}
	}

	key := linkDialerKey{
		peerID:      peerID.Pretty(),
		dialAddress: opts.GetAddress(),
	}
	go c.startLinkDialer(peerID, key, opts, tptDialer)
	return nil
}

// startLinkDialer starts a new link dialer.
func (c *Controller) startLinkDialer(
	peerID peer.ID,
	key linkDialerKey,
	opts *dialer.DialerOpts,
	tptDialer transport.TransportDialer,
) {
	c.linksMtx.Lock()
	_, ok := c.linkDialers[key]
	if !ok {
		dialer := dialer.NewDialer(c.le, tptDialer, opts, peerID, key.dialAddress)
		ctx, ctxCancel := context.WithCancel(c.ctx)
		ld := &linkDialer{
			dialer: dialer,
			cancel: ctxCancel,
		}
		c.linkDialers[key] = ld
		go c.executeDialer(ctx, key, ld)
	}
	c.linksMtx.Unlock()
}

// flushEstablishedLink closes an established link and cleans it up.
// linksmtx is locked by caller
func (c *Controller) flushEstablishedLink(el *establishedLink) {
	/*
		peerIDStr := el.Link.GetLocalPeer().Pretty()
		for ldk := range c.linkDialers {
			if ldk.peerID == peerIDStr {
				c.linkDialers[ldk].dialer.
			}
		}
	*/

	le := c.loggerForLink(el.Link)
	le.Debug("link lost/closed")
	c.resolveLinkWaiters(el.Link, false)
	el.Cancel()
	el.Link.Close()
}

// loggerForLink wraps a logger with fields identifying the link.
func (c *Controller) loggerForLink(lnk link.Link) *logrus.Entry {
	return c.le.
		WithField("link-uuid", lnk.GetUUID()).
		WithField("link-peer", lnk.GetRemotePeer().Pretty()).
		WithField("tpt-uuid", lnk.GetTransportUUID())
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	// nil references to help GC along
	c.ctor = nil
	c.le = nil
	c.bus = nil
	c.peerIDConstraint = ""

	select {
	case tpt := <-c.tptCh:
		tpt.Close()
	default:
	}

	return nil
}

// _ is a type assertion
var _ transport.Controller = ((*Controller)(nil))
