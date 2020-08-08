package pubmessage

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
)

// Message fulfills the Message type
type Message struct {
	pktInner *PubMessageInner
	peerID   peer.ID
}

// NewMessage constructs a new Message object
func NewMessage(peerID peer.ID, pktInner *PubMessageInner) *Message {
	return &Message{
		peerID:   peerID,
		pktInner: pktInner,
	}
}

// GetFrom returns the peer ID of the sender.
func (m *Message) GetFrom() peer.ID {
	return m.peerID
}

// GetAuthenticated indicates if the signature is valid.
func (m *Message) GetAuthenticated() bool {
	return true
}

// GetData returns the Message data.
func (m *Message) GetData() []byte {
	return m.pktInner.GetData()
}

// _ is a type assertion
var _ pubsub.Message = ((*Message)(nil))
