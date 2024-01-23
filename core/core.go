package core

import (
	"context"

	bifrosteg "github.com/aperturerobotics/bifrost/entitygraph"
	http_listener "github.com/aperturerobotics/bifrost/http/listener"
	link_establish_controller "github.com/aperturerobotics/bifrost/link/establish"
	link_holdopen_controller "github.com/aperturerobotics/bifrost/link/hold-open"
	nctr "github.com/aperturerobotics/bifrost/peer/controller"
	floodsub_controller "github.com/aperturerobotics/bifrost/pubsub/floodsub/controller"
	pubsub_relay "github.com/aperturerobotics/bifrost/pubsub/relay"
	signaling_rpc_client "github.com/aperturerobotics/bifrost/signaling/rpc/client"
	signaling_rpc_server "github.com/aperturerobotics/bifrost/signaling/rpc/server"
	stream_api_accept "github.com/aperturerobotics/bifrost/stream/api/accept"
	stream_echo "github.com/aperturerobotics/bifrost/stream/echo"
	stream_forwarding "github.com/aperturerobotics/bifrost/stream/forwarding"
	stream_listening "github.com/aperturerobotics/bifrost/stream/listening"
	tptaddr_controller "github.com/aperturerobotics/bifrost/tptaddr/controller"
	tptaddr_static "github.com/aperturerobotics/bifrost/tptaddr/static"
	iproctpt "github.com/aperturerobotics/bifrost/transport/inproc"
	udptpt "github.com/aperturerobotics/bifrost/transport/udp"
	"github.com/aperturerobotics/bifrost/transport/webrtc"
	wtpt "github.com/aperturerobotics/bifrost/transport/websocket"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/resolver/static"
	cbc "github.com/aperturerobotics/controllerbus/core"
	egc "github.com/aperturerobotics/entitygraph/controller"
	"github.com/sirupsen/logrus"
)

// NewCoreBus constructs a standard in-memory bus stack with Bifrost controllers.
func NewCoreBus(
	ctx context.Context,
	le *logrus.Entry,
	opts ...cbc.Option,
) (bus.Bus, *static.Resolver, error) {
	b, sr, err := cbc.NewCoreBus(ctx, le, opts...)
	if err != nil {
		return nil, nil, err
	}

	AddFactories(b, sr)
	return b, sr, nil
}

// AddFactories adds factories to an existing static resolver.
func AddFactories(b bus.Bus, sr *static.Resolver) {
	// node controller
	sr.AddFactory(nctr.NewFactory())

	// link management controllers
	sr.AddFactory(link_holdopen_controller.NewFactory(b))
	sr.AddFactory(link_establish_controller.NewFactory(b))

	// stream controllers
	sr.AddFactory(stream_forwarding.NewFactory(b))
	sr.AddFactory(stream_echo.NewFactory(b))
	sr.AddFactory(stream_listening.NewFactory(b))
	sr.AddFactory(stream_api_accept.NewFactory(b))

	// in-proc transport
	sr.AddFactory(iproctpt.NewFactory(b))
	// udp transport
	sr.AddFactory(udptpt.NewFactory(b))
	// websocket transport
	sr.AddFactory(wtpt.NewFactory(b))
	// webrtc transport
	sr.AddFactory(webrtc.NewFactory(b))

	// pubsub
	sr.AddFactory(pubsub_relay.NewFactory(b))
	sr.AddFactory(floodsub_controller.NewFactory(b))

	// entity graph
	sr.AddFactory(egc.NewFactory(b))
	sr.AddFactory(bifrosteg.NewFactory(b))

	// tptaddr
	sr.AddFactory(tptaddr_controller.NewFactory(b))
	sr.AddFactory(tptaddr_static.NewFactory(b))

	// http listener
	sr.AddFactory(http_listener.NewFactory(b))

	// signaling
	sr.AddFactory(signaling_rpc_client.NewFactory(b))
	sr.AddFactory(signaling_rpc_server.NewFactory(b))
}
