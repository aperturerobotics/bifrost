//go:generate gopherjs build -o index.js index.go

package main

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/aperturerobotics/bifrost/handshake/identity/s2s"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var le = logrus.NewEntry(log)
var maxPacketSize = 1500

func init() {
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

	ws, err := websocket.Dial(wsBaseURL + "handshake")
	if err != nil {
		le.WithError(err).Warn("unable to dial websocket")
		return
	}

	le.Debug("contacted peer, starting handshake")
	hs, err := s2s.NewHandshaker(
		peerPrivKey, nil,
		func(data []byte) error {
			_, err := ws.Write(data)
			return err
		},
		nil, nil,
	)
	if err != nil {
		le.WithError(err).Warn("handshake failed")
		return
	}

	go func() {
		for {
			receivedData := make([]byte, maxPacketSize)
			n, err := ws.Read(receivedData)
			if err != nil {
				le.WithError(err).Warn("error in read loop")
				return
			}

			hs.Handle(receivedData[:n])
		}
	}()

	res, err := hs.Execute(context.Background(), true)
	if err != nil {
		le.WithError(err).Warn("handshake failed")
		return
	}

	remotePeerID, err := peer.IDFromPublicKey(res.Peer)
	if err != nil {
		le.WithError(err).Warn("unable to get id from remote peer public key")
		return
	}

	le.Infof("handshake complete, remote peer id: %s", remotePeerID.Pretty())
}
