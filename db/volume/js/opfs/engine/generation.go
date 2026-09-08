package engine

import (
	"context"
	"time"
)

// WaitGeneration waits for a durable generation newer than after.
// Broadcast hints only schedule reads. The bounded fallback recovers messages
// lost while a browser tab is suspended or its notification channel is absent.
func (e *Engine) WaitGeneration(ctx context.Context, after uint64) (uint64, error) {
	hints, release := e.backend.Subscribe()
	defer release()
	for {
		generation, err := e.RefreshGenerationContext(ctx)
		if err != nil || generation > after {
			return generation, err
		}
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return 0, ctx.Err()
		case <-e.done:
			timer.Stop()
			return 0, ErrClosed
		case <-hints:
			timer.Stop()
		case <-timer.C:
		}
	}
}
