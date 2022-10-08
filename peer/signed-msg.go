package peer

import (
	"bytes"
	"encoding/hex"

	"github.com/aperturerobotics/bifrost/hash"
	crypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/pkg/errors"
	"github.com/zeebo/blake3"
	"google.golang.org/protobuf/proto"
)

// NewSignedMsg constructs/signs/encodes a new signed message.
func NewSignedMsg(
	privKey crypto.PrivKey,
	hashType hash.HashType,
	innerData []byte,
) (*SignedMsg, error) {
	peerID, err := IDFromPrivateKey(privKey)
	if err != nil {
		return nil, err
	}
	msg := &SignedMsg{
		FromPeerId: IDB58Encode(peerID),
		Data:       innerData,
	}
	if err := msg.Sign(privKey, hashType); err != nil {
		return nil, err
	}
	return msg, nil
}

// UnmarshalSignedMsg parses a signed message.
func UnmarshalSignedMsg(data []byte) (*SignedMsg, error) {
	m := &SignedMsg{}
	err := proto.Unmarshal(data, m)
	if err == nil {
		err = m.Validate()
	}
	if err != nil {
		return nil, err
	}
	return m, err
}

// ComputeMessageID computes a message id for a signed message.
func (m *SignedMsg) ComputeMessageID() string {
	inner := bytes.Join([][]byte{
		m.GetSignature().GetSigData(),
		[]byte(m.GetFromPeerId()),
	}, nil)
	h := blake3.Sum256(inner)
	return hex.EncodeToString(h[:])
}

// ParseFromPeerID unmarshals the peer id.
func (m *SignedMsg) ParseFromPeerID() (ID, error) {
	return IDB58Decode(m.GetFromPeerId())
}

// ExtractPubKey extracts the public key from the peer id.
func (m *SignedMsg) ExtractPubKey() (crypto.PubKey, ID, error) {
	fromPeerID, err := m.ParseFromPeerID()
	if err != nil {
		return nil, ID(""), err
	}
	pubKey, err := fromPeerID.ExtractPublicKey()
	if err != nil {
		return nil, fromPeerID, err
	}
	return pubKey, fromPeerID, nil
}

// ExtractAndVerify extracts public key & uses it to verify message
func (m *SignedMsg) ExtractAndVerify() (crypto.PubKey, ID, error) {
	pubKey, peerID, err := m.ExtractPubKey()
	if err != nil {
		return nil, peerID, err
	}

	sigErr := m.Verify(pubKey)
	if sigErr != nil {
		return pubKey, peerID, err
	}

	return pubKey, peerID, nil
}

// Sign signs the inner body with the private key.
// Disallows empty message.
func (m *SignedMsg) Sign(privKey crypto.PrivKey, hashType hash.HashType) error {
	innerData := m.GetData()
	if len(innerData) == 0 {
		return ErrBodyEmpty
	}

	sig, err := NewSignature(
		privKey,
		hashType,
		innerData,
		false,
	)
	if err != nil {
		return err
	}
	m.Signature = sig
	return nil
}

// Verify verifies the signature against a public key.
func (m *SignedMsg) Verify(pubKey crypto.PubKey) error {
	// validate signature
	sigOk, sigErr := m.GetSignature().VerifyWithPublic(pubKey, m.GetData())
	if !sigOk && sigErr == nil {
		sigErr = ErrSignatureInvalid
	}
	return sigErr
}

// Validate checks the signed message.
func (m *SignedMsg) Validate() error {
	if len(m.GetData()) == 0 {
		return ErrBodyEmpty
	}
	if len(m.GetFromPeerId()) == 0 {
		return ErrEmptyPeerID
	}
	if err := m.GetSignature().Validate(); err != nil {
		return errors.Wrap(err, "message signature")
	}
	_, id, err := m.ExtractAndVerify()
	if err != nil {
		return errors.Wrap(err, "message verify")
	}
	if err := id.Validate(); err != nil {
		return errors.Wrap(err, "message peer id")
	}
	return nil
}
