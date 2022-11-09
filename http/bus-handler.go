package bifrost_http

import (
	"net/http"

	"github.com/aperturerobotics/controllerbus/bus"
)

// BusHandler implements http.Handler by calling LookupHTTPHandler.
type BusHandler struct {
	// b is the bus to use for lookups
	b bus.Bus
	// clientID is the client id to use for lookups
	clientID string
	// notFoundIfIdle indicates to return 404 not found if the lookup is idle
	notFoundIfIdle bool
}

// NewBusHandler constructs a new bus-backed HTTP handler.
func NewBusHandler(b bus.Bus, clientID string, notFoundIfIdle bool) *BusHandler {
	return &BusHandler{
		b:              b,
		clientID:       clientID,
		notFoundIfIdle: notFoundIfIdle,
	}
}

// ServeHTTP serves the http request.
func (h *BusHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	handler, handlerRef, err := ExLookupFirstHTTPHandler(ctx, h.b, req.URL.String(), "", h.notFoundIfIdle)
	if err != nil {
		rw.WriteHeader(500)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}
	if handlerRef == nil {
		rw.WriteHeader(404)
		_, _ = rw.Write([]byte("bldr: handler not found for url"))
		return
	}

	defer handlerRef.Release()
	handler.ServeHTTP(rw, req)
}

// _ is a type assertion
var _ http.Handler = ((*BusHandler)(nil))
