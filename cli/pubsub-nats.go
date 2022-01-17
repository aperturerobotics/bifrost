//go:build bifrost_pubsub_nats
// +build bifrost_pubsub_nats

package cli

import (
	"github.com/aperturerobotics/bifrost/pubsub/nats"
	nats_controller "github.com/aperturerobotics/bifrost/pubsub/nats/controller"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
)

func init() {
	pubsubFactories = append(pubsubFactories, func(b bus.Bus) controller.Factory {
		return nats_controller.NewFactory(b)
	})
	pubsubProviders["nats"] = func(args *DaemonArgs) (config.Config, error) {
		return &nats_controller.Config{
			PeerId: "any",
			NatsConfig: &nats.Config{
				ClusterName: "bifrost-cli-default",
			},
		}, nil
	}
}
