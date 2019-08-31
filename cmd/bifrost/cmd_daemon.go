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
	api_controller "github.com/aperturerobotics/bifrost/daemon/api/controller"
	egctr "github.com/aperturerobotics/bifrost/entitygraph"
	"github.com/aperturerobotics/bifrost/keypem"
	stream_forwarding "github.com/aperturerobotics/bifrost/stream/forwarding"
	stream_grpc_accept "github.com/aperturerobotics/bifrost/stream/grpc/accept"
	stream_listening "github.com/aperturerobotics/bifrost/stream/listening"
	xbtpt "github.com/aperturerobotics/bifrost/transport/xbee"
	"github.com/aperturerobotics/controllerbus/bus"
	configset "github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	configset_json "github.com/aperturerobotics/controllerbus/controller/configset/json"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	"github.com/aperturerobotics/controllerbus/directive"
	egc "github.com/aperturerobotics/entitygraph/controller"
	entitygraph_logger "github.com/aperturerobotics/entitygraph/logger"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

// _ enables the profiling endpoints
import _ "net/http/pprof"

var daemonFlags struct {
	bcli.DaemonArgs

	WriteConfig  bool
	ConfigPath   string
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
					Name:        "config, c",
					Usage:       "path to configuration yaml file",
					EnvVar:      "BIFROST_CONFIG",
					Value:       "bifrost_daemon.yaml",
					Destination: &daemonFlags.ConfigPath,
				},
				cli.BoolFlag{
					Name:        "write-config",
					Usage:       "write the daemon config file on startup",
					EnvVar:      "BIFROST_WRITE_CONFIG",
					Destination: &daemonFlags.WriteConfig,
				},
				cli.StringFlag{
					Name:        "node-priv",
					Usage:       "path to node private key, will be generated if doesn't exist",
					EnvVar:      "BIFROST_NODE_PRIV",
					Value:       "bifrost_daemon.pem",
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

	// Construct config set.
	confSet := configset.ConfigSet{}

	// Load config file
	configLe := le.WithField("config", daemonFlags.ConfigPath)
	if confPath := daemonFlags.ConfigPath; confPath != "" {
		confDat, err := ioutil.ReadFile(confPath)
		if err != nil {
			if os.IsNotExist(err) {
				if daemonFlags.WriteConfig {
					configLe.Info("cannot find config but write-config is set, continuing")
				} else {
					return errors.Wrapf(
						err,
						"cannot find config at %s",
						daemonFlags.ConfigPath,
					)
				}
			} else {
				return errors.Wrap(err, "load config")
			}
		}

		err = configset_json.UnmarshalYAML(ctx, b, confDat, confSet, true)
		if err != nil {
			return errors.Wrap(err, "unmarshal config yaml")
		}
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

	// ConfigSet controller
	_, csRef, err := b.AddDirective(
		resolver.NewLoadControllerWithConfig(&configset_controller.Config{}),
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "construct configset controller")
	}
	defer csRef.Release()

	// TODO: Load these from CLI/yaml configuration.
	// For now, hardcode it.
	confs, err := daemonFlags.BuildControllerConfigs()
	if err != nil {
		return err
	}

	for id, conf := range confs {
		confSet[id] = configset.NewControllerConfig(1, conf)
	}

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

	if daemonFlags.ConfigPath != "" && daemonFlags.WriteConfig {
		confDat, err := configset_json.MarshalYAML(confSet)
		if err != nil {
			return errors.Wrap(err, "marshal config")
		}
		err = ioutil.WriteFile(daemonFlags.ConfigPath, confDat, 0644)
		if err != nil {
			return errors.Wrap(err, "write config file")
		}
	}

	_, bdbRef, err := b.AddDirective(
		configset.NewApplyConfigSet(confSet),
		nil,
	)
	if err != nil {
		return err
	}
	defer bdbRef.Release()

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
