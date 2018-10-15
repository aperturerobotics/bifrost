//+build js
//go:generate gopherjs build -o browser.js index.go

package main

import (
	"context"
	"fmt"

	"github.com/aperturerobotics/bifrost/toys/websocket-browser-link/common"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
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

	_, wsRef, err := b.AddDirective(
		resolver.NewLoadControllerWithConfigSingleton(&wtpt.Config{
			DialAddrs: []string{wsBaseURL + "bifrost-0.1"},
		}),
		func(val directive.Value) {
			le.Infof("websocket transport resolved: %#v", val)
		},
	)
	defer wsRef.Release()

	<-ctx.Done()
}
