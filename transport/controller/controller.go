package transport_controller

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/bifrost/tptaddr"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/ccontainer"
	"github.com/libp2p/go-libp2p/core/crypto"
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
	ctx atomic.Pointer[context.Context]
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
	// info is the controller info
	info *controller.Info

	// tptCtr contains the transport
	tptCtr *ccontainer.CContainer[*transport.Transport]

	// mtx guards the below fields
	mtx sync.Mutex
	// links is the set of active links, keyed by link uuid
	links map[uint64]*establishedLink
	// linkWaiters is a set of callbacks waiting for connections with peers.
	linkWaiters map[peer.ID][]*linkWaiter
	// linkDialers tracks ongoing dial attempts
	linkDialers map[linkDialerKey]*linkDialer
	// staticPeerMap maps a peer ID to a peermap.DialPeer
	// when EstablishLink matches a peer ID in this map,
	// the transport controller will dial the peer.
	staticPeerMap map[string]*dialer.DialerOpts
}

// NewController constructs a new transport controller.
func NewController(
	le *logrus.Entry,
	bus bus.Bus,
	info *controller.Info,
	nodePeerIDConstraint peer.ID,
	ctor Constructor,
	staticPeerMap map[string]*dialer.DialerOpts,
) *Controller {
	return &Controller{
		le:   le,
		bus:  bus,
		ctor: ctor,
		info: info,

		links:       make(map[uint64]*establishedLink),
		linkWaiters: make(map[peer.ID][]*linkWaiter),

		tptCtr: ccontainer.NewCContainer[*transport.Transport](nil),

		peerIDConstraint: nodePeerIDConstraint,
		localPeerID:      nodePeerIDConstraint,
		linkDialers:      make(map[linkDialerKey]*linkDialer),

		staticPeerMap: staticPeerMap,
	}
}

// GetControllerID returns the controller ID.
func (c *Controller) GetControllerID() string {
	return c.info.GetId()
}

// GetControllerInfo returns information about the controller.
func (c *Controller) GetControllerInfo() *controller.Info {
	return c.info.Clone()
}

// SetStaticPeerMap sets the static dialing peer map.
func (c *Controller) SetStaticPeerMap(m map[string]*dialer.DialerOpts) {
	c.mtx.Lock()
	c.staticPeerMap = m
	c.mtx.Unlock()
}

// PushStaticPeer pushes a static peer dialer.
func (c *Controller) PushStaticPeer(id string, opts *dialer.DialerOpts) {
	c.mtx.Lock()
	if c.staticPeerMap == nil {
		c.staticPeerMap = make(map[string]*dialer.DialerOpts)
	}
	c.staticPeerMap[id] = opts
	c.mtx.Unlock()
}

// GetPeerLinks returns all links with the peer.
func (c *Controller) GetPeerLinks(peerID peer.ID) []link.Link {
	var lnks []link.Link
	c.mtx.Lock()
	for _, lnk := range c.links {
		if lnk.lnk.GetRemotePeer() == peerID {
			lnks = append(lnks, lnk.lnk)
		}
	}
	c.mtx.Unlock()
	return lnks
}

// Execute executes the transport controller and the transport.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	c.ctx.Store(&ctx)
	// Acquire a handle to the node.
	c.le.
		WithField("peer-id", c.peerIDConstraint.Pretty()).
		Debug("looking up peer with ID")
	n, _, nRef, err := peer.GetPeerWithID(ctx, c.bus, c.peerIDConstraint, false, nil)
	if err != nil {
		return err
	}
	defer nRef.Release()

	// Get the priv key
	privKey, err := n.GetPrivKey(ctx)
	if err != nil {
		return err
	}
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

	c.le.Debug("executing transport")
	c.tptCtr.SetValue(&tpt)
	err = tpt.Execute(ctx)
	if err != nil {
		c.tptCtr.SetValue(nil)
	}
	return err
}

// GetTransport returns the controlled transport.
// This may be nil until the transport is constructed.
func (c *Controller) GetTransport(ctx context.Context) (transport.Transport, error) {
	tptv, err := c.tptCtr.WaitValue(ctx, nil)
	if err != nil {
		return nil, err
	}
	return *tptv, nil
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	dir := di.GetDirective()
	switch d := dir.(type) {
	case link.OpenStreamWithPeer:
		return c.resolveOpenStreamWithPeer(ctx, di, d)
	case link.OpenStreamViaLink:
		return c.resolveOpenStreamViaLink(ctx, di, d)
	case link.EstablishLinkWithPeer:
		return c.resolveEstablishLink(ctx, di, d)
	case transport.LookupTransport:
		return c.resolveLookupTransport(ctx, di, d)
	case tptaddr.DialTptAddr:
		return c.resolveDialTptAddr(ctx, di, d)
	}

	return nil, nil
}

// HandleLinkEstablished is called by the transport when a link is established.
func (c *Controller) HandleLinkEstablished(lnk link.Link) {
	le := c.loggerForLink(lnk)
	c.mtx.Lock()
	defer c.mtx.Unlock()

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
		if el.lnk == lnk {
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
	ctxPtr := c.ctx.Load()
	if ctxPtr == nil {
		c.le.Warn("dropping link: handle established link called before execute")
		go lnk.Close()
		return
	}

	el, err := newEstablishedLink(c.le, *c.ctx.Load(), c.bus, lnk, c)
	if err != nil {
		c.le.WithError(err).Warn("unable to construct established link")
		go lnk.Close()
		return
	}
	c.links[luuid] = el
	le.Info("link established")

	// flush any relevant link waiters
	c.resolveLinkWaiters(el.lnk, true)
}

// HandleIncomingStream handles an incoming stream from a link. It negotiates
// the protocol for the stream, acquires a handler for the protocol, and hands
// the stream to the protocol handler, then returns. Uses the ctx for
// cancellation.
//
// rctx is the link Context, which is canceled when the link is closed.
func (c *Controller) HandleIncomingStream(
	rctx context.Context,
	lnk link.Link,
	strm stream.Stream,
	strmOpts stream.OpenOpts,
) {
	// Assert EstablishLink to keep the stream open during the header exchange.
	_, elRef, err := c.bus.AddDirective(
		link.NewEstablishLinkWithPeer(lnk.GetLocalPeer(), lnk.GetRemotePeer()),
		nil,
	)
	if err == nil {
		defer elRef.Release()
	}

	readDeadline := time.Now().Add(streamEstablishTimeout)
	ctx, ctxCancel := context.WithDeadline(rctx, readDeadline)
	defer ctxCancel()
	_ = strm.SetReadDeadline(readDeadline)

	// process stream establish header;
	streamEst, err := readStreamEstablishHeader(strm)
	if err != nil {
		if err != io.EOF && err != context.Canceled {
			c.le.WithError(err).Warn("unable to read stream establish header")
		}
		strm.Close()
		return
	}
	_ = strm.SetDeadline(time.Time{})

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

	// bus is the controller bus
	le := c.loggerForLink(lnk).WithField("protocol-id", pid)
	le.
		WithField("stream-reliable", strmOpts.Reliable).
		WithField("stream-encrypted", strmOpts.Encrypted).
		WithField("remote-peer", lnk.GetRemotePeer().Pretty()).
		Debug("accepted stream")
	dir := link.NewHandleMountedStream(pid, c.localPeerID, mstrm.GetPeerID())
	dval, _, dref, err := bus.ExecOneOff(ctx, c.bus, dir, false, nil)
	if err != nil {
		le.WithError(err).Warn("error retrieving stream handler for stream")
		strm.Close()
		return
	}
	defer dref.Release()

	mhnd, ok := dval.GetValue().(link.MountedStreamHandler)
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
	c.mtx.Lock()
	defer c.mtx.Unlock()

	// fast path: clear by uuid
	luuid := lnk.GetUUID()
	if el, elOk := c.links[luuid]; elOk {
		delete(c.links, luuid)
		c.flushEstablishedLink(el)
		return
	}

	// slow path: equality check to be sure
	for k, l := range c.links {
		if l.lnk == lnk {
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
	tpt, err := c.GetTransport(ctx)
	if err != nil {
		return err
	}
	tptDialer, ok := tpt.(transport.TransportDialer)
	if !ok {
		c.le.Warn("ignoring dial attempt: transport is not a TransportDialer")
		return nil
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
	ctxPtr := c.ctx.Load()
	if ctxPtr == nil {
		return
	}
	ctx := *ctxPtr
	c.mtx.Lock()
	_, ok := c.linkDialers[key]
	if !ok {
		dialer := dialer.NewDialer(c.le, tptDialer, opts, peerID, key.dialAddress)
		ctx, ctxCancel := context.WithCancel(ctx)
		ld := &linkDialer{
			dialer: dialer,
			cancel: ctxCancel,
		}
		c.linkDialers[key] = ld
		go c.executeDialer(ctx, key, ld)
	}
	c.mtx.Unlock()
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

	le := c.loggerForLink(el.lnk)
	le.Info("link lost/closed")
	c.resolveLinkWaiters(el.lnk, false)
	el.cancel()
	el.lnk.Close()
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
	_ = c.tptCtr.SwapValue(func(val *transport.Transport) *transport.Transport {
		if val != nil {
			_ = (*val).Close()
		}
		return nil
	})
	return nil
}

// _ is a type assertion
var _ transport.Controller = ((*Controller)(nil))
