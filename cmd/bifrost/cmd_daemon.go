//+build !js

package main

import (
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/aperturerobotics/bifrost/daemon"
	"github.com/aperturerobotics/bifrost/daemon/api"
	egctr "github.com/aperturerobotics/bifrost/entitygraph"
	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/stream"
	"github.com/aperturerobotics/bifrost/stream/forwarding"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
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
)

var daemonFlags struct {
	PeerPrivPath    string
	WebsocketListen string
	APIListen       string
	UDPListen       string

	// UDPDial
	// Temporary
	UDPDial cli.StringSlice
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
					Name:        "websocket-listen",
					Usage:       "if set, will listen on address for websocket connections, ex :5111",
					Destination: &daemonFlags.WebsocketListen,
				},
				cli.StringFlag{
					Name:        "udp-listen",
					Usage:       "if set, will listen on address for udp connections, ex :5112",
					Destination: &daemonFlags.UDPListen,
				},
				cli.StringSliceFlag{
					Name:  "udp-dial",
					Usage: "if set, dial address with udp on startup, udp-listen must also be set",
					Value: &daemonFlags.UDPDial,
				},
			},
		},
	)
}

// runDaemon runs the daemon.
func runDaemon(c *cli.Context) error {
	ctx := context.Background()
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	le := logrus.NewEntry(log)

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
	sr.AddFactory(stream_forwarding.NewFactory())

	// Entity graph controller.
	{
		_, egRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfigSingleton(&egc.Config{}),
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
			resolver.NewLoadControllerWithConfigSingleton(&egctr.Config{}),
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
			resolver.NewLoadControllerWithConfigSingleton(&api.Config{
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
		_, wsRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfigSingleton(&wtpt.Config{
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

	if daemonFlags.UDPListen != "" {
		_, udpRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfigSingleton(&udptpt.Config{
				ListenAddr: daemonFlags.UDPListen,
				DialAddrs:  []string(daemonFlags.UDPDial),
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

	// TEST
	{
		rid, err := peer.IDB58Decode("12D3KooWSk3AvMVENL5dXxNFk5CmmUfT93tnSpmAxAGGKesszWTK")
		// rid, err := peer.IDB58Decode("12D3KooWASCtX4bsU1SQAcSW5U19bMK2Qx48kyjoZJkv7Nia6kfu")
		if err != nil {
			return err
		}
		_, udpRef, err := b.AddDirective(
			link.NewOpenStreamWithPeer("test/protocol/1", peer.ID(""), rid, 0, stream.OpenOpts{
				// Encrypted: true,
				// Reliable:  true,
			}),
			bus.NewCallbackHandler(func(val directive.Value) {
				mstrm := val.(link.MountedStream)
				le.
					WithField("protocol-id", mstrm.GetProtocolID()).
					WithField("stream-encrypted", mstrm.GetOpenOpts().Encrypted).
					WithField("stream-reliable", mstrm.GetOpenOpts().Reliable).
					Debug("stream opened with peer")
				strm := mstrm.GetStream()
				strm.Write([]byte("GET / HTTP/1.1\n\n"))

				var dat []byte
				b := make([]byte, 1500)
				for {
					nr, err := strm.Read(b)
					if err != nil {
						if err == io.EOF {
							break
						}

						le.WithError(err).Warn("error reading data")
						break
					}

					dat = append(dat, b[:nr]...)
				}
				le.Debugf("received data: %s", string(dat))
			}, nil, nil),
		)
		if err != nil {
			return errors.Wrap(err, "dial peer")
		}
		defer udpRef.Release()
	}

	_ = d
	<-ctx.Done()
	return nil
}
