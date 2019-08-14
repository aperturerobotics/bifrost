//+build !js

package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	bcli "github.com/aperturerobotics/bifrost/cli"
	"github.com/aperturerobotics/bifrost/daemon"
	"github.com/aperturerobotics/bifrost/daemon/api/controller"
	egctr "github.com/aperturerobotics/bifrost/entitygraph"
	"github.com/aperturerobotics/bifrost/keypem"
	"github.com/aperturerobotics/bifrost/stream/forwarding"
	"github.com/aperturerobotics/bifrost/stream/grpc/accept"
	"github.com/aperturerobotics/bifrost/stream/listening"
	xbtpt "github.com/aperturerobotics/bifrost/transport/xbee"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	egc "github.com/aperturerobotics/entitygraph/controller"
	"github.com/aperturerobotics/entitygraph/logger"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

// _ enables the profiling endpoints
import _ "net/http/pprof"

var daemonFlags struct {
	bcli.DaemonArgs

	PeerPrivPath string
	APIListen    string
	ProfListen   string
}

func init() {
	commands = append(
		commands,
		cli.Command{
			Name:   "daemon",
			Usage:  "run a bifrost daemon",
			Action: runDaemon,
			Flags: append(
				(&daemonFlags.DaemonArgs).BuildFlags(),
				cli.StringFlag{
					Name:        "node-priv",
					Usage:       "path to node private key, will be generated if doesn't exist",
					EnvVar:      "BIFROST_NODE_PRIV",
					Value:       "daemon_node_priv.pem",
					Destination: &daemonFlags.PeerPrivPath,
				},
				cli.StringFlag{
					Name:        "api-listen",
					Usage:       "if set, will listen on address for API grpc connections, ex :5110",
					EnvVar:      "BIFROST_API_LISTEN",
					Value:       ":5110",
					Destination: &daemonFlags.APIListen,
				},
				cli.StringFlag{
					Name:        "prof-listen",
					Usage:       "if set, debug profiler will be hosted on the port, ex :8080",
					EnvVar:      "BIFROST_PROF_LISTEN",
					Destination: &daemonFlags.ProfListen,
				},
			),
		},
	)
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
	sr.AddFactory(api_controller.NewFactory(b))
	sr.AddFactory(stream_forwarding.NewFactory(b))
	sr.AddFactory(stream_listening.NewFactory(b))
	sr.AddFactory(stream_grpc_accept.NewFactory(b))

	// Entity graph controller.
	{
		_, egRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfig(&egc.Config{}),
			bus.NewCallbackHandler(func(val directive.AttachedValue) {
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
			bus.NewCallbackHandler(func(val directive.AttachedValue) {
				le.Info("entitygraph bifrost reporter running")
			}, nil, nil),
		)
		if err != nil {
			return errors.Wrap(err, "start entitygraph bifrost reporter")
		}
	}

	_, err = entitygraph_logger.AttachBasicLogger(b, le)
	if err != nil {
		return errors.Wrap(err, "start entitygraph logger")
	}

	// Daemon API
	if daemonFlags.APIListen != "" {
		_, apiRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfig(&api_controller.Config{
				ListenAddr: daemonFlags.APIListen,
			}),
			bus.NewCallbackHandler(func(val directive.AttachedValue) {
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
	confs, err := daemonFlags.BuildControllerConfigs()
	if err != nil {
		return err
	}
	for id, conf := range confs {
		_, confRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfig(conf),
			nil,
		)
		if err != nil {
			return errors.Wrap(err, id+" controller")
		}
		defer confRef.Release()
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
