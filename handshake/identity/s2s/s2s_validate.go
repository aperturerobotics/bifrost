package s2s

import (
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/pkg/errors"
	"golang.org/x/crypto/nacl/box"
)

// Validate validates the message.
func (m *Packet_Init) Validate() error {
	if len(m.GetSenderPeerId()) == 0 {
		return errors.Wrap(peer.ErrPeerIDEmpty, "sender")
	}

	if len(m.GetSenderEphPub()) != 32 {
		return errors.New("invalid sender eph pub")
	}

	return nil
}

// Validate validates the init ack packet.
func (m *Packet_InitAck) Validate() error {
	if len(m.GetSenderEphPub()) != 32 {
		return errors.New("invalid sender eph pub")
	}

	if len(m.GetCiphertext()) <= box.Overhead {
		return errors.New("ciphertext less than overhead, too short")
	}

	return nil
}
