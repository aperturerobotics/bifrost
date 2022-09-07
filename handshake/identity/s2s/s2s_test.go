package s2s

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
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

	ex1 := make([]byte, 50)
	ex2 := make([]byte, 50)
	if _, err := rand.Read(ex1); err != nil {
		t.Fatal(err.Error())
	}
	if _, err := rand.Read(ex2); err != nil {
		t.Fatal(err.Error())
	}

	s1, err = NewHandshaker(s1Priv, nil, s1w, nil, true, ex1)
	if err != nil {
		t.Fatal(err.Error())
	}

	s2, err = NewHandshaker(s2Priv, nil, s2w, nil, false, ex2)
	if err != nil {
		t.Fatal(err.Error())
	}

	ctx := context.Background()
	errCh := make(chan error, 1)
	go func() {
		r, err := s1.Execute(ctx)
		if err != nil {
			errCh <- err
			return
		}

		if !r.Peer.Equals(s2Pub) {
			errCh <- errors.New("s1 remote pub mismatch")
			return
		}

		if !bytes.Equal(ex2, r.ExtraData) {
			errCh <- errors.New("s1 remote extradata mismatch")
			return
		}

		le.Info("s1 complete")
		errCh <- nil
	}()

	r2, err := s2.Execute(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !r2.Peer.Equals(s1Pub) {
		t.Fatalf("s2 remote pub mismatch")
	}

	if !bytes.Equal(ex1, r2.ExtraData) {
		t.Fatalf("s1 remote extradata mismatch")
	}

	le.Info("s2 complete")
	if err := <-errCh; err != nil {
		t.Fatal(err.Error())
	}
}
