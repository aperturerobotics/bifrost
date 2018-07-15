package node

import (
	"context"
	"sync"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// Node is a full routing peer in the network.
// Has its own identity.
type Node struct {
	peer.Peer

	// le is the root logger
	le *logrus.Entry
	// ctx is the node context
	ctx context.Context

	// transportsMtx guards the transports map
	transportsMtx sync.Mutex
	// transports are the running transports
	transports map[uint64]transport.Transport
}

// NewNode constructs a node.
// If privKey is nil, one will be generated.
func NewNode(ctx context.Context, le *logrus.Entry, privKey crypto.PrivKey) (*Node, error) {
	var err error

	p, err := peer.NewPeer(privKey)
	if err != nil {
		return nil, err
	}

	return &Node{
		Peer: *p,
		ctx:  ctx,
		le:   le,

		transports: make(map[uint64]transport.Transport),
	}, nil
}

// AddTransport adds a transport to the node.
// The Node will call Execute() on the transport.
func (n *Node) AddTransport(tpt transport.Transport) {
	// First, kill any transports with the same uuid.
	uuid := tpt.GetUUID()
	n.transportsMtx.Lock()
	_, ok := n.transports[uuid]
	if ok {
		delete(n.transports, uuid)
	}
	n.transportsMtx.Unlock()

}

// removeTransport removes and cleans up a transport
func (n *Node) removeTransport(uuid uint64) {
	// TODO: remove from directive handler set?
}
