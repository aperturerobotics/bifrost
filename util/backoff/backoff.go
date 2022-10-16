package backoff

import (
	"time"

	"github.com/cenkalti/backoff"
)

// Stop indicates that no more retries should be made for use in NextBackOff().
const Stop = backoff.Stop

// GetEmpty returns if the backoff config is empty.
func (b *Backoff) GetEmpty() bool {
	return b.GetBackoffKind() == 0
}

// Construct constructs the backoff.
// Validates the options.
func (b *Backoff) Construct() backoff.BackOff {
	switch b.GetBackoffKind() {
	default:
		fallthrough
	case BackoffKind_BackoffKind_EXPONENTIAL:
		return b.constructExpo()
	case BackoffKind_BackoffKind_CONSTANT:
		return b.constructConstant()
	}
}

// constructExpo constructs an exponential backoff.
func (b *Backoff) constructExpo() backoff.BackOff {
	expo := backoff.NewExponentialBackOff()
	opts := b.GetExponential()

	initialInterval := opts.GetInitialInterval()
	if initialInterval == 0 {
		// default to 800ms
		initialInterval = 800
	}
	expo.InitialInterval = time.Duration(initialInterval) * time.Millisecond

	multiplier := opts.GetMultiplier()
	if multiplier == 0 {
		multiplier = 1.8
	}
	expo.Multiplier = float64(multiplier)

	maxInterval := opts.GetMaxInterval()
	if maxInterval == 0 {
		maxInterval = 20000
	}
	expo.MaxInterval = time.Duration(maxInterval) * time.Millisecond
	expo.RandomizationFactor = float64(opts.GetRandomizationFactor())
	if opts.GetMaxElapsedTime() == 0 {
		expo.MaxElapsedTime = 0
	} else {
		expo.MaxElapsedTime = time.Duration(opts.GetMaxElapsedTime()) * time.Millisecond
	}
	return expo
}

// constructConstant constructs a constant backoff.
func (b *Backoff) constructConstant() backoff.BackOff {
	dur := b.GetConstant().GetInterval()
	if dur == 0 {
		dur = 5000
	}
	return backoff.NewConstantBackOff(time.Duration(dur) * time.Millisecond)
}
