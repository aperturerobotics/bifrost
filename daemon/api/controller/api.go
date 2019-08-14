package bifrost_api_controller

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/bifrost/peer"
	"github.com/aperturerobotics/bifrost/stream/grpc"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/controller/exec"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// API implements the GRPC API.
type API struct {
	bus bus.Bus
}

// NewAPI constructs a new instance of the API.
func NewAPI(bus bus.Bus) (*API, error) {
	return &API{bus: bus}, nil
}

// ForwardStreams forwards streams to the target multiaddress.
// Handles HandleMountedStream directives by contacting the target.
func (a *API) ForwardStreams(
	req *stream_grpc.ForwardStreamsRequest,
	serv stream_grpc.StreamService_ForwardStreamsServer,
) error {
	ctx := serv.Context()
	conf := req.GetForwardingConfig()
	if err := conf.Validate(); err != nil {
		return err
	}

	targetPeerID, err := req.GetForwardingConfig().ParsePeerID()
	if err != nil {
		return err
	}

	reqCtx, reqCtxCancel := context.WithCancel(ctx)
	defer reqCtxCancel()

	plCtx, plCtxCancel := context.WithTimeout(reqCtx, time.Second*3)
	defer plCtxCancel()

	// if the peer is unloaded the request will be canceled.
	_, peerRef, err := bus.ExecOneOff(
		plCtx,
		a.bus,
		peer.NewGetPeer(targetPeerID),
		reqCtxCancel,
	)
	if err != nil {
		return errors.Errorf("peer not loaded: %s", targetPeerID.Pretty())
	}
	defer peerRef.Release()

	return controller_exec.ExecuteController(
		reqCtx,
		a.bus,
		conf,
		func(status controller_exec.ControllerStatus) {
			_ = serv.Send(
				&stream_grpc.ForwardStreamsResponse{
					ControllerStatus: status,
				},
			)
		},
	)
}

// ListenStreams listens for streams on the multiaddress.
func (a *API) ListenStreams(
	req *stream_grpc.ListenStreamsRequest,
	serv stream_grpc.StreamService_ListenStreamsServer,
) error {
	ctx := serv.Context()
	conf := req.GetListeningConfig()
	if err := conf.Validate(); err != nil {
		return err
	}

	reqCtx, reqCtxCancel := context.WithCancel(ctx)
	defer reqCtxCancel()

	return controller_exec.ExecuteController(
		reqCtx,
		a.bus,
		conf,
		func(status controller_exec.ControllerStatus) {
			_ = serv.Send(
				&stream_grpc.ListenStreamsResponse{
					ControllerStatus: status,
				},
			)
		},
	)
}

// RegisterAsGRPCServer registers the API to the GRPC instance.
func (a *API) RegisterAsGRPCServer(grpcServer *grpc.Server) {
	api.RegisterAsGRPCServer(a, grpcServer)
}

// _ is a type assertion
var _ api.BifrostDaemonServer = ((*API)(nil))
