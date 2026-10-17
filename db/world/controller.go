package world

import (
	"context"

	"github.com/aperturerobotics/controllerbus/controller"
)

// Controller is a Hydra World controller managing an Engine.
type Controller interface {
	// Controller indicates this is a controller-bus controller.
	controller.Controller
	// GetWorldEngine waits for the engine to be built.
	GetWorldEngine(context.Context) (Engine, error)
}

// ReplacingController is a Controller that can replace its engine while it
// runs, such as after losing and regaining read access.
type ReplacingController interface {
	Controller
	// WaitWorldEngineReplaced waits until eng is no longer the engine the
	// controller serves.
	WaitWorldEngineReplaced(ctx context.Context, eng Engine) error
}
