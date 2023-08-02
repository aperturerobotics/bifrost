package pubmessage

import (
	"github.com/aperturerobotics/bifrost/hash"
	"github.com/aperturerobotics/bifrost/peer"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	proto "google.golang.org/protobuf/proto"
	timestamp "google.golang.org/protobuf/types/known/timestamppb"

	// _ ensures the typescript version of google/protobuf/timestamp exists in vendor/.
	_ "github.com/aperturerobotics/ts-proto-common-types/google/protobuf"
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
	innerData, err := proto.Marshal(inner)
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
	err = proto.Unmarshal(data, out)
	if err == nil {
		err = out.Validate()
	}
	if err != nil {
		return nil, pubKey, peerID, err
	}
	return out, pubKey, peerID, nil
}
