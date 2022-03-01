package bifrost_api

import (
	"context"
	"time"

	"github.com/aperturerobotics/bifrost/peer"
	stream_api "github.com/aperturerobotics/bifrost/stream/api"
	"github.com/aperturerobotics/controllerbus/bus"
	bus_api "github.com/aperturerobotics/controllerbus/bus/api"
	controller_exec "github.com/aperturerobotics/controllerbus/controller/exec"
	"github.com/pkg/errors"
	"storj.io/drpc"
)

// BusAPI implements the controller bus api.
type BusAPI = bus_api.API

// API implements the daemon API.
type API struct {
	*BusAPI

	bus  bus.Bus
	conf *Config
}

// NewAPI constructs a new instance of the API.
func NewAPI(bus bus.Bus, conf *Config) (*API, error) {
	return &API{
		BusAPI: bus_api.NewAPI(bus, conf.GetBusConfig()),

		bus:  bus,
		conf: conf,
	}, nil
}

// ForwardStreams forwards streams to the target multiaddress.
// Handles HandleMountedStream directives by contacting the target.
func (a *API) ForwardStreams(
	req *stream_api.ForwardStreamsRequest,
	serv stream_api.DRPCStreamService_ForwardStreamsStream,
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
		false,
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
				&stream_api.ForwardStreamsResponse{
					ControllerStatus: status,
				},
			)
		},
	)
}

// ListenStreams listens for streams on the multiaddress.
func (a *API) ListenStreams(
	req *stream_api.ListenStreamsRequest,
	serv stream_api.DRPCStreamService_ListenStreamsStream,
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
				&stream_api.ListenStreamsResponse{
					ControllerStatus: status,
				},
			)
		},
	)
}

// RegisterAsDRPCServer registers the API to the DRPC mux.
func (a *API) RegisterAsDRPCServer(mux drpc.Mux) {
	RegisterAsDRPCServer(a, mux)
}

// _ is a type assertion
var _ BifrostAPIServer = ((*API)(nil))
