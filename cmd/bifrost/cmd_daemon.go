//+build !js

package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/aperturerobotics/bifrost/daemon"
	"github.com/aperturerobotics/bifrost/daemon/api/controller"
	egctr "github.com/aperturerobotics/bifrost/entitygraph"
	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/stream/forwarding"
	"github.com/aperturerobotics/bifrost/stream/grpc/accept"
	"github.com/aperturerobotics/bifrost/stream/listening"
	"github.com/aperturerobotics/bifrost/transport/common/dialer"
	"github.com/aperturerobotics/bifrost/transport/common/pconn"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	xbtpt "github.com/aperturerobotics/bifrost/transport/xbee"
	"github.com/aperturerobotics/bifrost/util/backoff"
	"github.com/aperturerobotics/bifrost/util/confparse"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/entitygraph"
	egc "github.com/aperturerobotics/entitygraph/controller"
	"github.com/aperturerobotics/entitygraph/entity"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

// _ enables the profiling endpoints
import _ "net/http/pprof"

var daemonFlags struct {
	PeerPrivPath    string
	WebsocketListen string
	APIListen       string
	UDPListen       string
	ProfListen      string

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

func init() {
	commands = append(
		commands,
		cli.Command{
			Name:   "daemon",
			Usage:  "run a bifrost daemon",
			Action: runDaemon,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "node-priv",
					Usage:       "path to node private key, will be generated if doesn't exist",
					Destination: &daemonFlags.PeerPrivPath,
					Value:       "daemon_node_priv.pem",
				},
				cli.StringFlag{
					Name:        "api-listen",
					Usage:       "if set, will listen on address for API grpc connections, ex :5110",
					Destination: &daemonFlags.APIListen,
					Value:       ":5110",
				},
				cli.StringFlag{
					Name:        "prof-listen",
					Usage:       "if set, debug profiler will be hosted on the port, ex :8080",
					Destination: &daemonFlags.ProfListen,
				},
				cli.StringFlag{
					Name:        "websocket-listen",
					Usage:       "if set, will listen on address for websocket connections, ex :5111",
					Destination: &daemonFlags.WebsocketListen,
				},
				cli.StringFlag{
					Name:        "udp-listen",
					Usage:       "if set, will listen on address for udp connections, ex :5112",
					Destination: &daemonFlags.UDPListen,
				},
				cli.StringFlag{
					Name:        "xbee-device-path",
					Usage:       "xbee device path to open, if set",
					Destination: &daemonFlags.XBeePath,
				},
				cli.IntFlag{
					Name:        "xbee-device-baud",
					Usage:       "xbee device baudrate to use, defaults to 115200",
					Destination: &daemonFlags.XBeeBaud,
					Value:       115200,
				},
				cli.StringSliceFlag{
					Name:  "xbee-peers",
					Usage: "list of peer-id=address known XBee peers",
					Value: &daemonFlags.XbeePeers,
				},
				cli.StringSliceFlag{
					Name:  "udp-peers",
					Usage: "list of peer-id@address known UDP peers",
					Value: &daemonFlags.UDPPeers,
				},
				cli.StringSliceFlag{
					Name:  "websocket-peers",
					Usage: "list of peer-id=address known WebSocket peers",
					Value: &daemonFlags.WebsocketPeers,
				},
			},
		},
	)
}

// parseDialerAddrs parses a dialer map from a string slice
func parseDialerAddrs(ss cli.StringSlice) (map[string]*dialer.DialerOpts, error) {
	m := make(map[string]*dialer.DialerOpts)
	for _, s := range ss {
		pair := strings.Split(s, "@")
		if len(pair) < 2 {
			continue
		}
		pid, err := confparse.ParsePeerID(strings.TrimSpace(pair[0]))
		if err != nil {
			return nil, err
		}
		if pid == peer.ID("") {
			continue
		}
		m[pid.Pretty()] = &dialer.DialerOpts{
			Address: strings.TrimSpace(pair[1]),
		}
	}
	return m, nil
}

// runDaemon runs the daemon.
func runDaemon(c *cli.Context) error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)
	grpc.EnableTracing = daemonFlags.ProfListen != ""

	// Load private key.
	var peerPriv crypto.PrivKey
	peerPrivDat, err := ioutil.ReadFile(daemonFlags.PeerPrivPath)
	if err != nil {
		if os.IsNotExist(err) {
			le.Debug("generating daemon node private key")
			peerPriv, _, err = keypem.GeneratePrivKey()
			if err != nil {
				return errors.Wrap(err, "generate priv key")
			}
		} else {
			return errors.Wrap(err, "read priv key")
		}

		peerPrivDat, err = keypem.MarshalPrivKeyPem(peerPriv)
		if err != nil {
			return errors.Wrap(err, "marshal priv key")
		}

		if err := ioutil.WriteFile(daemonFlags.PeerPrivPath, peerPrivDat, 0644); err != nil {
			return errors.Wrap(err, "write priv key")
		}
	} else {
		peerPriv, err = keypem.ParsePrivKeyPem(peerPrivDat)
		if err != nil {
			return errors.Wrap(err, "parse node priv key")
		}
	}

	d, err := daemon.NewDaemon(ctx, peerPriv, daemon.ConstructOpts{
		LogEntry: le,
	})
	if err != nil {
		return errors.Wrap(err, "construct daemon")
	}

	b := d.GetControllerBus()
	sr := d.GetStaticResolver()
	sr.AddFactory(egctr.NewFactory(b))
	sr.AddFactory(xbtpt.NewFactory(b))
	sr.AddFactory(stream_forwarding.NewFactory(b))
	sr.AddFactory(stream_listening.NewFactory(b))
	sr.AddFactory(stream_grpc_accept.NewFactory(b))

	// Entity graph controller.
	{
		_, egRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfig(&egc.Config{}),
			bus.NewCallbackHandler(func(val directive.Value) {
				le.Info("entity graph controller running")
			}, nil, nil),
		)
		if err != nil {
			return errors.Wrap(err, "start entity graph controller")
		}
		defer egRef.Release()
	}

	// Entity graph reporter for bifrost
	{
		_, _, err = b.AddDirective(
			resolver.NewLoadControllerWithConfig(&egctr.Config{}),
			bus.NewCallbackHandler(func(val directive.Value) {
				le.Info("entitygraph bifrost reporter running")
			}, nil, nil),
		)
		if err != nil {
			return errors.Wrap(err, "start entitygraph bifrost reporter")
		}
	}

	// TODO: something better than this logger
	{
		le.Debug("constructing entitygraph logger")
		_, _, err = b.AddDirective(
			entitygraph.NewObserveEntityGraph(),
			bus.NewCallbackHandler(func(val directive.Value) {
				ent := val.(entity.Entity)
				le.Infof("EntityGraph: value added: %s: %s", ent.GetEntityTypeName(), ent.GetEntityID())
			}, func(val directive.Value) {
				ent := val.(entity.Entity)
				le.Infof("EntityGraph: value removed: %s: %s", ent.GetEntityTypeName(), ent.GetEntityID())
			}, nil),
		)
		if err != nil {
			return errors.Wrap(err, "start entitygraph logger")
		}
	}

	// Daemon API
	if daemonFlags.APIListen != "" {
		_, apiRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfig(&api_controller.Config{
				ListenAddr: daemonFlags.APIListen,
			}),
			bus.NewCallbackHandler(func(val directive.Value) {
				le.Infof("grpc api listening on: %s", daemonFlags.APIListen)
			}, nil, nil),
		)
		if err != nil {
			return errors.Wrap(err, "listen on grpc api")
		}
		defer apiRef.Release()
	}

	// TODO: Load these from CLI/yaml configuration.
	// For now, hardcode it.
	if daemonFlags.WebsocketListen != "" {
		staticPeers, err := parseDialerAddrs(daemonFlags.WebsocketPeers)
		if err != nil {
			return errors.Wrap(err, "websocket-peers")
		}

		_, wsRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfig(&wtpt.Config{
				Dialers:    staticPeers,
				ListenAddr: daemonFlags.WebsocketListen,
			}),
			bus.NewCallbackHandler(func(val directive.Value) {
				le.Infof("websocket listening on: %s", daemonFlags.WebsocketListen)
			}, nil, nil),
		)
		if err != nil {
			return errors.Wrap(err, "listen on websocket")
		}
		defer wsRef.Release()
	}

	if daemonFlags.XBeePath != "" {
		staticPeers, err := parseDialerAddrs(daemonFlags.XbeePeers)
		if err != nil {
			return errors.Wrap(err, "xbee-peers")
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

		_, xbRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfig(&xbtpt.Config{
				DevicePath: daemonFlags.XBeePath,
				DeviceBaud: int32(daemonFlags.XBeeBaud),
				Dialers:    staticPeers,
				PacketOpts: &pconn.Opts{
					Mtu: 200,
					// KcpMode: pconn.KCPMode_KCPMode_FAST3,
					KcpMode:       pconn.KCPMode_KCPMode_SLOW1,
					BlockCrypt:    pconn.BlockCrypt_BlockCrypt_TWOFISH,
					BlockCompress: pconn.BlockCompress_BlockCompress_SNAPPY,
					// BlockCrypt: pconn.BlockCrypt_BlockCrypt_TWOFISH,
					// DataShards:   3,
					// ParityShards: 3,
				},
			}),
			bus.NewCallbackHandler(func(val directive.Value) {
				le.Infof("xbee listening on: %s@%d", daemonFlags.XBeePath, daemonFlags.XBeeBaud)
			}, nil, nil),
		)
		if err != nil {
			return errors.Wrap(err, "listen on xbee")
		}
		defer xbRef.Release()
	}

	if daemonFlags.UDPListen != "" {
		staticPeers, err := parseDialerAddrs(daemonFlags.UDPPeers)
		if err != nil {
			return errors.Wrap(err, "udp-peers")
		}

		_, udpRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfig(&udptpt.Config{
				Dialers:    staticPeers,
				ListenAddr: daemonFlags.UDPListen,
				PacketOpts: &pconn.Opts{
					// KcpMode:    pconn.KCPMode_KCPMode_FAST3,
					KcpMode: pconn.KCPMode_KCPMode_NORMAL,
					// BlockCrypt: pconn.BlockCrypt_BlockCrypt_AES256,
					BlockCrypt: pconn.BlockCrypt_BlockCrypt_NONE,
					// DataShards:   10,
					// ParityShards: 3,
				},
			}),
			bus.NewCallbackHandler(func(val directive.Value) {
				le.Infof("UDP listening on: %s", daemonFlags.UDPListen)
			}, nil, nil),
		)
		if err != nil {
			return errors.Wrap(err, "listen on udp")
		}
		defer udpRef.Release()
	}

	if daemonFlags.ProfListen != "" {
		runtime.SetBlockProfileRate(1)
		runtime.SetMutexProfileFraction(1)
		go func() {
			le.Debugf("profiling listener running: %s", daemonFlags.ProfListen)
			err := http.ListenAndServe(daemonFlags.ProfListen, nil)
			le.WithError(err).Warn("profiling listener exited")
		}()
	}

	_ = d
	<-ctx.Done()
	return nil
}
