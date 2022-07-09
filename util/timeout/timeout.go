package timeout

import (
	"context"
	"time"
)

// BuildTimeoutCtx builds a context with the configured timeout.
//
// Checks if the timeout duration is <= 0: if so, returns w/o timeout.
func BuildTimeoutCtx(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	if duration <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, duration)
}
