package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/aperturerobotics/bifrost/daemon"
	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/keypem"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
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

	if daemonFlags.APIListen != "" {
		_, apiRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfigSingleton(&api.Config{
				ListenAddr: daemonFlags.APIListen,
			}),
			func(val directive.Value) {
				le.Infof("grpc api listening on: %s", daemonFlags.APIListen)
			},
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
			func(val directive.Value) {
				le.Infof("websocket listening on: %s", daemonFlags.WebsocketListen)
			},
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
			}),
			func(val directive.Value) {
				le.Infof("UDP listening on: %s", daemonFlags.UDPListen)
			},
		)
		if err != nil {
			return errors.Wrap(err, "listen on udp")
		}
		defer udpRef.Release()
	}

	_ = d
	<-ctx.Done()
	return nil
}
