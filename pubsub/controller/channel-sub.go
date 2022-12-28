package pubsub_controller

import (
	"context"

	"github.com/aperturerobotics/bifrost/pubsub"
	"github.com/aperturerobotics/controllerbus/directive"
)

// resolveBuildChannelSub is the BuildChannelSubscription resolver.
type resolveBuildChannelSub struct {
	ctx context.Context
	c   *Controller
	di  directive.Instance
	d   pubsub.BuildChannelSubscription
}

// emittedSubscription is a wrapped subscription with proxied calls
type emittedSubscription struct {
	pubsub.Subscription
	relFunc func()
}

// Release releases the subscription handle, clearing the handlers.
func (e *emittedSubscription) Release() {
	if e.relFunc != nil {
		e.relFunc()
	}
}

// resolveBuildChannelSub resolves building a channel subscription.
func (c *Controller) resolveBuildChannelSub(
	ctx context.Context,
	di directive.Instance,
	d pubsub.BuildChannelSubscription,
) ([]directive.Resolver, error) {
	// accept directive always
	return directive.Resolvers(&resolveBuildChannelSub{
		ctx: ctx,
		c:   c,
		di:  di,
		d:   d,
	}), nil
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (r *resolveBuildChannelSub) Resolve(
	ctx context.Context,
	handler directive.ResolverHandler,
) error {
	handler.ClearValues()
	ps, err := r.c.GetPubSub(ctx)
	if err != nil {
		return err
	}

	sub, err := ps.AddSubscription(r.ctx, r.d.BuildChannelSubscriptionPrivKey(), r.d.BuildChannelSubscriptionChannelID())
	if err != nil {
		return err
	}

	var valueID uint32
	var accepted bool
	var val pubsub.BuildChannelSubscriptionValue = &emittedSubscription{
		Subscription: sub,
		relFunc: func() {
			_, _ = handler.RemoveValue(valueID)
			sub.Release()
		},
	}
	valueID, accepted = handler.AddValue(val)
	if !accepted {
		val.Release()
		return nil
	}
	r.di.AddDisposeCallback(func() {
		val.Release()
	})
	return nil
}

// _ is a type assertion
var _ pubsub.Subscription = ((*emittedSubscription)(nil))

// _ is a type assertion
var _ directive.Resolver = ((*resolveBuildChannelSub)(nil))
