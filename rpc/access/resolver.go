package bifrost_rpc_access

import (
	context "context"

	bifrost_rpc "github.com/aperturerobotics/bifrost/rpc"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LookupRpcServiceResolver resolves a LookupRpcService directive with a RPC service.
type LookupRpcServiceResolver struct {
	dir bifrost_rpc.LookupRpcService
	svc SRPCAccessRpcServiceClient
}

// NewLookupRpcServiceResolver constructs the directive resolver.
func NewLookupRpcServiceResolver(
	dir bifrost_rpc.LookupRpcService,
	svc SRPCAccessRpcServiceClient,
) *LookupRpcServiceResolver {
	return &LookupRpcServiceResolver{dir: dir, svc: svc}
}

// Resolve resolves the values, emitting them to the handler.
func (r *LookupRpcServiceResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	req := RequestFromDirective(r.dir)
	strm, err := r.svc.LookupRpcService(ctx, req)
	if err != nil {
		return err
	}

	var valID uint32
	for {
		resp, err := strm.Recv()
		if err != nil {
			return err
		}
		if exists := resp.GetExists(); exists && valID == 0 {
			var val bifrost_rpc.LookupRpcServiceValue = NewProxyInvoker(r.svc, req)
			valID, _ = handler.AddValue(val)
		}
		if removed := resp.GetRemoved(); removed && valID != 0 {
			_, _ = handler.RemoveValue(valID)
		}
		if resp.GetIdle() {
			handler.MarkIdle()
		}
	}
}

// _ is a type assertion
var _ directive.Resolver = ((*LookupRpcServiceResolver)(nil))
