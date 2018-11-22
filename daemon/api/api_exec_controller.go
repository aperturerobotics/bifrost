//+build !js

package api

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller/resolver"
)

// executeController executes a controller and calls the callback with state.
func (a *API) executeController(
	ctx context.Context,
	conf config.Config,
	cb func(ControllerStatus),
) error {
	if cb == nil {
		cb = func(ControllerStatus) {}
	}
	dir := resolver.NewLoadControllerWithConfigSingleton(conf)

	cb(ControllerStatus_ControllerStatus_CONFIGURING)
	_, valRef, err := bus.ExecOneOff(ctx, a.bus, dir, nil)
	if err != nil {
		return err
	}
	defer valRef.Release()

	cb(ControllerStatus_ControllerStatus_RUNNING)
	<-ctx.Done()
	return nil
}
