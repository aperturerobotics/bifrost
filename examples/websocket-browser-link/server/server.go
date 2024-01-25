package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/examples/websocket-browser-link/common"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sirupsen/logrus"
)

var (
	log          = logrus.New()
	le           = logrus.NewEntry(log)
	localPrivKey crypto.PrivKey
	localPeerID  peer.ID
)

func init() {
	log.SetLevel(logrus.DebugLevel)
}

func main() {
	if err := run(); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	b, privKey, err := common.BuildCommonBus(ctx, "websocket-browser-link/server")
	if err != nil {
		return err
	}
	localPeerID, err := peer.IDFromPrivateKey(privKey)
	if err != nil {
		return err
	}

	_, wsRef, err := b.AddDirective(
		resolver.NewLoadControllerWithConfig(&wtpt.Config{
			ListenAddr: ":2015",
		}),
		bus.NewCallbackHandler(func(val directive.AttachedValue) {
			le.Debug("websocket transport resolved")
		}, nil, nil),
	)
	defer wsRef.Release()

	// accept & echo the pubsub channel
	channelID := "test-channel"
	channelSub, _, channelSubRef, err := pubsub.ExBuildChannelSubscription(ctx, b, false, channelID, privKey, nil)
	if err != nil {
		return err
	}
	defer channelSubRef.Release()
	le.Infof("built channel subscription for channel %s", channelID)

	relHandler := channelSub.AddHandler(func(m pubsub.Message) {
		from := m.GetFrom()
		// ignore unauthenticated and/or from ourselves
		if !m.GetAuthenticated() || from == localPeerID {
			return
		}
		le.Infof("got pubsub message from %s: %s", from.String(), string(m.GetData()))
	})
	defer relHandler()

	<-ctx.Done()
	return nil
}
