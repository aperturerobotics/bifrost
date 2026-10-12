package s4wave_objecttype_registry

import (
	"context"

	"github.com/aperturerobotics/starpc/srpc"
	resource_server "github.com/s4wave/spacewave/bldr/resource/server"
	"github.com/s4wave/spacewave/sdk/world/objecttype"
)

// CallerInvoker retains a typed dispatch caller through descendant RPCs.
type CallerInvoker struct {
	// invoker serves one Resource's methods.
	invoker srpc.Invoker
	// caller is the immutable authenticated identity of the granting dispatch.
	caller objecttype.Caller
}

// NewCallerInvoker wraps a Resource under its granting caller's identity.
func NewCallerInvoker(invoker srpc.Invoker, caller objecttype.Caller) *CallerInvoker {
	caller.Participant = caller.Participant.CloneVT()
	return &CallerInvoker{invoker: invoker, caller: caller}
}

// InvokeMethod binds the granting caller without changing the RPC lifetime.
func (i *CallerInvoker) InvokeMethod(serviceID, methodID string, stream srpc.Stream) (bool, error) {
	ctx := bindCallerResources(stream.Context(), i.caller)
	return i.invoker.InvokeMethod(serviceID, methodID, srpc.NewStreamWithContext(stream, ctx))
}

// bindCallerResources preserves caller identity through child construction and registration.
func bindCallerResources(ctx context.Context, caller objecttype.Caller) context.Context {
	ctx = objecttype.WithCaller(ctx, caller)
	if resources := resource_server.GetResourceClientContext(ctx); resources != nil {
		ctx = resource_server.WithResourceClientContext(ctx, &callerResources{resources: resources, caller: caller})
	}
	return ctx
}

// _ is a type assertion.
var _ srpc.Invoker = (*CallerInvoker)(nil)
