package pairing

import (
	"crypto/rand"
	"errors"
	"testing"

	"github.com/s4wave/spacewave/core/provider"
	"github.com/s4wave/spacewave/core/session"
	"github.com/s4wave/spacewave/net/crypto"
	"github.com/s4wave/spacewave/net/peer"
)

func TestResultRequiresDurableEnrollment(t *testing.T) {
	remote := newEngineTestPeer(t)
	engine := &Engine{}
	if _, err := engine.Result(remote); !errors.Is(err, ErrExchangeMissing) {
		t.Fatalf("missing exchange returned %v", err)
	}
	engine.active = &attempt{snapshot: Snapshot{RemotePeerID: remote, Status: StatusBothConfirmed}}
	if _, err := engine.Result(remote); !errors.Is(err, ErrExchangeUnconfirmed) {
		t.Fatalf("completed status without enrollment returned %v", err)
	}
	if _, err := engine.Result(newEngineTestPeer(t)); !errors.Is(err, ErrExchangePeerMismatch) {
		t.Fatalf("different peer returned %v", err)
	}
	ref := &session.SessionRef{ProviderResourceRef: &provider.ProviderResourceRef{ProviderId: "local", ProviderAccountId: "account", Id: "session"}}
	engine.active.result = ref
	result, err := engine.Result(remote)
	if err != nil {
		t.Fatal(err)
	}
	if !result.EqualVT(ref) || result == ref {
		t.Fatal("result did not return an independent copy of the enrolled Session")
	}
}

func newEngineTestPeer(t *testing.T) peer.ID {
	t.Helper()
	key, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	id, err := peer.IDFromPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
