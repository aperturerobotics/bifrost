package bifrost_http

import (
	"context"
	"net/http"
)

// HTTPHandlerBuilder builds a HTTP Handle.
//
// returns the http handler and an optional release function
// can return nil to indicate not found.
type HTTPHandlerBuilder func(ctx context.Context, released func()) (*http.Handler, func(), error)

// NewHTTPHandlerBuilder creates a new HTTPHandlerBuilder with a static handler.
func NewHTTPHandlerBuilder(handler http.Handler) HTTPHandlerBuilder {
	return func(ctx context.Context, released func()) (*http.Handler, func(), error) {
		if handler == nil {
			return nil, nil, nil
		}
		return &handler, nil, nil
	}
}
