package s2s

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

// TestS2S tests S2S end-to-end.
func TestS2S(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	s1Priv, s1Pub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err.Error())
	}

	s2Priv, s2Pub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		t.Fatal(err.Error())
	}

	var s1, s2 *Handshaker
	s1w := func(data []byte) error {
		go s2.Handle(data)
		return nil
	}
	s2w := func(data []byte) error {
		go s1.Handle(data)
		return nil
	}

	s1, err = NewHandshaker(s1Priv, nil, s1w, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	s2, err = NewHandshaker(s2Priv, nil, s2w, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	ctx := context.Background()
	go func() {
		r, err := s1.Execute(ctx, true)
		if err != nil {
			t.Fatal(err.Error())
		}

		if !r.Peer.Equals(s2Pub) {
			t.Fatalf("s1 remote pub mismatch")
		}

		le.Info("s1 complete")
	}()

	r2, err := s2.Execute(ctx, false)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !r2.Peer.Equals(s1Pub) {
		t.Fatalf("s2 remote pub mismatch")
	}

	le.Info("s2 complete")
}
