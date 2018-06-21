package node

import (
	"crypto/rand"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/libp2p/go-libp2p-crypto"
)

// Node is a full routing peer in the network.
// Tracks transports, links, and updates the graph database.
// Has its own identity.
// It can connect to agents over IPC and receive routing claims for them.
type Node struct {
	// privKey is the node private key
	privKey crypto.PrivKey
	// peerID is the local peer id
	peerID peer.ID
}

// NewNode constructs a node.
// If privKey is nil, one will be generated.
// TODO: transport configs, graph database (?)
func NewNode(privKey crypto.PrivKey) (*Node, error) {
	var err error
	if privKey == nil {
		privKey, _, err = crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			return nil, err
		}
	}

	n := &Node{privKey: privKey}
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
