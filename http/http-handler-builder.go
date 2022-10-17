package bifrost_http

import (
	"context"
	"net/http"

	"github.com/aperturerobotics/controllerbus/directive"
)

// HTTPHandlerBuilder builds a HTTP Handle.
//
// returns the http handler and an optional release function
// can return nil to indicate not found.
type HTTPHandlerBuilder func(ctx context.Context) (*http.Handler, func(), error)

// NewHTTPHandlerBuilder creates a new HTTPHandlerBuilder with a static handler.
func NewHTTPHandlerBuilder(handler http.Handler) HTTPHandlerBuilder {
	return func(ctx context.Context) (*http.Handler, func(), error) {
		if handler == nil {
			return nil, nil, nil
		}
		return &handler, nil, nil
	}
}

// LookupHTTPHandlerBuilderResolver resolves LookupHTTPHandler with a handler builder.
type LookupHTTPHandlerBuilderResolver struct {
	inst     directive.Instance
	resolver HTTPHandlerBuilder
}

// NewLookupHTTPHandlerBuilderResolver constructs a new resolver.
func NewLookupHTTPHandlerBuilderResolver(di directive.Instance, resolver HTTPHandlerBuilder) *LookupHTTPHandlerBuilderResolver {
	if resolver == nil {
		return nil
	}
	return &LookupHTTPHandlerBuilderResolver{inst: di, resolver: resolver}
}

// Resolve resolves the values, emitting them to the handler.
func (r *LookupHTTPHandlerBuilderResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	hhandler, hhandlerRel, err := r.resolver(ctx)
	if err != nil {
		return err
	}
	if hhandler == nil || *hhandler == nil {
		return nil
	}
	_, accepted := handler.AddValue(*hhandler)
	if !accepted {
		if hhandlerRel != nil {
			hhandlerRel()
		}
		return nil
	}
	if hhandlerRel != nil {
		r.inst.AddDisposeCallback(hhandlerRel)
	}
	return nil
}

// _ is a type assertion
var _ directive.Resolver = ((*LookupHTTPHandlerBuilderResolver)(nil))
