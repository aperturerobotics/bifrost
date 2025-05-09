package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	signaling "github.com/aperturerobotics/bifrost/signaling/rpc"
	signaling_rpc_client "github.com/aperturerobotics/bifrost/signaling/rpc/client"
	stream_srpc_client "github.com/aperturerobotics/bifrost/stream/srpc/client"
	webrtc "github.com/aperturerobotics/bifrost/transport/webrtc"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config/configset"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/sirupsen/logrus"
)

var (
	signalingServerURL = "ws://localhost:2020/ws"
	signalingID        = "webrtc-chat-signaling"
)

type Message struct {
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := bus.NewBus(resolver.NewResolver(le))

	js.Global().Set("initWebRTC", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		username := args[0].String()
		go initWebRTC(ctx, le, b, username)
		return nil
	}))

	<-make(chan bool)
}

func initWebRTC(ctx context.Context, le *logrus.Entry, b bus.Bus, username string) {
	updateStatus("Connecting to signaling server...")

	signalClientConf := &signaling_rpc_client.Config{
		SignalingId: signalingID,
		Client: &stream_srpc_client.Config{
			WebsocketUrl: signalingServerURL,
		},
	}

	webrtcTptConf := &webrtc.Config{
		SignalingId: signalingID,
		AllPeers:    true,
		Verbose:     true,
	}

	cs := configset.NewConfigSet()
	cs.AddConfig("signaling-client", signalClientConf)
	cs.AddConfig("webrtc-tpt", webrtcTptConf)
	if err := b.AddConfig(ctx, cs); err != nil {
		le.WithError(err).Error("failed to add config")
		updateStatus("Failed to connect: " + err.Error())
		return
	}

	updateStatus("Connected! Waiting for peers...")

	js.Global().Set("sendMessage", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		content := args[0].String()
		msg := Message{
			Sender:    username,
			Content:   content,
			Timestamp: time.Now(),
		}
		
		
		return nil
	}))
}

func updateStatus(status string) {
	js.Global().Get("document").Call("getElementById", "status").Set("textContent", status)
}

func addMessage(sender, content string, isSent bool) {
	messageType := "received"
	if isSent {
		messageType = "sent"
	}
	
	js.Global().Call("addMessage", sender, content, isSent)
}
