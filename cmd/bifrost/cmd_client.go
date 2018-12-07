//+build !js

package main

import (
	"strings"
	"sync"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

var clientDialAddr string
var clientCommands []cli.Command

var remotePeerIdsCsv string

func parseRemotePeerIdsCsv() []string {
	pts := strings.Split(remotePeerIdsCsv, ",")
	var peerIds []string
	for _, pt := range pts {
		pt = strings.TrimSpace(pt)
		peerIds = append(peerIds, pt)
	}
	return peerIds
}

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
			Name:   "identify",
			Usage:  "Private key will be loaded with a peer controller",
			Action: runIdentifyController,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "peer-priv",
					Usage:       "path to private key file",
					Destination: &identifyKeyPath,
				},
				&cli.BoolFlag{
					Name:        "generate-priv",
					Usage:       "if set, generate private key if file does not exist",
					Destination: &identifyGenKey,
				},
			},
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
			Name:   "accept",
			Usage:  "Single incoming stream with Protocol ID will be accepted",
			Action: runAcceptController,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "local-peer-id",
					Usage:       "local peer ID to match incoming streams to",
					Destination: &grpcacceptConf.LocalPeerId,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID to match incoming streams to",
					Destination: &grpcacceptConf.ProtocolId,
				},
				&cli.StringFlag{
					Name:        "remote-peer-ids",
					Usage:       "remote peer ids, comma separated, to match, if empty accepts any",
					Destination: &remotePeerIdsCsv,
				},
				&cli.Uint64Flag{
					Name:        "transport-id",
					Usage:       "if set, filter the transport id",
					Destination: &grpcacceptConf.TransportId,
				},
			},
		},
		cli.Command{
			Name:   "dial",
			Usage:  "Single outgoing stream with Protocol ID will be dialed",
			Action: runDialController,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "local-peer-id",
					Usage:       "local peer ID to dial from, can be empty",
					Destination: &grpcdialConf.LocalPeerId,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID to dial with",
					Destination: &grpcdialConf.ProtocolId,
				},
				&cli.StringFlag{
					Name:        "peer-id",
					Usage:       "remote peer id to dial",
					Destination: &grpcdialConf.PeerId,
				},
				&cli.Uint64Flag{
					Name:        "transport-id",
					Usage:       "if set, filter the transport id",
					Destination: &grpcdialConf.TransportId,
				},
				&cli.BoolTFlag{
					Name:        "encrypted",
					Usage:       "encrypted stream",
					Destination: &grpcdialConf.Encrypted,
				},
				&cli.BoolTFlag{
					Name:        "reliable",
					Usage:       "reliable stream",
					Destination: &grpcdialConf.Reliable,
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
