//+build !js

package main

import (
	"sync"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

var clientDialAddr string
var clientCommands []cli.Command

func init() {
	clientCommands = append(
		clientCommands,
		cli.Command{
			Name:   "local-peers",
			Usage:  "returns local peer info",
			Action: runPeerInfo,
		},
		cli.Command{
			Name:   "bus-info",
			Usage:  "returns bus information",
			Action: runBusInfo,
		},
		cli.Command{
			Name:   "forward",
			Usage:  "Protocol ID will be forwarded to the target multiaddress",
			Action: runForwardController,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "peer-id",
					Usage:       "peer ID to match incoming streams to",
					Destination: &forwardingConf.PeerId,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID to match incoming streams to",
					Destination: &forwardingConf.ProtocolId,
				},
				&cli.StringFlag{
					Name:        "target",
					Usage:       "target multiaddr to forward streams to",
					Destination: &forwardingConf.TargetMultiaddr,
				},
			},
		},
		cli.Command{
			Name:   "listen",
			Usage:  "Listen on the multiaddress and forward the connection to a remote stream.",
			Action: runListeningController,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "peer-id",
					Usage:       "peer ID to route traffic to",
					Destination: &listeningConf.RemotePeerId,
				},
				&cli.StringFlag{
					Name:        "from-peer-id",
					Usage:       "peer ID to route traffic from, optional",
					Destination: &listeningConf.LocalPeerId,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID for outgoing streams",
					Destination: &listeningConf.ProtocolId,
				},
				&cli.StringFlag{
					Name:        "listen",
					Usage:       "listen multiaddr",
					Destination: &listeningConf.ListenMultiaddr,
				},
				&cli.BoolTFlag{
					Name:        "encrypted",
					Usage:       "encrypted stream",
					Destination: &listeningConf.Encrypted,
				},
				&cli.BoolTFlag{
					Name:        "reliable",
					Usage:       "reliable stream",
					Destination: &listeningConf.Reliable,
				},
			},
		})
	commands = append(
		commands,
		cli.Command{
			Name:        "client",
			Usage:       "client sub-commands",
			After:       runCloseClient,
			Subcommands: clientCommands,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "dial-addr",
					Usage:       "address to dial API on",
					Destination: &clientDialAddr,
					Value:       "localhost:5110",
				},
			},
		},
	)
}

var clientMtx sync.Mutex
var client api.BifrostDaemonServiceClient
var clientConn *grpc.ClientConn

// GetClient builds / returns the client.
func GetClient() (api.BifrostDaemonServiceClient, error) {
	clientMtx.Lock()
	defer clientMtx.Unlock()

	if client != nil {
		return client, nil
	}

	var err error
	clientConn, err = grpc.Dial(clientDialAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client = api.NewBifrostDaemonServiceClient(clientConn)
	return client, nil
}

func runCloseClient(ctx *cli.Context) error {
	if clientConn != nil {
		clientConn.Close()
		clientConn = nil
	}
	return nil
}
