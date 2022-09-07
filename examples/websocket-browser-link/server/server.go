package main

import (
	"context"

	"github.com/aperturerobotics/bifrost/examples/websocket-browser-link/common"
	"github.com/aperturerobotics/bifrost/peer"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var le = logrus.NewEntry(log)
var localPrivKey crypto.PrivKey
var localPeerID peer.ID

func init() {
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	ctx := context.Background()
	b, privKey, err := common.BuildCommonBus(ctx, "websocket-browser-link/server")
	if err != nil {
		panic(err)
	}

	_ = privKey
	_, wsRef, err := b.AddDirective(
		resolver.NewLoadControllerWithConfig(&wtpt.Config{
			ListenAddr: ":2015",
		}),
		bus.NewCallbackHandler(func(val directive.AttachedValue) {
			le.Debug("websocket transport resolved")
		}, nil, nil),
	)
	defer wsRef.Release()

	<-ctx.Done()
}
