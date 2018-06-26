package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
)

// DefaultBackoff returns the default backoff.
func DefaultBackoff() backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 500 * time.Millisecond
	b.Multiplier = 1.7
	b.MaxInterval = time.Duration(10) * time.Second
	return b
}

// Retry uses a backoff to re-try a process.
// If the process returns nil or context canceled, it exits.
// If bo is nil, a default one is created.
// The defaults are: 500Ms initial backoff,
func Retry(
	ctx context.Context,
	le *logrus.Entry,
	f func(context.Context) error,
	bo backoff.BackOff,
) error {
	if bo == nil {
		bo = DefaultBackoff()
	}

	for {
		le.Debug("starting process")
		err := f(ctx)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err == nil {
			return nil
		}

		b := bo.NextBackOff()
		le.
			WithError(err).
			WithField("backoff", b.String()).
			Warn("process failed, retrying")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(b):
		}
	}
}
