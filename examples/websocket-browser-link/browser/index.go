//go:build js
// +build js

package main

import (
	"context"
	"fmt"
	"syscall/js"
	"time"

	"github.com/aperturerobotics/bifrost/examples/websocket-browser-link/common"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

func getWSBaseURL() string {
	document := js.Global().Get("window").Get("document")
	location := document.Get("location")

	wsProtocol := "ws"
	if location.Get("protocol").String() == "https:" {
		wsProtocol = "wss"
	}

	return fmt.Sprintf("%s://%s:%d/ws/", wsProtocol, location.Get("hostname"), 2015)
}

func main() {
	ctx := context.Background()
	le := common.GetLogEntry()
	if err := run(ctx, le); err != nil {
		le.WithError(err).Fatal("error running demo")
	}
}

// run runs the demo.
func run(ctx context.Context, le *logrus.Entry) error {
	b, privKey, err := common.BuildCommonBus(ctx, "websocket-browser-link/client")
	if err != nil {
		return err
	}

	wsBaseURL := getWSBaseURL()
	le.
		WithField("base-url", wsBaseURL).
		Debug("contacting websocket peer")

	// NOTE: deterministic due to the use of a prng to generate the key
	peerIDStr := "12D3KooWCupw8xy9uxGzjhRnCbab3sGr7X67zS4KX64k38dm7XpW"
	_, wsRef, err := b.AddDirective(
		resolver.NewLoadControllerWithConfig(&wtpt.Config{
			Dialers: map[string]*dialer.DialerOpts{
				peerIDStr: {
					Address: wsBaseURL + "bifrost-0.1",
				},
			},
		}),
		bus.NewCallbackHandler(func(val directive.AttachedValue) {
			le.Infof("websocket transport resolved: %#v", val.GetValue())
		}, nil, nil),
	)
	defer wsRef.Release()

	pid, _ := peer.IDB58Decode(peerIDStr)
	_, dialRef, err := b.AddDirective(
		link.NewEstablishLinkWithPeer("", pid),
		nil,
	)
	if err != nil {
		return err
	}
	defer dialRef.Release()

	// open a pubsub channel
	channelID := "test-channel"
	channelSub, channelSubRef, err := pubsub.ExBuildChannelSubscription(ctx, b, channelID, privKey)
	if err != nil {
		return err
	}
	defer channelSubRef.Release()
	le.Infof("built channel subscription for channel %s", channelID)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Second):
		}

		msg := fmt.Sprintf("Hello from browser: %s", time.Now().String())
		if err := channelSub.Publish([]byte(msg)); err != nil {
			le.WithError(err).Warn("unable to publish pubsub message")
		}
	}

	<-ctx.Done()
	return nil
}
