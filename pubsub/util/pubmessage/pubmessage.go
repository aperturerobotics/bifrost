package pubmessage

import (
	"github.com/aperturerobotics/bifrost/hash"
	"github.com/aperturerobotics/bifrost/peer"
	timestamp "github.com/aperturerobotics/protobuf-go-lite/types/known/timestamppb"
	"github.com/aperturerobotics/bifrost/crypto"
)

const pubMessageEncContext = "bifrost/pubsub/pubmessage 2024-06-05T02:38:47.55258Z channel/"

// NewPubMessage constructs/signs/encodes a new pub-message and inner message.
//
// Uses time.Now() for the timestamp: not deterministic.
func NewPubMessage(
	channelID string,
	privKey crypto.PrivKey,
	hashType hash.HashType,
	data []byte,
) (*peer.SignedMsg, *PubMessageInner, error) {
	inner := &PubMessageInner{
		Data:      data,
		Channel:   channelID,
		Timestamp: timestamp.Now(),
	}
	innerData, err := inner.MarshalVT()
	if err != nil {
		return nil, inner, err
	}

	sig, err := peer.NewSignedMsg(pubMessageEncContext+channelID, privKey, hashType, innerData)
	return sig, inner, err
}

// ExtractAndVerify extracts the inner message from a signed message.
func ExtractAndVerify(msg *peer.SignedMsg) (*PubMessageInner, crypto.PubKey, peer.ID, error) {
	data := msg.GetData()

	out := &PubMessageInner{}
	err := out.UnmarshalVT(data)
	if err == nil {
		err = out.Validate()
	}
	if err != nil {
		return nil, nil, "", err
	}

	pubKey, peerID, err := msg.ExtractAndVerify(pubMessageEncContext + out.GetChannel())
	if err != nil {
		return nil, nil, "", err
	}
	return out, pubKey, peerID, nil
}
