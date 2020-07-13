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
	"github.com/aperturerobotics/bifrost/keypem/keyfile"
	stream_forwarding "github.com/aperturerobotics/bifrost/stream/forwarding"
	stream_grpc_accept "github.com/aperturerobotics/bifrost/stream/grpc/accept"
	stream_listening "github.com/aperturerobotics/bifrost/stream/listening"
	xbtpt "github.com/aperturerobotics/bifrost/transport/xbee"
	configset "github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	configset_json "github.com/aperturerobotics/controllerbus/controller/configset/json"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	egc "github.com/aperturerobotics/entitygraph/controller"
	entitygraph_logger "github.com/aperturerobotics/entitygraph/logger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"

	// _ enables the profiling endpoints

	_ "net/http/pprof"
)

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
	peerPriv, err := keyfile.OpenOrWritePrivKey(le, daemonFlags.PeerPrivPath)
	if err != nil {
		return err
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
		_, _, apiRef, err := loader.WaitExecResolverRunning(
			ctx,
			b,
			resolver.NewLoadControllerWithConfig(&api_controller.Config{
				ListenAddr: daemonFlags.APIListen,
			}),
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "listen on grpc api")
		}
		defer apiRef.Release()
		le.Infof("grpc api listening on: %s", daemonFlags.APIListen)
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
	if err := daemonFlags.ApplyToConfigSet(confSet, true); err != nil {
		return err
	}

	// Entity graph controller.
	{
		_, egRef, err := b.AddDirective(
			resolver.NewLoadControllerWithConfig(&egc.Config{}),
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "start entity graph controller")
		}
		defer egRef.Release()
		le.Info("entity graph controller running")
	}

	// Entity graph reporter for bifrost
	{
		_, _, err = b.AddDirective(
			resolver.NewLoadControllerWithConfig(&egctr.Config{}),
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "start entitygraph bifrost reporter")
		}
		le.Info("entitygraph bifrost reporter running")
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
