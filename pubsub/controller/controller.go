package pubsub_controller

import (
	"context"
	"sync"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// Constructor constructs a PubSub with common parameters.
type Constructor func(
	ctx context.Context,
	le *logrus.Entry,
	peer peer.Peer,
	handler pubsub.PubSubHandler,
) (pubsub.PubSub, error)

// Controller implements a common PubSub controller. The controller monitors
// active links, and pushes peer ID <-> link UUID tuples to the router. It
// handles PubSub event callbacks and PubSub related directives, yielding PubSub
// client handles and managing active topics.
type Controller struct {
	// le is the logger
	le *logrus.Entry
	// bus is the controller bus
	bus bus.Bus
	// ctor is the constructor
	ctor Constructor
	// info is the controller info
	info *controller.Info
	// protocolID is the protocol id
	protocolID protocol.ID
	// peerID is the peer ID to use
	peerID peer.ID

	// pubSubCh holds the PubSub like a bucket
	pubSubCh chan pubsub.PubSub
	// peerCh holds the Peer like a bucket
	peerCh chan peer.Peer
	// wakeCh wakes the controller
	wakeCh chan struct{}

	mtx sync.Mutex
	// cleanupRefs are the refs to cleanup
	cleanupRefs []directive.Reference
	// incLinks are links that need to be established
	incLinks []link.Link
	// links are tracked links
	links map[pubsub.PeerLinkTuple]*trackedLink
}

// NewController constructs a new transport controller.
func NewController(
	le *logrus.Entry,
	bus bus.Bus,
	controllerInfo *controller.Info,
	peerID peer.ID,
	protocolID protocol.ID,
	ctor Constructor,
) *Controller {
	return &Controller{
		le:     le,
		bus:    bus,
		ctor:   ctor,
		info:   controllerInfo,
		peerID: peerID,

		protocolID: protocolID,
		pubSubCh:   make(chan pubsub.PubSub, 1),
		wakeCh:     make(chan struct{}, 1),
		peerCh:     make(chan peer.Peer, 1),
		links:      make(map[pubsub.PeerLinkTuple]*trackedLink),
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

// Execute executes the transport controller and the transport.
// Returning nil ends execution.
// Returning an error triggers a retry with backoff.
func (c *Controller) Execute(ctx context.Context) error {
	// Fetch the peer if the peer ID is set.
	var cpeer peer.Peer
	var err error
	if len(c.peerID) != 0 {
		// special value: "any"
		if c.peerID == peer.ID("any") {
			c.peerID = peer.ID("")
		}

		var ref directive.Reference
		cpeer, ref, err = peer.GetPeerWithID(ctx, c.bus, c.peerID)
		if err != nil {
			return err
		}
		if ref != nil {
			defer ref.Release()
		}
		c.peerID = cpeer.GetPeerID()
	}

	c.peerCh <- cpeer
	defer func() {
		<-c.peerCh
	}()

	// Construct the PubSub
	ps, err := c.ctor(
		ctx,
		c.le,
		cpeer,
		c,
	)
	if err != nil {
		return err
	}
	defer ps.Close()
	c.pubSubCh <- ps

	psErr := make(chan error, 1)
	go func() {
		c.le.Debug("executing pubsub")
		defer ps.Close()
		if err := ps.Execute(ctx); err != nil {
			psErr <- err
		}
	}()

	for {
		c.mtx.Lock()
		for _, vl := range c.incLinks {
			tpl := pubsub.NewPeerLinkTuple(vl)
			if e, ok := c.links[tpl]; ok {
				e.ctxCancel()
			}
			tlCtx, tlCtxCancel := context.WithCancel(ctx)
			tl := &trackedLink{
				c:         c,
				ctxCancel: tlCtxCancel,
				tpl:       tpl,
				lnk:       vl,
				le: c.le.
					WithField("link-uuid", vl.GetUUID()).
					WithField("link-remote-peer", vl.GetRemotePeer().Pretty()),
			}
			c.links[tpl] = tl
			go func() {
				err := tl.trackLink(tlCtx)
				tlCtxCancel()
				if err != context.Canceled && err != nil {
					tl.le.WithError(err).Warn("link tracker returned fatal error")
				}
				c.mtx.Lock()
				if ol, ok := c.links[tpl]; ok && ol == tl {
					delete(c.links, tpl)
				}
				c.mtx.Unlock()
			}()
		}
		c.incLinks = nil
		c.mtx.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-psErr:
			return err
		case <-c.wakeCh:
		}
	}
}

// GetPubSub returns the controlled PubSub.
func (c *Controller) GetPubSub(ctx context.Context) (pubsub.PubSub, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case tpt := <-c.pubSubCh:
		c.pubSubCh <- tpt
		return tpt, nil
	}
}

// GetPeer returns the controlled peer ID.
func (c *Controller) GetPeer(ctx context.Context) (peer.Peer, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case tpt := <-c.peerCh:
		c.peerCh <- tpt
		return tpt, nil
	}
}

// HandleDirective asks if the handler can resolve the directive.
// If it can, it returns a resolver. If not, returns nil.
// Any exceptional errors are returned for logging.
// It is safe to add a reference to the directive during this call.
func (c *Controller) HandleDirective(ctx context.Context, di directive.Instance) ([]directive.Resolver, error) {
	dir := di.GetDirective()
	switch d := dir.(type) {
	case link.EstablishLinkWithPeer:
		c.handleEstablishLink(ctx, di, d)
	case link.HandleMountedStream:
		return c.handleMountedStream(ctx, di, d)
	case pubsub.BuildChannelSubscription:
		return c.resolveBuildChannelSub(ctx, di, d)
	}

	return nil, nil
}

// Close releases any resources used by the controller.
// Error indicates any issue encountered releasing.
func (c *Controller) Close() error {
	// nil references to help GC along
	c.ctor = nil
	c.le = nil
	c.bus = nil

	c.mtx.Lock()
	for _, ref := range c.cleanupRefs {
		ref.Release()
	}
	c.cleanupRefs = nil
	for k, l := range c.links {
		l.ctxCancel()
		delete(c.links, k)
	}
	c.mtx.Unlock()

	select {
	case ps := <-c.pubSubCh:
		ps.Close()
	default:
	}

	return nil
}

// wake wakes the controller
func (c *Controller) wake() {
	select {
	case c.wakeCh <- struct{}{}:
	default:
	}
}

// _ is a type assertion
var _ pubsub.Controller = ((*Controller)(nil))
