package cli

import (
	"github.com/aperturerobotics/bifrost/link/hold-open"
	"github.com/aperturerobotics/bifrost/transport/common/pconn"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	xbtpt "github.com/aperturerobotics/bifrost/transport/xbee"
	"github.com/aperturerobotics/bifrost/util/backoff"
	"github.com/aperturerobotics/bifrost/util/blockcrypt"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// DaemonArgs contains common flags for bifrost-powered daemons.
type DaemonArgs struct {
	WebsocketListen string
	UDPListen       string
	HoldOpenLinks   bool

	XBeePath string
	XBeeBaud int

	// UDPPeers is a static peer list
	// peer-id@address
	UDPPeers cli.StringSlice
	// WebsocketPeers is a static peer list
	// peer-id@address
	WebsocketPeers cli.StringSlice
	// XbeePeers is a static peer list
	// peer-id@address
	XbeePeers cli.StringSlice
}

// BuildFlags attaches the flags to a flag set.
func (a *DaemonArgs) BuildFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:        "hold-open-links",
			Usage:       "if set, hold open links without an inactivity timeout",
			EnvVar:      "BIFROST_HOLD_OPEN_LINKS",
			Destination: &a.HoldOpenLinks,
		},
		cli.StringFlag{
			Name:        "websocket-listen",
			Usage:       "if set, will listen on address for websocket connections, ex :5111",
			EnvVar:      "BIFROST_WS_LISTEN",
			Destination: &a.WebsocketListen,
		},
		cli.StringFlag{
			Name:        "udp-listen",
			Usage:       "if set, will listen on address for udp connections, ex :5112",
			EnvVar:      "BIFROST_UDP_LISTEN",
			Destination: &a.UDPListen,
		},
		cli.StringFlag{
			Name:        "xbee-device-path",
			Usage:       "xbee device path to open, if set",
			EnvVar:      "BIFROST_XBEE_PATH",
			Destination: &a.XBeePath,
		},
		cli.IntFlag{
			Name:        "xbee-device-baud",
			Usage:       "xbee device baudrate to use, defaults to 115200",
			EnvVar:      "BIFROST_XBEE_BAUD",
			Destination: &a.XBeeBaud,
			Value:       115200,
		},
		cli.StringSliceFlag{
			Name:   "xbee-peers",
			Usage:  "list of peer-id@address known XBee peers",
			EnvVar: "BIFROST_XBEE_PEERS",
			Value:  &a.XbeePeers,
		},
		cli.StringSliceFlag{
			Name:   "udp-peers",
			Usage:  "list of peer-id@address known UDP peers",
			EnvVar: "BIFROST_UDP_PEERS",
			Value:  &a.UDPPeers,
		},
		cli.StringSliceFlag{
			Name:   "websocket-peers",
			Usage:  "list of peer-id@address known WebSocket peers",
			EnvVar: "BIFROST_WS_PEERS",
			Value:  &a.WebsocketPeers,
		},
	}

}

// BuildControllerConfigs builds controller configurations from the args.
// Map is from string descriptor to config object.
func (a *DaemonArgs) BuildControllerConfigs() (map[string]config.Config, error) {
	confs := make(map[string]config.Config)

	if a.WebsocketListen != "" {
		staticPeers, err := parseDialerAddrs(a.WebsocketPeers)
		if err != nil {
			return nil, errors.Wrap(err, "websocket-peers")
		}

		confs["websocket"] = &wtpt.Config{
			Dialers:    staticPeers,
			ListenAddr: a.WebsocketListen,
		}
	}

	if a.XBeePath != "" {
		staticPeers, err := parseDialerAddrs(a.XbeePeers)
		if err != nil {
			return nil, errors.Wrap(err, "xbee-peers")
		}
		for _, peer := range staticPeers {
			peer.Backoff = &backoff.Backoff{
				BackoffKind: backoff.BackoffKind_BackoffKind_EXPONENTIAL,
				Exponential: &backoff.Exponential{
					InitialInterval:     1000,
					RandomizationFactor: 0.8,
					Multiplier:          1.7,
				},
			}
		}

		confs["xbee"] = &xbtpt.Config{
			DevicePath: a.XBeePath,
			DeviceBaud: int32(a.XBeeBaud),
			Dialers:    staticPeers,
			PacketOpts: &pconn.Opts{
				Mtu:     150,
				KcpMode: pconn.KCPMode_KCPMode_FAST3,
				// KcpMode: pconn.KCPMode_KCPMode_SLOW1,
				// BlockCrypt: pconn.BlockCrypt_BlockCrypt_TWOFISH,
				BlockCrypt:    blockcrypt.BlockCrypt_BlockCrypt_SALSA20,
				BlockCompress: pconn.BlockCompress_BlockCompress_SNAPPY,
				// BlockCompress: pconn.BlockCompress_BlockCompress_LZ4,
				// DataShards:   3,
				// ParityShards: 3,
			},
		}
	}

	if a.UDPListen != "" {
		staticPeers, err := parseDialerAddrs(a.UDPPeers)
		if err != nil {
			return nil, errors.Wrap(err, "udp-peers")
		}

		confs["udp"] = &udptpt.Config{
			Dialers:    staticPeers,
			ListenAddr: a.UDPListen,
			PacketOpts: &pconn.Opts{
				KcpMode:       pconn.KCPMode_KCPMode_FAST3,
				BlockCrypt:    blockcrypt.BlockCrypt_BlockCrypt_SALSA20,
				BlockCompress: pconn.BlockCompress_BlockCompress_NONE,
				// KcpMode: pconn.KCPMode_KCPMode_NORMAL,
				// BlockCrypt: pconn.BlockCrypt_BlockCrypt_AES256,
				// DataShards:   10,
				// ParityShards: 3,
			},
		}
	}

	if a.HoldOpenLinks {
		confs["hold-open"] = &link_holdopen_controller.Config{}
	}

	for id, conf := range confs {
		if err := conf.Validate(); err != nil {
			return nil, errors.Wrap(err, id)
		}
	}
	return confs, nil
}
