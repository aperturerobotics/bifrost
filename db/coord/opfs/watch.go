//go:build js

package opfs

import (
	"context"
	"sync/atomic"

	"github.com/s4wave/spacewave/db/coord"
)

// watch combines detailed local events with durable cross-runtime invalidations.
type watch struct {
	// ctx bounds delivery and both watched event sources.
	ctx context.Context
	// cancel interrupts queued delivery and durable generation waits.
	cancel context.CancelFunc
	// c supplies the authoritative generation source.
	c *Coordinator
	// scope identifies the watched volume and object store.
	scope coord.Scope
	// inner supplies local root and prefix events.
	inner coord.Watch
	// after excludes already observed durable revisions.
	after uint64
	// events is the public bounded delivery channel.
	events chan coord.Event
	// done joins the combined watch and its generation waiter.
	done chan struct{}
	// once makes source cancellation idempotent.
	once atomic.Bool
}

// Events yields ordered local events and durable invalidation hints.
func (w *watch) Events() <-chan coord.Event {
	return w.events
}

// Close cancels event sources and joins the delivery loop.
func (w *watch) Close() error {
	var err error
	if w.once.CompareAndSwap(false, true) {
		w.cancel()
		err = w.inner.Close()
		<-w.done
	}
	return err
}

// start owns the delivery loop and joins its generation waiter on every exit.
func (w *watch) start() {
	go func() {
		defer close(w.done)
		defer close(w.events)

		// Join the storage owner's durable generation waiter on every exit.
		ctx, cancel := context.WithCancel(w.ctx)
		generations := make(chan uint64)
		generationDone := make(chan struct{})
		go func() {
			defer close(generationDone)
			defer close(generations)
			if w.c.meta == nil {
				return
			}
			after := w.after
			for {
				next, err := w.c.meta.WaitGeneration(ctx, after)
				if err != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case generations <- next:
				}
				after = next
			}
		}()
		defer func() {
			cancel()
			<-generationDone
		}()
		innerEvents := w.inner.Events()
		for {
			select {
			case <-w.ctx.Done():
				return
			case event, ok := <-innerEvents:
				if !ok {
					innerEvents = nil
					continue
				}
				w.send(event)
			case generation, ok := <-generations:
				if !ok {
					generations = nil
					continue
				}
				w.send(coord.Event{
					Generation:    generation,
					VolumeID:      w.scope.VolumeID,
					ObjectStoreID: w.scope.ObjectStoreID,
				})
			}
			if innerEvents == nil {
				return
			}
		}
	}()
}

// send resolves unstamped local events before cancellation-aware delivery.
func (w *watch) send(event coord.Event) {
	if event.Generation == 0 {
		generation, err := w.c.generation(w.ctx, w.scope)
		if err != nil {
			return
		}
		event.Generation = generation
	}
	select {
	case <-w.ctx.Done():
	case w.events <- event:
	}
}

// _ is a type assertion
var _ coord.Watch = (*watch)(nil)
