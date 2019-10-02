//+build js
//go:generate gopherjs build -o browser.js index.go

package main

import (
	"context"
	"fmt"

	"github.com/aperturerobotics/bifrost/examples/toys/websocket-browser-link/common"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/gopherjs/gopherjs/js"
)

func getWSBaseURL() string {
	document := js.Global.Get("window").Get("document")
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
	b, peerPrivKey, err := common.BuildCommonBus(ctx)
	if err != nil {
		panic(err)
	}

	_ = peerPrivKey
	wsBaseURL := getWSBaseURL()
	le.
		WithField("base-url", wsBaseURL).
		Debug("contacting websocket peer")

	peerIDStr := "12D3KooWKDHwreWoMwZH2oeijxawWPNvQaAf3zXZfVQW2sbCW21n"
	_, wsRef, err := b.AddDirective(
		resolver.NewLoadControllerWithConfig(&wtpt.Config{
			Dialers: map[string]*dialer.DialerOpts{
				peerIDStr: {
					Address: wsBaseURL + "bifrost-0.1",
				},
			},
		}),
		bus.NewCallbackHandler(func(val directive.AttachedValue) {
			le.Infof("websocket transport resolved: %#v", val)
		}, nil, nil),
	)
	defer wsRef.Release()

	pid, _ := peer.IDB58Decode(peerIDStr)
	_, dialRef, err := b.AddDirective(
		link.NewEstablishLinkWithPeer(pid),
		nil,
	)
	if err != nil {
		panic(err)
	}
	defer dialRef.Release()

	<-ctx.Done()
}
