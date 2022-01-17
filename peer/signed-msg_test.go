package peer

import (
	"bytes"
	"testing"

	"github.com/aperturerobotics/bifrost/hash"
)

// TestSignedMsg tests signing a message.
func TestSignedMsg(t *testing.T) {
	p, err := NewPeer(nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	exPeerID, privKey := p.GetPeerID(), p.GetPrivKey()

	msg := "hello world from signed message test"
	smsg, err := NewSignedMsg(privKey, hash.RecommendedHashType, []byte(msg))
	if err == nil {
		err = smsg.Validate()
	}
	if err != nil {
		t.Fatal(err.Error())
	}

	_, peerID, err := smsg.ExtractAndVerify()
	if err != nil {
		t.Fatal(err.Error())
	}
	if peerID != exPeerID {
		t.Fatalf("peer id mismatch: %s != %s", exPeerID.Pretty(), peerID.Pretty())
	}

	if !bytes.Equal(smsg.Data, []byte(msg)) {
		t.FailNow()
	}
}
