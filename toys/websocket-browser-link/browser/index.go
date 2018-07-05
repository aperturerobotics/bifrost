//+build js
//go:generate gopherjs build -o index.js index.go

package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	wst "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/gopherjs/gopherjs/js"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var le = logrus.NewEntry(log)
var maxPacketSize = 1500

func init() {
	log.Formatter = &logrus.TextFormatter{
		DisableColors: true,
	}
	log.SetLevel(logrus.DebugLevel)
}

func getWSBaseURL() string {
	document := js.Global.Get("window").Get("document")
	location := document.Get("location")

	wsProtocol := "ws"
	if location.Get("protocol").String() == "https:" {
		wsProtocol = "wss"
	}

	return fmt.Sprintf("%s://%s:%s/ws/", wsProtocol, location.Get("hostname"), location.Get("port"))
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
	wsBaseURL := getWSBaseURL()
	peerID, peerPrivKey := genPeerIdentity()

	le.
		WithField("base-url", wsBaseURL).
		WithField("local-peer-id", peerID.Pretty()).
		Debug("contacting websocket peer")

	tpt := wst.New(le, peerPrivKey)
	go tpt.Execute(context.Background(), nil)
	if err := tpt.Dial(context.Background(), wsBaseURL+"bifrost-0.1"); err != nil {
		le.WithError(err).Warn("unable to start handshake")
		return
	}

	time.Sleep(time.Second * 5000)
}
