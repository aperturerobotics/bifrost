package bifrost_http

import (
	"context"
	"net/http"

	"github.com/aperturerobotics/controllerbus/util/ccontainer"
	"github.com/aperturerobotics/controllerbus/util/refcount"
)

// RefCountHandler constructs a http.Handler when it has at least one reference.
type RefCountHandler struct {
	// handleCtr is the refcount handle to the UnixFS
	handleCtr *ccontainer.CContainer[*http.Handler]
	// errCtr contains any error building FSHandle
	errCtr *ccontainer.CContainer[*error]
	// rc is the refcount container
	rc *refcount.RefCount[*http.Handler]
}

// NewRefCountHandler constructs a new RefCountHandler.
func NewRefCountHandler(
	ctx context.Context,
	resolver HTTPHandlerBuilder,
) *RefCountHandler {
	h := &RefCountHandler{
		handleCtr: ccontainer.NewCContainer[*http.Handler](nil),
		errCtr:    ccontainer.NewCContainer[*error](nil),
	}
	h.rc = refcount.NewRefCount(ctx, h.handleCtr, h.errCtr, resolver)
	return h
}

// ServeHTTP serves a http request.
func (h *RefCountHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	err := refcount.AccessRefCount(ctx, h.rc, func(handler *http.Handler) error {
		if handler == nil || *handler == nil {
			rw.WriteHeader(404)
			_, _ = rw.Write([]byte("404 not found"))
			return nil
		}
		(*handler).ServeHTTP(rw, req)
		return nil
	})
	if err != nil {
		rw.WriteHeader(500)
		_, _ = rw.Write([]byte(err.Error()))
	}
}

// _ is a type assertion
var _ http.Handler = ((*RefCountHandler)(nil))
