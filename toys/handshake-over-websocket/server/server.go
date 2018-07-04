package main

import (
	"context"
	"crypto/rand"
	"net/http"

	"github.com/aperturerobotics/bifrost/handshake/identity/s2s"
	"github.com/aperturerobotics/bifrost/peer"
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

func handshake(rw http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		le.WithError(err).Warn("unable to upgrade ws conn")
		return
	}
	defer conn.Close()

	hs, err := s2s.NewHandshaker(
		localPrivKey,
		nil,
		func(data []byte) error {
			return conn.WriteMessage(websocket.BinaryMessage, data)
		},
		nil,
		nil,
	)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				log.WithError(err).Warn("read message error")
				return
			}

			log.
				WithField("message-type", mt).
				WithField("message-len", len(message)).
				Debug("got message")
			if mt != websocket.BinaryMessage {
				continue
			}

			hs.Handle(message)
		}
	}()

	res, err := hs.Execute(context.Background(), false)
	if err != nil {
		panic(err)
	}

	remotePeerID, _ := peer.IDFromPublicKey(res.Peer)
	le.
		WithField("remote-peer", remotePeerID.Pretty()).
		Info("handshake complete")
}

func main() {
	localPeerID, localPrivKey = genPeerIdentity()
	le.WithField("local-peer-id", localPeerID.Pretty()).Debug("listening on :3000")

	// Serve test folder.
	http.Handle("/", http.FileServer(http.Dir("../browser/")))

	http.HandleFunc("/ws/handshake", handshake)

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		panic(err)
	}
}
