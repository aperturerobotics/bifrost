package pubmessage

import (
	"github.com/aperturerobotics/bifrost/hash"
	"github.com/aperturerobotics/bifrost/peer"
	timestamp "github.com/aperturerobotics/protobuf-go-lite/types/known/timestamppb"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
)

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

	sig, err := peer.NewSignedMsg(privKey, hashType, innerData)
	return sig, inner, err
}

// ExtractAndVerify extracts the inner message from a signed message.
func ExtractAndVerify(msg *peer.SignedMsg) (*PubMessageInner, crypto.PubKey, peer.ID, error) {
	pubKey, peerID, err := msg.ExtractAndVerify()
	if err != nil {
		return nil, nil, "", err
	}
	data := msg.GetData()
	out := &PubMessageInner{}
	err = out.UnmarshalVT(data)
	if err == nil {
		err = out.Validate()
	}
	if err != nil {
		return nil, pubKey, peerID, err
	}
	return out, pubKey, peerID, nil
}
