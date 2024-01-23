package signaling_rpc

import (
	"errors"

	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/protocol"
	"github.com/aperturerobotics/bifrost/util/confparse"
)

// ProtocolID is the default protocol ID for the signaling server.
var ProtocolID = protocol.ID("bifrost/signaling")

// Validate validates the Listen request.
func (r *ListenRequest) Validate() error {
	return nil
}

// Validate validates the Session request.
func (r *SessionRequest) Validate() error {
	switch b := r.GetBody().(type) {
	case *SessionRequest_Init:
		if err := b.Init.Validate(); err != nil {
			return err
		}
	case *SessionRequest_AckMsg:
	case *SessionRequest_ClearMsg:
	case *SessionRequest_SendMsg:
		// verify the signed message & the signature.
		if err := b.SendMsg.Validate(); err != nil {
			return err
		}
	default:
		return errors.New("unknown session request message type")
	}
	return nil
}

// Validate validates the Session response.
//
// NOTE: this checks the signature but not which peer the signature is from.
// Be sure to additionally check the origin of the signature is what you expect.
func (r *SessionResponse) Validate() error {
	switch b := r.GetBody().(type) {
	case *SessionResponse_Opened:
	case *SessionResponse_Closed:
	case *SessionResponse_ClearMsg:
	case *SessionResponse_AckMsg:
	case *SessionResponse_RecvMsg:
		// verify the signed message & the signature.
		if err := b.RecvMsg.Validate(); err != nil {
			return err
		}
	default:
		return errors.New("unknown session response message type")
	}
	return nil
}

// Validate validates the SessionMsg.
func (m *SessionMsg) Validate() error {
	return m.GetSignedMsg().Validate()
}

// Validate validates the SessionInit.
func (i *SessionInit) Validate() error {
	pid, err := i.ParsePeerID()
	if err != nil {
		return err
	}
	if len(pid) == 0 {
		return peer.ErrEmptyPeerID
	}
	return nil
}

// ParsePeerID parses the peer ID.
func (s *SessionInit) ParsePeerID() (peer.ID, error) {
	return confparse.ParsePeerID(s.GetPeerId())
}
