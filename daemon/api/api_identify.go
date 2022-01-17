package bifrost_api

import (
	"context"

	peer_grpc "github.com/aperturerobotics/bifrost/peer/grpc"
	controller_exec "github.com/aperturerobotics/controllerbus/controller/exec"
)

// Identify loads and manages a private key identity.
func (a *API) Identify(
	req *peer_grpc.IdentifyRequest,
	serv peer_grpc.PeerService_IdentifyServer,
) error {
	ctx := serv.Context()
	conf := req.GetConfig()
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
			_ = serv.Send(&peer_grpc.IdentifyResponse{
				ControllerStatus: status,
			})
		},
	)
}
