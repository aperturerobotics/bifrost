package signaling_rpc_server

import (
	"github.com/aperturerobotics/bifrost/protocol"
	signaling "github.com/aperturerobotics/bifrost/signaling/rpc"
	signaling_rpc "github.com/aperturerobotics/bifrost/signaling/rpc"
	stream_srpc_server "github.com/aperturerobotics/bifrost/stream/srpc/server"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/starpc/srpc"
	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "bifrost/signaling/rpc/server"

// Controller is the signaling server controller.
type Controller struct {
	*stream_srpc_server.Server
	// srv is the server
	srv *Server
}

// NewController constructs a new signaling server controller.
func NewController(le *logrus.Entry, b bus.Bus, c *Config) (*Controller, error) {
	srvConf := c.GetServer().ApplyDefaults([]protocol.ID{signaling_rpc.ProtocolID})

	var err error
	srv := NewServer(le)
	ctrl := &Controller{srv: srv}
	ctrl.Server, err = srvConf.BuildServer(
		b,
		le,
		controller.NewInfo(ControllerID, Version, "signaling server"),
		[]stream_srpc_server.RegisterFn{
			func(mux srpc.Mux) error {
				return signaling.SRPCRegisterSignaling(mux, srv)
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return ctrl, nil
}

// _ is a type assertion
var _ controller.Controller = ((*Controller)(nil))
