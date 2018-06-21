package controller

import (
	"context"
)

// Controller tracks a particular process.
type Controller interface {
	// Execute executes the given controller.
	// Returning nil ends execution.
	// Returning an error triggers a retry with backoff.
	Execute(ctx context.Context) error
}
