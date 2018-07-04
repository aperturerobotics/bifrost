package node

import (
	"context"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// Node is a full routing peer in the network.
// Has its own identity.
// It can connect to agents over IPC and receive routing claims for them.
type Node struct {
	peer.Peer

	// le is the root logger
	le *logrus.Entry
	// ctx is the node context
	ctx context.Context
}

// NewNode constructs a node.
// If privKey is nil, one will be generated.
// TODO: transport configs, graph database (?)
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
	}, nil
}

// AddLink adds a new link for processing.
func (n *Node) AddLink(l link.Link) {
	n.le.Debug("link added")
}
