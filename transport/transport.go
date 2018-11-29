package transport

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/controllerbus/controller"
)

// Transport is similar to a NIC, yielding links to remote peers.
type Transport interface {
	// Execute executes the transport as configured, returning any fatal error.
	Execute(ctx context.Context) error
	// GetUUID returns a host-unique ID for this transport.
	GetUUID() uint64
	// GetPeerID returns the peer ID.
	GetPeerID() peer.ID
	// Close closes the transport, returning any errors closing.
	Close() error
}

// TransportDialer is a transport that supports dialing string-serialized remote
// addresses. The Transport controller will call Dial if provided an address for
// the transport, and directed to connect to the peer.
type TransportDialer interface {
	// DialPeer dials a peer given an address. The yielded link should be
	// emitted to the transport handler. DialPeer should return nil if the link
	// was established. DialPeer will then not be called again for the same peer
	// ID and address tuple until the yielded link is lost.
	// Returns fatal and error.
	DialPeer(
		ctx context.Context,
		peerID peer.ID,
		addr string,
	) (fatal bool, err error)
}

// TransportHandler manages a Transport and receives event callbacks.
// This is typically fulfilled by the transport controller.
type TransportHandler interface {
	// HandleLinkEstablished is called when a link is established.
	HandleLinkEstablished(lnk link.Link)
	// HandleLinkLost is called when a link is lost.
	HandleLinkLost(lnk link.Link)
}

// Controller is a transport controller.
type Controller interface {
	// Controller is the controllerbus controller interface.
	controller.Controller

	// GetTransport returns the controlled transport.
	// This may wait for the controller to be ready.
	GetTransport(ctx context.Context) (Transport, error)
}
