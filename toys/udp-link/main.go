package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"time"
	// "net"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport/udp"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var le = logrus.NewEntry(log)

func init() {
	log.SetLevel(logrus.DebugLevel)
}

type handler struct {
}

// AddLink handles a link.
func (h *handler) AddLink(nk link.Link) {
	fmt.Printf("link built: %#v\n", nk)
}

func genPeerIdentity() (peer.ID, crypto.PrivKey) {
	pk1, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	pid1, _ := peer.IDFromPrivateKey(pk1)
	log.Debugf("generated peer id: %s", pid1.Pretty())

	return pid1, pk1
}

func main() {
	ctx := context.Background()
	pid1, pk1 := genPeerIdentity()
	pid2, pk2 := genPeerIdentity()

	utp1, err := udp.NewUDP(le.WithField("utp", 1), "127.0.0.1:0", pk1)
	if err != nil {
		log.Fatal(err)
	}
	defer utp1.Close()
	fmt.Printf("%s: listening on %s\n", pid1.Pretty(), utp1.LocalAddr().String())

	utp2, err := udp.NewUDP(le.WithField("utp", 2), "127.0.0.1:0", pk2)
	if err != nil {
		log.Fatal(err)
	}
	defer utp2.Close()
	fmt.Printf("%s: listening on %s\n", pid2.Pretty(), utp2.LocalAddr().String())

	h := &handler{}
	go func() {
		ctx2, ctx2Cancel := context.WithTimeout(ctx, time.Second*10)
		defer ctx2Cancel()
		if err := utp2.Execute(ctx2, h); err != nil {
			fmt.Println(err.Error())
			// os.Exit(1)
		}
	}()

	go func() {
		<-time.After(time.Millisecond * 500)
		utp2.Dial(ctx, utp1.LocalAddr())
	}()

	if err := utp1.Execute(ctx, h); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	/*

		b := make([]byte, 1024)
		_, addr, err := pc.ReadFrom(b)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("received packet from %v: %v\n", addr.String(), string(b))
	*/
}
