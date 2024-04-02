//go:build !js
// +build !js

package main

import (
	"context"
	"net/http"
	"os"
	"runtime"
	"time"

	"net/http/pprof"

	bcli "github.com/aperturerobotics/bifrost/cli"
	"github.com/aperturerobotics/bifrost/daemon"
	bifrost_api "github.com/aperturerobotics/bifrost/daemon/api"
	api_controller "github.com/aperturerobotics/bifrost/daemon/api/controller"
	egctr "github.com/aperturerobotics/bifrost/entitygraph"
	"github.com/aperturerobotics/bifrost/keypem/keyfile"
	bus_api "github.com/aperturerobotics/controllerbus/bus/api"
	configset "github.com/aperturerobotics/controllerbus/controller/configset"
	configset_controller "github.com/aperturerobotics/controllerbus/controller/configset/controller"
	configset_json "github.com/aperturerobotics/controllerbus/controller/configset/json"
	"github.com/aperturerobotics/controllerbus/controller/loader"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
	egc "github.com/aperturerobotics/entitygraph/controller"
	entitygraph_logger "github.com/aperturerobotics/entitygraph/logger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
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
		&cli.Command{
			Name:   "daemon",
			Usage:  "run a bifrost daemon",
			Action: runDaemon,
			Flags: append(
				(&daemonFlags.DaemonArgs).BuildFlags(),
				&cli.StringFlag{
					Name:        "config",
					Aliases:     []string{"c"},
					Usage:       "path to configuration yaml file",
					EnvVars:     []string{"BIFROST_CONFIG"},
					Value:       "bifrost_daemon.yaml",
					Destination: &daemonFlags.ConfigPath,
				},
				&cli.BoolFlag{
					Name:        "write-config",
					Usage:       "write the daemon config file on startup",
					EnvVars:     []string{"BIFROST_WRITE_CONFIG"},
					Destination: &daemonFlags.WriteConfig,
				},
				&cli.StringFlag{
					Name:        "node-priv",
					Usage:       "path to node private key, will be generated if doesn't exist",
					EnvVars:     []string{"BIFROST_NODE_PRIV"},
					Value:       "bifrost_daemon.pem",
					Destination: &daemonFlags.PeerPrivPath,
				},
				&cli.StringFlag{
					Name:        "api-listen",
					Usage:       "if set, will listen on address for API drpc connections, ex :5110",
					EnvVars:     []string{"BIFROST_API_LISTEN"},
					Destination: &daemonFlags.APIListen,
				},
				&cli.StringFlag{
					Name:        "prof-listen",
					Usage:       "if set, debug profiler will be hosted on the port, ex :8080",
					EnvVars:     []string{"BIFROST_PROF_LISTEN"},
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

	// Construct config set.
	confSet := configset.ConfigSet{}

	// Load config file
	configLe := le.WithField("config", daemonFlags.ConfigPath)
	if confPath := daemonFlags.ConfigPath; confPath != "" {
		confDat, err := os.ReadFile(confPath)
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

		_, err = configset_json.UnmarshalYAML(ctx, b, confDat, confSet, true)
		if err != nil {
			return errors.Wrap(err, "unmarshal config yaml")
		}
	}

	// Apply factories
	daemonFlags.ApplyFactories(b, sr)

	// Daemon API
	if daemonFlags.APIListen != "" {
		_, _, apiRef, err := loader.WaitExecControllerRunning(
			ctx,
			b,
			resolver.NewLoadControllerWithConfig(&api_controller.Config{
				ListenAddr: daemonFlags.APIListen,
				ApiConfig:  &bifrost_api.Config{},
				BusApiConfig: &bus_api.Config{
					EnableExecController: true,
				},
			}),
			nil,
		)
		if err != nil {
			return errors.Wrap(err, "listen on api")
		}
		defer apiRef.Release()
		le.Infof("api listening on: %s", daemonFlags.APIListen)
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

	// Load config sets and factories
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
		err = os.WriteFile(daemonFlags.ConfigPath, confDat, 0o644)
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
			mux := http.NewServeMux()

			// Register pprof handlers
			mux.HandleFunc("/debug/pprof/", pprof.Index)
			mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
			// Manually add support for paths linked to by index page at /debug/pprof/
			mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
			mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
			mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
			mux.Handle("/debug/pprof/block", pprof.Handler("block"))

			server := &http.Server{Addr: daemonFlags.ProfListen, Handler: mux, ReadHeaderTimeout: time.Second * 10}
			err := server.ListenAndServe()
			le.WithError(err).Warn("profiling listener exited")
		}()
	}

	_ = d
	<-ctx.Done()
	return nil
}
