package main

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/aperturerobotics/bifrost/handshake/identity/s2s"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	s1Priv, s1Pub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		log.Fatal(err.Error())
	}

	s2Priv, s2Pub, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		log.Fatal(err.Error())
	}

	var s1, s2 *s2s.Handshaker
	s1w := func(data []byte) error {
		go s2.Handle(data)
		return nil
	}
	s2w := func(data []byte) error {
		go s1.Handle(data)
		return nil
	}

	s1, err = s2s.NewHandshaker(s1Priv, nil, s1w, nil, true, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	s2, err = s2s.NewHandshaker(s2Priv, nil, s2w, nil, false, nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	t1 := time.Now()
	ctx := context.Background()
	go func() {
		r, err := s1.Execute(ctx)
		if err != nil {
			log.Fatal(err.Error())
		}

		if !r.Peer.Equals(s2Pub) {
			log.Fatalf("s1 remote pub mismatch")
		}

		le.Info("s1 complete")
	}()

	r2, err := s2.Execute(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}

	if !r2.Peer.Equals(s1Pub) {
		log.Fatalf("s2 remote pub mismatch")
	}

	t2 := time.Now()
	le.Info("s2 complete")
	le.Debugf("timing: %s", t2.Sub(t1).String())
}
