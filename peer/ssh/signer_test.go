package peer_ssh

import (
	"testing"

	"github.com/aperturerobotics/bifrost/peer"
)

// TestSshKeys tests generating a key and converting to a ssh signer and pub key.
func TestSshKeys(t *testing.T) {
	// generate peer
	p, err := peer.NewPeer(nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	// create signer
	_, err = NewSigner(p.GetPrivKey())
	if err != nil {
		t.Fatal(err.Error())
	}

	// create pub key
	_, err = NewPublicKey(p.GetPubKey())
	if err != nil {
		t.Fatal(err.Error())
	}
}
