package store

import (
	"context"
	"time"

	"github.com/aperturerobotics/util/broadcast"
)

const (
	// backgroundFetchSlots bounds the background fetches one store runs at
	// once, so they share the link with foreground fetches but never fill it.
	backgroundFetchSlots = 2
	// foregroundQuietWindow is how long after the last foreground fetch a
	// background fetch waits, so it stays out of the short gaps between a
	// program's sequential reads.
	foregroundQuietWindow = 200 * time.Millisecond
)

// demandGate admits remote fetches by class for every reader of one store.
//
// A foreground fetch enters at once and holds back background fetches while it
// runs and for the quiet window after it. A background fetch enters when no
// foreground fetch holds it back and a background slot is free, or at once
// when a foreground read joins it. A nil gate admits every fetch at once.
type demandGate struct {
	// quiet is the quiet window after a foreground fetch.
	quiet time.Duration
	// slots is the number of background fetches allowed at once.
	slots int

	// bcast guards the fields below and wakes waiting background fetches.
	bcast broadcast.Broadcast
	// foreground is the number of foreground fetches in flight.
	foreground int
	// background is the number of background fetches in flight.
	background int
	// lastForeground is when the last foreground fetch ended.
	lastForeground time.Time
}

// newDemandGate builds a gate with the default quiet window and slots, or nil
// when owner work runs inline and a fetch can never wait for another.
func newDemandGate() *demandGate {
	if !ownerWorkConcurrent {
		return nil
	}
	return &demandGate{quiet: foregroundQuietWindow, slots: backgroundFetchSlots}
}

// enter admits one fetch and returns the function that ends it. A background
// fetch waits until it may start, until promote closes, or until ctx ends.
func (g *demandGate) enter(ctx context.Context, background bool, promote <-chan struct{}) (func(), error) {
	// Admit every fetch at once without a gate.
	if g == nil {
		return func() {}, nil
	}

	// Admit by class.
	if !background {
		return g.enterForeground(), nil
	}
	return g.enterBackground(ctx, promote)
}

// enterForeground counts a foreground fetch until the returned function runs.
func (g *demandGate) enterForeground() func() {
	// Hold background fetches back while this fetch runs.
	g.bcast.HoldLock(func(_ func(), _ func() <-chan struct{}) {
		g.foreground++
	})

	// End the fetch and start its quiet window.
	return func() {
		g.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
			g.foreground--
			g.lastForeground = time.Now()
			broadcast()
		})
	}
}

// enterBackground waits until a background fetch may start, then counts it
// until the returned function runs. A closed promote admits it as foreground.
func (g *demandGate) enterBackground(ctx context.Context, promote <-chan struct{}) (func(), error) {
	for {
		// Take a slot once no foreground fetch is in flight or quieting.
		var admitted bool
		var delay time.Duration
		var wait <-chan struct{}
		g.bcast.HoldLock(func(_ func(), getWaitCh func() <-chan struct{}) {
			if g.foreground == 0 && g.background < g.slots {
				delay = g.quiet - time.Since(g.lastForeground)
				if delay <= 0 {
					g.background++
					admitted = true
					return
				}
			}
			wait = getWaitCh()
		})
		if admitted {
			return g.leaveBackground, nil
		}

		// Wait for a change, the end of the quiet window, promotion, or cancellation.
		var quietEnd <-chan time.Time
		var timer *time.Timer
		if delay > 0 {
			timer = time.NewTimer(delay)
			quietEnd = timer.C
		}
		var err error
		var promoted bool
		select {
		case <-ctx.Done():
			err = context.Cause(ctx)
		case <-promote:
			promoted = true
		case <-wait:
		case <-quietEnd:
		}
		if timer != nil {
			timer.Stop()
		}
		if err != nil {
			return nil, err
		}
		if promoted {
			return g.enterForeground(), nil
		}
	}
}

// leaveBackground ends a background fetch and wakes the fetches waiting for
// its slot.
func (g *demandGate) leaveBackground() {
	g.bcast.HoldLock(func(broadcast func(), _ func() <-chan struct{}) {
		g.background--
		broadcast()
	})
}
