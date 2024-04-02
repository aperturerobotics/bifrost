package bifrost_http

import (
	"context"
	"net/http"

	"github.com/aperturerobotics/controllerbus/bus"
)

// NewBusHTTPHandlerBuilder constructs a HTTPHandlerBuilder which looks up the handler on the bus.
func NewBusHTTPHandlerBuilder(b bus.Bus, baseURL, clientID string, notFoundIfIdle bool) HTTPHandlerBuilder {
	return func(ctx context.Context, released func()) (http.Handler, func(), error) {
		handler, _, handlerRef, err := ExLookupFirstHTTPHandler(
			ctx,
			b,
			baseURL,
			clientID,
			notFoundIfIdle,
			released,
		)
		if err != nil {
			return nil, nil, err
		}
		if handlerRef == nil {
			return nil, nil, nil
		}
		return handler, handlerRef.Release, nil
	}
}

// NewBusHTTPHandler constructs a HTTPHandler which looks up the HTTP
// handler on the bus when at least one request is active.
//
// baseURL is the URL to use for the client lookup.
func NewBusHTTPHandler(
	ctx context.Context,
	b bus.Bus,
	baseURL, clientID string,
	notFoundIfIdle bool,
) *HTTPHandler {
	return NewHTTPHandler(
		ctx,
		NewBusHTTPHandlerBuilder(
			b,
			baseURL, clientID,
			notFoundIfIdle,
		))
}
