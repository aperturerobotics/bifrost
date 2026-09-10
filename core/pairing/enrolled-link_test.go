package pairing

import (
	"crypto/rand"
	"testing"

	"github.com/s4wave/spacewave/core/provider"
	"github.com/s4wave/spacewave/core/session"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/peer"
)

// TestEnrolledLinkIdentityBinding proves that the independently enrolled key
// can replace its setup identity only on the connection named by both proofs.
func TestEnrolledLinkIdentityBinding(t *testing.T) {
	source, receiving := newEngineTestPeer(t), newEngineTestPeer(t)
	key, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	storage, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	enrolled, err := peer.IDFromPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	offer := &AccountOffer{ProviderId: "local", AccountId: "account", OperationId: "operation"}
	ref := &session.SessionRef{ProviderResourceRef: &provider.ProviderResourceRef{ProviderId: "local", ProviderAccountId: "account", Id: "session"}}
	identity, err := BuildIdentity(offer, ref, key, storage, source, receiving)
	if err != nil {
		t.Fatal(err)
	}
	for _, localSource := range []bool{true, false} {
		connection := &enrolledLink{local: source, remote: receiving}
		wantLocal, wantRemote := source, enrolled
		if !localSource {
			connection.local, connection.remote = receiving, source
			wantLocal, wantRemote = enrolled, source
		}
		bound, err := BindEnrolledLink(connection, offer, identity, source, receiving)
		if err != nil {
			t.Fatal(err)
		}
		if bound.GetLocalPeer() != wantLocal || bound.GetRemotePeer() != wantRemote {
			t.Fatal("enrollment mapped the wrong authenticated peer")
		}
		changed := offer.CloneVT()
		changed.OperationId = "another operation"
		if _, err := BindEnrolledLink(connection, changed, identity, source, receiving); err == nil {
			t.Fatal("accepted another pairing operation")
		}
		if _, err := BindEnrolledLink(connection, offer, identity, receiving, source); err == nil {
			t.Fatal("accepted reversed proof roles")
		}
		connection.remote = newEngineTestPeer(t)
		if _, err := BindEnrolledLink(connection, offer, identity, source, receiving); err == nil {
			t.Fatal("accepted a proof on another authenticated connection")
		}
	}
}
