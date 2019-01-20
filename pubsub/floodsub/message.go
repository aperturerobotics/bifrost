package floodsub

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
)

// message fulfills the message type
type message struct {
	pktInner *PubMessageInner
	peerID   peer.ID
}

// newMessage constructs a new message object
func newMessage(peerID peer.ID, pktInner *PubMessageInner) *message {
	return &message{
		peerID:   peerID,
		pktInner: pktInner,
	}
}

// GetFrom returns the peer ID of the sender.
func (m *message) GetFrom() peer.ID {
	return m.peerID
}

// GetAuthenticated indicates if the signature is valid.
func (m *message) GetAuthenticated() bool {
	return true
}

// GetData returns the message data.
func (m *message) GetData() []byte {
	return m.pktInner.GetData()
}

// _ is a type assertion
var _ pubsub.Message = ((*message)(nil))
