//go:build js
// +build js

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
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

var port = 2015

func getHTTPBaseURL() string {
	document := js.Global().Get("window").Get("document")
	location := document.Get("location")
	proto := location.Get("protocol").String()

	return fmt.Sprintf("%s//%s:%d", proto, location.Get("hostname"), port)
}

func getWSBaseURL() string {
	document := js.Global().Get("window").Get("document")
	location := document.Get("location")

	wsProtocol := "ws"
	if location.Get("protocol").String() == "https:" {
		wsProtocol = "wss"
	}

	return fmt.Sprintf("%s://%s:%d/ws/", wsProtocol, location.Get("hostname"), port)
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
	b, privKey, err := common.BuildCommonBus(ctx)
	if err != nil {
		return err
	}

	// get the peer id from an http endpoint
	peerIDURL := getHTTPBaseURL() + wtpt.PeerPathSuffix
	le.
		WithField("url", peerIDURL).
		Debug("getting peer id")
	resp, err := http.Get(peerIDURL)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()

	peerIDStr := string(body)
	remotePeerID, err := peer.IDB58Decode(peerIDStr)
	if err != nil {
		return err
	}
	peerIDStr = remotePeerID.String()

	wsBaseURL := getWSBaseURL()
	le.
		WithField("url", wsBaseURL).
		WithField("peer", peerIDStr).
		Debug("contacting websocket peer")

	// NOTE: deterministic due to the use of a prng to generate the key
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

	_, dialRef, err := b.AddDirective(
		link.NewEstablishLinkWithPeer("", remotePeerID),
		nil,
	)
	if err != nil {
		return err
	}
	defer dialRef.Release()

	// open a pubsub channel
	channelID := "test-channel"
	channelSub, _, channelSubRef, err := pubsub.ExBuildChannelSubscription(ctx, b, false, channelID, privKey, nil)
	if err != nil {
		return err
	}
	defer channelSubRef.Release()
	le.Infof("built channel subscription for channel %s", channelID)

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-time.After(time.Second):
		}

		msg := fmt.Sprintf("Hello from browser: %s", time.Now().String())
		if err := channelSub.Publish([]byte(msg)); err != nil {
			le.WithError(err).Warn("unable to publish pubsub message")
		}
	}
}
