package inproc

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/net/peer"
)

// Network connects transports in one explicitly shared process-local network.
// It supplies packet reachability; transport handshakes still authenticate peers.
// Each attached transport must be detached before its lifecycle ends.
type Network struct {
	// mtx serializes topology changes and guards peers.
	mtx sync.Mutex
	// peers contains the currently attached transport for each identity.
	peers map[peer.ID]*Inproc
}

// NewNetwork constructs an isolated network with no attached transports.
func NewNetwork() *Network { return &Network{peers: make(map[peer.ID]*Inproc)} }

// Attach connects a transport to the network and returns its idempotent detach.
// The same peer cannot be attached twice; replacement first releases its old attachment.
func (n *Network) Attach(ctx context.Context, transport *Inproc) (func(), error) {
	// Serialize both directions of every connection against other topology changes.
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if transport == nil {
		return nil, errors.New("in-process network requires a transport")
	}
	n.mtx.Lock()
	defer n.mtx.Unlock()
	id := transport.GetPeerID()
	if n.peers[id] != nil {
		return nil, errors.New("peer is already attached to the in-process network")
	}
	for _, other := range n.peers {
		transport.ConnectToInproc(ctx, other)
		other.ConnectToInproc(ctx, transport)
	}
	n.peers[id] = transport
	attached := true
	return func() {
		n.mtx.Lock()
		defer n.mtx.Unlock()
		if !attached {
			return
		}
		attached = false
		delete(n.peers, id)
		for _, other := range n.peers {
			transport.DisconnectInproc(ctx, other)
			other.DisconnectInproc(ctx, transport)
		}
	}, nil
}
