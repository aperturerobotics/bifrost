package main

import (
	"context"
	"os"

	"github.com/aperturerobotics/bifrost/core"
	"github.com/aperturerobotics/bifrost/keypem/keyfile"
	"github.com/aperturerobotics/bifrost/peer"
	peer_controller "github.com/aperturerobotics/bifrost/peer/controller"
	signaling_rpc_server "github.com/aperturerobotics/bifrost/signaling/rpc/server"
	stream_srpc_server "github.com/aperturerobotics/bifrost/stream/srpc/server"
	"github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/sirupsen/logrus"
	"github.com/aperturerobotics/cli"
)

// listenAddr is the listen address
var listenAddr string = ":2253"

// httpPath is the http path to listen on
var httpPath string = "/bifrost-ws"

func main() {
	// Create the context and logger.
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

	// Create the CLI flags.
	app := cli.NewApp()
	app.Name = "webrtc-signaling-server"
	app.Usage = "Hosts a WebSocket server and a signaling service"
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

	// Run the server.
	if err := app.Run(os.Args); err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

func run(ctx context.Context, le *logrus.Entry) error {
	// Create the Controller Bus which will manage our controllers.
	b, sr, err := core.NewCoreBus(ctx, le)
	if err != nil {
		return err
	}

	// Add the signaling RPC server factory to the bus.
	sr.AddFactory(signaling_rpc_server.NewFactory(b))

	// Load or create the peer private key at ./server.pem.
	privKey, err := keyfile.OpenOrWritePrivKey(le, "../../priv/node-3.pem")
	if err != nil {
		return err
	}

	// Load the peer from the private key
	localPeer, err := peer.NewPeer(privKey)
	if err != nil {
		return err
	}

	// Local peer ID
	localPeerID := localPeer.GetPeerID()
	localPeerIDStr := localPeerID.String()
	le.Infof("starting with peer id: %v", localPeerIDStr)

	// Load the peer to the bus so the RPC server can use it.
	peerCtrl := peer_controller.NewController(le, localPeer)
	relPeerCtrl, err := b.AddController(ctx, peerCtrl, nil)
	if err != nil {
		return err
	}
	defer relPeerCtrl()

	// Load the signaling RPC server, listening on localPeer at protocolID.
	protocolID := "webrtc/signaling"
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

	// Listen for incoming WebSocket connections.
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

	// The program is now running and accepting connections.
	<-ctx.Done()
	return nil
}
