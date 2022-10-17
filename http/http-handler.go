package bifrost_http

import (
	"context"
	"net/http"

	"github.com/aperturerobotics/controllerbus/util/ccontainer"
	"github.com/aperturerobotics/controllerbus/util/refcount"
)

// HTTPHandler implements a HTTP handler which deduplicates with a reference count.
type HTTPHandler struct {
	// handleCtr is the refcount handle to the UnixFS
	handleCtr *ccontainer.CContainer[*http.Handler]
	// errCtr contains any error building FSHandle
	errCtr *ccontainer.CContainer[*error]
	// rc is the refcount container
	rc *refcount.RefCount[*http.Handler]
}

// NewHTTPHandler constructs a new HTTPHandler.
//
// NOTE: if ctx == nil the handler won't work until SetContext is called.
func NewHTTPHandler(
	ctx context.Context,
	builder HTTPHandlerBuilder,
) *HTTPHandler {
	h := &HTTPHandler{
		handleCtr: ccontainer.NewCContainer[*http.Handler](nil),
		errCtr:    ccontainer.NewCContainer[*error](nil),
	}
	h.rc = refcount.NewRefCount(ctx, h.handleCtr, h.errCtr, builder)
	return h
}

// SetContext sets the context for the HTTPHandler.
func (h *HTTPHandler) SetContext(ctx context.Context) {
	h.rc.SetContext(ctx)
}

// ServeHTTP serves a http request.
func (h *HTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	err := refcount.AccessRefCount(ctx, h.rc, func(access *http.Handler) error {
		if access == nil {
			rw.WriteHeader(404)
			_, _ = rw.Write([]byte("404 not found"))
			return nil
		}

		(*access).ServeHTTP(rw, req)
		return nil
	})
	if err != nil {
		rw.WriteHeader(500)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}
}

// _ is a type assertion
var _ http.Handler = ((*HTTPHandler)(nil))
