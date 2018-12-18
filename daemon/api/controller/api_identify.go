package api_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/daemon/api"
	"github.com/aperturerobotics/controllerbus/controller/exec"
)

// Identify loads and manages a private key identity.
func (a *API) Identify(
	req *api.IdentifyRequest,
	serv api.BifrostDaemonService_IdentifyServer,
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
			_ = serv.Send(&api.IdentifyResponse{
				ControllerStatus: status,
			})
		},
	)
}
