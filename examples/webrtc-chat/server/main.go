package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/core"
	"github.com/aperturerobotics/bifrost/peer"
	peer_controller "github.com/aperturerobotics/bifrost/peer/controller"
	signaling_rpc_server "github.com/aperturerobotics/bifrost/signaling/rpc/server"
	stream_srpc_server "github.com/aperturerobotics/bifrost/stream/srpc/server"
	"github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/cli"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/sirupsen/logrus"
)

// listenAddr is the listen address
var listenAddr string = ":2253"

var httpPath string = "/bifrost-ws"

func main() {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	app := cli.NewApp()
	app.Name = "webrtc-chat-signaling-server"
	app.Usage = "Hosts a WebSocket server and a signaling service for WebRTC chat"
	app.HideHelpCommand = true

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:        "listen",
			Usage:       "address to listen on",
			EnvVars:     []string{"LISTEN"},
			Value:       listenAddr,
			Destination: &listenAddr,
		},
		&cli.StringFlag{
			Name:        "http path",
			Usage:       "http path to listen on",
			EnvVars:     []string{"HTTP_PATH"},
			Value:       httpPath,
			Destination: &httpPath,
		},
	}
	app.Action = func(c *cli.Context) error {
		return run(ctx, le)
	}

	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func run(ctx context.Context, le *logrus.Entry) error {
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return err
	}

	sr.AddFactory(signaling_rpc_server.NewFactory(b))

	localPeer, _, _, err := peer.NewPeerWithGenerateED25519()
	if err != nil {
		return err
	}
	
	le.Infof("Generated new private key")


	localPeerID := localPeer.GetPeerID()
	localPeerIDStr := localPeerID.String()
	le.Infof("starting with peer id: %v", localPeerIDStr)

	peerCtrl := peer_controller.NewController(le, localPeer)
	relPeerCtrl, err := b.AddController(ctx, peerCtrl, nil)
	if err != nil {
		return err
	}
	defer relPeerCtrl()

	protocolID := "webrtc-chat"
	_, _, serverRef, err := loader.WaitExecControllerRunning(
		ctx,
		b,
		resolver.NewLoadControllerWithConfig(&signaling_rpc_server.Config{
			Server: &stream_srpc_server.Config{
				PeerIds:     []string{localPeer.GetPeerID().String()},
				ProtocolIds: []string{protocolID},
			},
		}),
		nil,
	)
	if err != nil {
		return err
	}
	defer serverRef.Release()

	_, _, wsRef, err := loader.WaitExecControllerRunning(
		ctx,
		b,
		resolver.NewLoadControllerWithConfig(&websocket.Config{
			TransportPeerId: localPeerIDStr,
			ListenAddr:      listenAddr,
			HttpPath:        httpPath,
		}),
		nil,
	)
	if err != nil {
		return err
	}
	defer wsRef.Release()

	le.Infof("WebRTC Chat signaling server running at %s%s", listenAddr, httpPath)
	le.Infof("Protocol ID: %s", protocolID)

	<-ctx.Done()
	return nil
}
