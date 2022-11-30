package bifrost_rpc

import (
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/starpc/srpc"
)

// Invoker implements the RPC invoker with a directive.
type Invoker struct {
	// b is the bus
	b bus.Bus
	// serverID is the server identifier.
	// can be empty
	serverID string
}

// NewInvoker constructs a new rpc method invoker.
// serverID can be empty, will be used for directives.
func NewInvoker(b bus.Bus, serverID string) *Invoker {
	return &Invoker{
		b:        b,
		serverID: serverID,
	}
}

// InvokeMethod invokes the method matching the service & method ID.
// Returns false, nil if not found.
// If service string is empty, ignore it.
func (i *Invoker) InvokeMethod(serviceID, methodID string, strm srpc.Stream) (bool, error) {
	ctx := strm.Context()

	invokers, invokerRef, err := ExLookupRpcService(ctx, i.b, serviceID, i.serverID)
	if err != nil || invokerRef == nil {
		return false, err
	}
	defer invokerRef.Release()

	var sl srpc.InvokerSlice = invokers
	return sl.InvokeMethod(serviceID, methodID, strm)
}

// _ is a type assertion
var _ srpc.Invoker = ((*Invoker)(nil))
