package main

import (
	"context"
	"crypto/rand"
	"net/http"

	"github.com/aperturerobotics/bifrost/peer"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var le = logrus.NewEntry(log)
var upgrader = websocket.Upgrader{}
var localPrivKey crypto.PrivKey
var localPeerID peer.ID

func init() {
	log.SetLevel(logrus.DebugLevel)
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
	localPeerID, localPrivKey = genPeerIdentity()
	le.WithField("local-peer-id", localPeerID.Pretty()).Debug("listening on :3000")

	tpt := wtpt.New(le, "", localPrivKey)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("../browser/")))
	mux.Handle("/ws/bifrost-0.1", tpt)

	go func() {
		err := tpt.Execute(context.Background())
		if err != nil {
			panic(err)
		}
	}()

	if err := http.ListenAndServe(":3000", mux); err != nil {
		panic(err)
	}
}
