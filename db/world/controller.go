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
// runs, such as after losing and regaining read access. It closes a replaced
// engine only after every holder released it, so a lookup that withdraws the
// engine before releasing it never publishes a closed engine.
type ReplacingController interface {
	Controller
	// AcquireWorldEngine waits for the served engine and holds it open until
	// release is called.
	AcquireWorldEngine(ctx context.Context) (eng Engine, release func(), err error)
	// WaitWorldEngineReplaced waits until eng is no longer the engine the
	// controller serves.
	WaitWorldEngineReplaced(ctx context.Context, eng Engine) error
}
