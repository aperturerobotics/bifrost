package cli

import (
	"context"
	"errors"
	"strings"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/peer/controller"
	"github.com/aperturerobotics/bifrost/pubsub/grpc"
	"github.com/aperturerobotics/bifrost/stream/forwarding"
	"github.com/aperturerobotics/bifrost/stream/grpc/accept"
	"github.com/aperturerobotics/bifrost/stream/grpc/dial"
	"github.com/aperturerobotics/bifrost/stream/listening"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

// ForwardingConf is the forwarding configuration.
type ForwardingConf = stream_forwarding.Config

// AcceptConf is the GRPC accept configuration.
type AcceptConf = stream_grpc_accept.Config

// DialConf is the dial configuration.
type DialConf = stream_grpc_dial.Config

// IdentifyConf is the identify configuration
type IdentifyConf = peer_controller.Config

// SubscribeConf configures the subscribe call
type SubscribeConf = pubsub_grpc.SubscribeRequest

// ListeningConf is the listening configuration
type ListeningConf = stream_listening.Config

// ClientArgs contains the client arguments and functions.
type ClientArgs struct {
	// DialConf is the dialing configuration.
	DialConf
	// ForwardingConf is the forwarding configuration.
	ForwardingConf
	// IdentifyConf is the identify configuration
	IdentifyConf
	// SubscribeConf is the configuration for the subscribe call.
	SubscribeConf
	// AcceptConf is the accept configuration.
	AcceptConf
	// ListeningConf is the listening configuration.
	ListeningConf

	// ctx is the context
	ctx context.Context
	// client is the client instance
	client api.BifrostDaemonClient

	// DialAddr is the address to dial.
	DialAddr string

	// IdentifyKeyPath is the path to the key to read.
	IdentifyKeyPath string
	// IdentifyGenKey indicates we should generate the key if it doesn't exist.
	IdentifyGenKey bool

	// RemotePeerIdsCsv are the set of remote peer IDs to connect to.
	RemotePeerIdsCsv string
}

// ParseRemotePeerIdsCsv parses the RemotePeerIdsCsv field.
func (a *ClientArgs) ParseRemotePeerIdsCsv() []string {
	pts := strings.Split(a.RemotePeerIdsCsv, ",")
	var peerIds []string
	for _, pt := range pts {
		pt = strings.TrimSpace(pt)
		peerIds = append(peerIds, pt)
	}
	return peerIds
}

// BuildFlags attaches the flags to a flag set.
func (a *ClientArgs) BuildFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "dial-addr",
			Usage:       "address to dial API on",
			Destination: &a.DialAddr,
			Value:       "127.0.0.1:5110",
		},
	}
}

// SetClient sets the client instance.
func (a *ClientArgs) SetClient(client api.BifrostDaemonClient) {
	a.client = client
}

// BuildClient builds the client or returns it if it has been set.
func (a *ClientArgs) BuildClient() (api.BifrostDaemonClient, error) {
	if a.client != nil {
		return a.client, nil
	}

	if a.DialAddr == "" {
		return nil, errors.New("dial address is not set")
	}

	clientConn, err := grpc.Dial(a.DialAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	a.client = api.NewBifrostDaemonClient(clientConn)
	return a.client, nil
}

// BuildCommands attaches the commands.
func (a *ClientArgs) BuildCommands() []cli.Command {
	return []cli.Command{
		cli.Command{
			Name:   "local-peers",
			Usage:  "returns local peer info",
			Action: a.RunPeerInfo,
		},
		cli.Command{
			Name:   "bus-info",
			Usage:  "returns bus information",
			Action: a.RunBusInfo,
		},
		cli.Command{
			Name:   "identify",
			Usage:  "Private key will be loaded with a peer controller",
			Action: a.RunIdentifyController,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "peer-priv",
					Usage:       "path to private key file",
					Destination: &a.IdentifyKeyPath,
				},
				&cli.BoolFlag{
					Name:        "generate-priv",
					Usage:       "if set, generate private key if file does not exist",
					Destination: &a.IdentifyGenKey,
				},
			},
		},
		cli.Command{
			Name:   "subscribe",
			Usage:  "Subscribe to a pubsub channel and publish with identified peers",
			Action: a.RunSubscribe,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "channel-id",
					Usage:       "channel id",
					Destination: &a.SubscribeConf.ChannelId,
				},
			},
		},
		cli.Command{
			Name:   "forward",
			Usage:  "Protocol ID will be forwarded to the target multiaddress",
			Action: a.RunForwarding,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "peer-id",
					Usage:       "peer ID to match incoming streams to",
					Destination: &a.ForwardingConf.PeerId,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID to match incoming streams to",
					Destination: &a.ForwardingConf.ProtocolId,
				},
				&cli.StringFlag{
					Name:        "target",
					Usage:       "target multiaddr to forward streams to",
					Destination: &a.ForwardingConf.TargetMultiaddr,
				},
			},
		},
		cli.Command{
			Name:   "accept",
			Usage:  "Single incoming stream with Protocol ID will be accepted",
			Action: a.RunAccept,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "local-peer-id",
					Usage:       "local peer ID to match incoming streams to",
					Destination: &a.AcceptConf.LocalPeerId,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID to match incoming streams to",
					Destination: &a.AcceptConf.ProtocolId,
				},
				&cli.StringFlag{
					Name:        "remote-peer-ids",
					Usage:       "remote peer ids, comma separated, to match, if empty accepts any",
					Destination: &a.RemotePeerIdsCsv,
				},
				&cli.Uint64Flag{
					Name:        "transport-id",
					Usage:       "if set, filter the transport id",
					Destination: &a.AcceptConf.TransportId,
				},
			},
		},
		cli.Command{
			Name:   "dial",
			Usage:  "Single outgoing stream with Protocol ID will be dialed",
			Action: a.RunDial,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "local-peer-id",
					Usage:       "local peer ID to dial from, can be empty",
					Destination: &a.DialConf.LocalPeerId,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID to dial with",
					Destination: &a.DialConf.ProtocolId,
				},
				&cli.StringFlag{
					Name:        "peer-id",
					Usage:       "remote peer id to dial",
					Destination: &a.DialConf.PeerId,
				},
				&cli.Uint64Flag{
					Name:        "transport-id",
					Usage:       "if set, filter the transport id",
					Destination: &a.DialConf.TransportId,
				},
				&cli.BoolTFlag{
					Name:        "encrypted",
					Usage:       "encrypted stream",
					Destination: &a.DialConf.Encrypted,
				},
				&cli.BoolTFlag{
					Name:        "reliable",
					Usage:       "reliable stream",
					Destination: &a.DialConf.Reliable,
				},
			},
		},
		cli.Command{
			Name:   "listen",
			Usage:  "Listen on the multiaddress and forward the connection to a remote stream.",
			Action: a.RunListen,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:        "peer-id",
					Usage:       "peer ID to route traffic to",
					Destination: &a.ListeningConf.RemotePeerId,
				},
				&cli.StringFlag{
					Name:        "from-peer-id",
					Usage:       "peer ID to route traffic from, optional",
					Destination: &a.ListeningConf.LocalPeerId,
				},
				&cli.StringFlag{
					Name:        "protocol-id",
					Usage:       "protocol ID for outgoing streams",
					Destination: &a.ListeningConf.ProtocolId,
				},
				&cli.StringFlag{
					Name:        "listen",
					Usage:       "listen multiaddr",
					Destination: &a.ListeningConf.ListenMultiaddr,
				},
				&cli.BoolTFlag{
					Name:        "encrypted",
					Usage:       "encrypted stream",
					Destination: &a.ListeningConf.Encrypted,
				},
				&cli.BoolTFlag{
					Name:        "reliable",
					Usage:       "reliable stream",
					Destination: &a.ListeningConf.Reliable,
				},
			},
		},
	}
}

// SetContext sets the context.
func (a *ClientArgs) SetContext(c context.Context) {
	a.ctx = c
}

// GetContext returns the context.
func (a *ClientArgs) GetContext() context.Context {
	if c := a.ctx; c != nil {
		return c
	}
	return context.TODO()
}
