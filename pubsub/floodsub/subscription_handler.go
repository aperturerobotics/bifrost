package floodsub

import (
	"github.com/aperturerobotics/bifrost/pubsub"
)

// subscriptionHandler contains a handler added with AddHandler
type subscriptionHandler struct {
	cb func(m pubsub.Message)
}
