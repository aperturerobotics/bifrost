package node

import (
	"context"
	"crypto/rand"
	"sync"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// Node is a full routing peer in the network.
// Tracks transports, links, and updates the graph database.
// Has its own identity.
// It can connect to agents over IPC and receive routing claims for them.
type Node struct {
	// le is the root logger
	le *logrus.Entry
	// ctx is the node context
	ctx context.Context
	// privKey is the node private key
	privKey crypto.PrivKey
	// peerID is the local peer id
	peerID peer.ID

	// transportFactoriesMtx locks transportFactories
	transportFactoriesMtx sync.Mutex
	// transportFactories contains all running transport factories
	transportFactories map[transport.Factory]context.CancelFunc
}

// NewNode constructs a node.
// If privKey is nil, one will be generated.
// TODO: transport configs, graph database (?)
func NewNode(ctx context.Context, le *logrus.Entry, privKey crypto.PrivKey) (*Node, error) {
	var err error
	if privKey == nil {
		privKey, _, err = crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return nil, err
		}
	}

	n := &Node{ctx: ctx, le: le, privKey: privKey}
	n.peerID, err = peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// GetPubKey returns the public key of the node.
func (n *Node) GetPubKey() crypto.PubKey {
	return n.privKey.GetPublic()
}

// GetPrivKey returns the private key.
func (n *Node) GetPrivKey() crypto.PrivKey {
	return n.privKey
}

// GetPeerID returns the peer ID.
func (n *Node) GetPeerID() peer.ID {
	return n.peerID
}

// AddTransportFactory adds a transport factory to the node.
func (n *Node) AddTransportFactory(f transport.Factory) {
	n.transportFactoriesMtx.Lock()
	defer n.transportFactoriesMtx.Unlock()

	if _, ok := n.transportFactories[f]; ok {
		return
	}

	ctx, ctxCancel := context.WithCancel(n.ctx)
	n.transportFactories[f] = ctxCancel
}

// AddLink adds a new link for processing.
func (n *Node) AddLink(l link.Link) {
	n.le.Debug("link added")
}

// AddTransport adds a new transport for processing.
func (n *Node) AddTransport(t transport.Transport) {
	n.le.Debug("transport added")
}
