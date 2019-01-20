package floodsub_controller

import (
	"github.com/aperturerobotics/bifrost/pubsub/controller"
	"github.com/blang/semver"
)

// Version is the version of the controller implementation.
var Version = semver.MustParse("0.0.1")

// ControllerID is the ID of the controller.
const ControllerID = "bifrost/floodsub/1"

// Controller implements the FloodSub controller.
type Controller = pubsub_controller.Controller
