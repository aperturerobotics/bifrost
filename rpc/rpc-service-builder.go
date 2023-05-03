package bifrost_rpc

import (
	"context"

	"github.com/aperturerobotics/starpc/srpc"
)

// RpcServiceBuilder builds a rpc service invoker.
//
// returns the srpc invoker and an optional release function
// can return nil to indicate not found.
type RpcServiceBuilder func(ctx context.Context, released func()) (*srpc.Invoker, func(), error)

// NewRpcServiceBuilder creates a new RpcServiceBuilder with a static invoker.
func NewRpcServiceBuilder(handler srpc.Invoker) RpcServiceBuilder {
	return func(ctx context.Context, released func()) (*srpc.Invoker, func(), error) {
		if handler == nil {
			return nil, nil, nil
		}
		return &handler, nil, nil
	}
}
