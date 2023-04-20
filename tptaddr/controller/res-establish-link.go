package tptaddr_controller

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/aperturerobotics/bifrost/link"
	"github.com/aperturerobotics/bifrost/tptaddr"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/util/broadcast"
)

// establishLinkResolver resolves establishLink directives
type establishLinkResolver struct {
	c   *Controller
	ctx context.Context
	di  directive.Instance
	dir link.EstablishLinkWithPeer
}

// resolveEstablishLinkWithPeer resolves a EstablishLinkWithPeer directive.
func (c *Controller) resolveEstablishLinkWithPeer(
	ctx context.Context,
	di directive.Instance,
	dir link.EstablishLinkWithPeer,
) ([]directive.Resolver, error) {
	return directive.R(&establishLinkResolver{
		c:   c,
		ctx: ctx,
		di:  di,
		dir: dir,
	}, nil)
}

// Resolve resolves the values, emitting them to the handler.
// The resolver may be canceled and restarted multiple times.
// Any fatal error resolving the value is returned.
// The resolver will not be retried after returning an error.
// Values will be maintained from the previous call.
func (o *establishLinkResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	var mtx sync.Mutex
	var bcast broadcast.Broadcast
	var lookupIdle bool
	var incomingDialers []string
	var delDialers []string

	// Create LookupTptAddr directive.
	// When a new address is added, add a directive to dial that address.
	lookupDi, lookupRef, err := o.c.bus.AddDirective(
		tptaddr.NewLookupTptAddr(o.dir.EstablishLinkTargetPeerId()),
		bus.NewCallbackHandler(func(av directive.AttachedValue) {
			val, ok := av.GetValue().(tptaddr.LookupTptAddrValue)
			if !ok || val == "" {
				return
			}
			mtx.Lock()
			incomingDialers = append(incomingDialers, val)
			bcast.Broadcast()
			mtx.Unlock()
		}, func(av directive.AttachedValue) {
			val, ok := av.GetValue().(tptaddr.LookupTptAddrValue)
			if !ok || val == "" {
				return
			}
			mtx.Lock()
			delDialers = append(delDialers, val)
			bcast.Broadcast()
			mtx.Unlock()
		}, nil),
	)
	if err != nil {
		return err
	}
	defer lookupRef.Release()

	// handle lookup becoming idle
	defer lookupDi.AddIdleCallback(func(resErrs []error) {
		mtx.Lock()
		if !lookupIdle {
			lookupIdle = true
			bcast.Broadcast()
		}
		mtx.Unlock()
	})()

	// The below fields are controlled by the below loop.
	type dialerInfo struct {
		// n is the number of references
		n int
		// relRef releases the reference
		relRef func()
		// idle indicates this dialer is idle
		idle atomic.Bool
	}
	dialers := make(map[string]*dialerInfo)

	defer func() {
		mtx.Lock()
		for k, dialer := range dialers {
			if dialer.relRef != nil {
				dialer.relRef()
			}
			delete(dialers, k)
		}
		mtx.Unlock()
	}()

	for {
		mtx.Lock()
		wait := bcast.GetWaitCh()
		for i := range incomingDialers {
			incomingTptAddr := incomingDialers[i]
			info := dialers[incomingTptAddr]
			if info == nil {
				info = &dialerInfo{n: 1}
				dialers[incomingTptAddr] = info
				dialInst, dialRef, err := o.c.bus.AddDirective(tptaddr.NewDialTptAddr(
					incomingTptAddr,
					o.dir.EstablishLinkSourcePeerId(),
					o.dir.EstablishLinkTargetPeerId(),
				), nil)
				if err != nil {
					o.c.le.WithError(err).Warn("unable to dial transport address")
				} else {
					relIdleCb := dialInst.AddIdleCallback(func(_ []error) {
						if !info.idle.Swap(true) {
							bcast.Broadcast()
						}
					})
					info.relRef = func() {
						relIdleCb()
						dialRef.Release()
					}
				}
			} else {
				info.n++
			}
		}
		incomingDialers = nil

		for i := range delDialers {
			delDialerID := delDialers[i]
			info, ok := dialers[delDialerID]
			if !ok {
				continue
			}
			info.n--
			if info.n <= 0 {
				if info.relRef != nil {
					info.relRef()
				}
				info.idle.Store(true)
				delete(dialers, delDialerID)
			}
		}
		delDialers = nil

		// check if everything is idle
		if lookupIdle {
			allDialersIdle := true
			for _, dialer := range dialers {
				if !dialer.idle.Load() {
					allDialersIdle = false
					break
				}
			}
			if allDialersIdle {
				handler.MarkIdle()
			}
		}
		mtx.Unlock()

		select {
		case <-ctx.Done():
			return context.Canceled
		case <-wait:
		}
	}
}

// _ is a type assertion
var _ directive.Resolver = ((*establishLinkResolver)(nil))
