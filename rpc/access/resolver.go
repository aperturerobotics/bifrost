package bifrost_rpc_access

import (
	context "context"

	bifrost_rpc "github.com/aperturerobotics/bifrost/rpc"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/sirupsen/logrus"
)

// LookupRpcServiceResolver resolves a LookupRpcService directive with a RPC service.
type LookupRpcServiceResolver struct {
	dir     bifrost_rpc.LookupRpcService
	svc     AccessClientFunc
	waitAck bool
}

// NewLookupRpcServiceResolver constructs the directive resolver.
//
// if waitAck is set, waits for ack from the remote before starting the proxied rpc.
// note: usually you do not need waitAck set to true.
func NewLookupRpcServiceResolver(
	dir bifrost_rpc.LookupRpcService,
	svc AccessClientFunc,
	waitAck bool,
) *LookupRpcServiceResolver {
	return &LookupRpcServiceResolver{dir: dir, svc: svc, waitAck: waitAck}
}

// Resolve resolves the values, emitting them to the handler.
func (r *LookupRpcServiceResolver) Resolve(ctx context.Context, handler directive.ResolverHandler) error {
	req := RequestFromDirective(r.dir)
	defer handler.ClearValues()
	return r.svc(ctx, func(ctx context.Context, client SRPCAccessRpcServiceClient) error {
		handler.ClearValues()
		strm, err := client.LookupRpcService(ctx, req)
		if err != nil {
			return err
		}

		var valID uint32
		for {
			resp, err := strm.Recv()
			if err != nil {
				handler.ClearValues()
				return err
			}

			if exists := resp.GetExists(); exists && valID == 0 {
				var val bifrost_rpc.LookupRpcServiceValue = NewProxyInvoker(client, req, r.waitAck)
				valID, _ = handler.AddValue(val)
			}
			if removed := resp.GetRemoved(); removed && valID != 0 {
				_, _ = handler.RemoveValue(valID)
			}
			if resp.GetIdle() {
				handler.MarkIdle()
			}
		}
	})

}

// _ is a type assertion
var _ directive.Resolver = ((*LookupRpcServiceResolver)(nil))
