package bifrost_rpc_access

import (
	context "context"
	"errors"

	bifrost_rpc "github.com/aperturerobotics/bifrost/rpc"
	"github.com/aperturerobotics/controllerbus/directive"
)

// LookupRpcServiceResolver resolves a LookupRpcService directive with a RPC service.
type LookupRpcServiceResolver struct {
	dir bifrost_rpc.LookupRpcService
	svc BuildClientFunc
}

// NewLookupRpcServiceResolver constructs the directive resolver.
func NewLookupRpcServiceResolver(
	dir bifrost_rpc.LookupRpcService,
	svc BuildClientFunc,
) *LookupRpcServiceResolver {
	return &LookupRpcServiceResolver{dir: dir, svc: svc}
}

// Resolve resolves the values, emitting them to the handler.
func (r *LookupRpcServiceResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	req := RequestFromDirective(r.dir)
	clientPtr, clientRel, err := r.svc(ctx)
	if clientRel != nil {
		defer clientRel()
	}
	if err == nil && clientPtr == nil {
		return errors.New("client constructor returned nil")
	}
	if err != nil {
		return err
	}

	client := *clientPtr
	strm, err := client.LookupRpcService(ctx, req)
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
			var val bifrost_rpc.LookupRpcServiceValue = NewProxyInvoker(client, req)
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
