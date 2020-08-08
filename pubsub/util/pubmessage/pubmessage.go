package pubmessage

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"math/rand"

	"github.com/aperturerobotics/bifrost/hash"
	"github.com/aperturerobotics/bifrost/peer"
	timestamp "github.com/aperturerobotics/timestamp"
	proto "github.com/golang/protobuf/proto"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
)

// NewPubMessage constructs/signs/encodes a new pub-message and inner message.
func NewPubMessage(
	channelID string,
	privKey crypto.PrivKey,
	hashType hash.HashType,
	data []byte,
) (*PubMessage, *PubMessageInner, error) {
	pid, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return nil, nil, err
	}

	tsNow := timestamp.Now()
	inner := &PubMessageInner{
		Data:      data,
		Channel:   channelID,
		Timestamp: &tsNow,
		Salt:      rand.Uint32(),
	}
	innerData, err := proto.Marshal(inner)
	if err != nil {
		return nil, inner, err
	}

	sig, err := peer.NewSignature(
		privKey,
		hashType,
		innerData,
		false,
	)
	if err != nil {
		return nil, inner, err
	}

	return &PubMessage{
		FromPeerId: peer.IDB58Encode(pid),
		Signature:  sig,
		Data:       innerData,
	}, inner, nil
}

// ComputeMessageID computes a message id for a packet
func (m *PubMessage) ComputeMessageID() string {
	inner := bytes.Join([][]byte{
		m.GetSignature().GetSigData(),
		[]byte(m.GetFromPeerId()),
	}, nil)
	s := sha1.Sum(inner)
	return hex.EncodeToString(s[:])
}

// ParseFromPeerID unmarshals the peer id.
func (m *PubMessage) ParseFromPeerID() (peer.ID, error) {
	return peer.IDB58Decode(m.GetFromPeerId())
}

// ExtractPubKey extracts the public key from the peer id.
func (m *PubMessage) ExtractPubKey() (crypto.PubKey, peer.ID, error) {
	fromPeerID, err := m.ParseFromPeerID()
	if err != nil {
		return nil, peer.ID(""), err
	}
	pubKey, err := fromPeerID.ExtractPublicKey()
	if err != nil {
		return nil, fromPeerID, err
	}
	return pubKey, fromPeerID, nil
}

// ExtractAndVerify extracts public key & uses it to verify message
func (m *PubMessage) ExtractAndVerify() (*PubMessageInner, crypto.PubKey, peer.ID, error) {
	pubKey, peerID, err := m.ExtractPubKey()
	if err != nil {
		return nil, nil, peerID, err
	}

	sigErr := m.Verify(pubKey)
	if sigErr != nil {
		return nil, pubKey, peerID, err
	}

	pktInner := &PubMessageInner{}
	if err := proto.Unmarshal(m.GetData(), pktInner); err != nil {
		return nil, pubKey, peerID, err
	}

	chid := pktInner.GetChannel()
	if chid == "" {
		return nil, pubKey, peerID, ErrInvalidChannelID
	}
	return pktInner, pubKey, peerID, nil
}

// Verify verifies the signature against a public key.
func (m *PubMessage) Verify(pubKey crypto.PubKey) error {
	// validate signature
	sigOk, sigErr := m.GetSignature().VerifyWithPublic(pubKey, m.GetData())
	if !sigOk && sigErr == nil {
		sigErr = ErrInvalidSignature
	}
	return sigErr
}
