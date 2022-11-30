package bifrost_rpc_access

import (
	"context"
	"errors"
	"sync"

	bifrost_rpc "github.com/aperturerobotics/bifrost/rpc"
	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/aperturerobotics/starpc/rpcstream"
	srpc "github.com/aperturerobotics/starpc/srpc"
	"github.com/aperturerobotics/util/broadcast"
)

// AccessRpcServiceServer is the server for AccessRpcService.
type AccessRpcServiceServer struct {
	b bus.Bus
}

// NewAccessRpcServiceServer builds a AccessRpcService server with a bus.
func NewAccessRpcServiceServer(b bus.Bus) *AccessRpcServiceServer {
	return &AccessRpcServiceServer{b: b}
}

// LookupRpcService looks up the rpc service via the bus.
func (s *AccessRpcServiceServer) LookupRpcService(
	req *LookupRpcServiceRequest,
	strm SRPCAccessRpcService_LookupRpcServiceStream,
) error {
	dir := req.ToDirective()

	var mtx sync.Mutex
	var bcast broadcast.Broadcast
	vals := make(map[uint32]struct{})
	var sendQueue []*LookupRpcServiceResponse
	var disposed bool
	var resErr error
	di, ref, err := s.b.AddDirective(dir, bus.NewCallbackHandler(
		func(av directive.AttachedValue) {
			mtx.Lock()
			defer mtx.Unlock()
			_, ok := av.GetValue().(bifrost_rpc.LookupRpcServiceValue)
			if !ok {
				return
			}
			vals[av.GetValueID()] = struct{}{}
			if len(vals) == 1 {
				sendQueue = append(sendQueue, &LookupRpcServiceResponse{
					Exists: true,
				})
				bcast.Broadcast()
			}
		}, func(av directive.AttachedValue) {
			mtx.Lock()
			defer mtx.Unlock()
			_, exists := vals[av.GetValueID()]
			if !exists {
				return
			}
			delete(vals, av.GetValueID())
			if len(vals) == 0 {
				sendQueue = append(sendQueue, &LookupRpcServiceResponse{
					Removed: true,
				})
				bcast.Broadcast()
			}
		}, func() {
			mtx.Lock()
			disposed = true
			bcast.Broadcast()
			mtx.Unlock()
		},
	))
	if err != nil {
		return err
	}
	defer ref.Release()

	defer di.AddIdleCallback(func(resErrs []error) {
		mtx.Lock()
		if resErr == nil {
			for _, err := range resErrs {
				if err != nil {
					resErr = err
					break
				}
			}
		}
		sendQueue = append(sendQueue, &LookupRpcServiceResponse{
			Idle: true,
		})
		bcast.Broadcast()
		mtx.Unlock()
	})()

	for {
		select {
		case <-strm.Context().Done():
			return context.Canceled
		case <-bcast.GetWaitCh():
		}

		mtx.Lock()
		currSendQueue, currIsDisposed := sendQueue, disposed
		sendQueue = nil
		mtx.Unlock()
		for _, msg := range currSendQueue {
			if err := strm.Send(msg); err != nil {
				return err
			}
		}
		if currIsDisposed {
			return errors.New("directive disposed")
		}
	}
}

// CallRpcService looks up the rpc service with the request & invokes the RPC.
func (s *AccessRpcServiceServer) CallRpcService(strm SRPCAccessRpcService_CallRpcServiceStream) error {
	return rpcstream.HandleRpcStream(strm, func(ctx context.Context, componentID string) (srpc.Invoker, func(), error) {
		// parse component id json
		req := &LookupRpcServiceRequest{}
		if err := req.UnmarshalComponentID(componentID); err != nil {
			return nil, nil, err
		}
		if err := req.Validate(); err != nil {
			return nil, nil, err
		}
		// lookup the rpc service invokers
		invokers, invokerRef, err := bifrost_rpc.ExLookupRpcService(ctx, s.b, req.GetServiceId(), req.GetServerId())
		if err != nil || invokerRef == nil {
			return nil, nil, err
		}
		if len(invokers) == 0 {
			invokerRef.Release()
			return nil, nil, nil
		}
		// return the invoker slice
		return srpc.InvokerSlice(invokers), invokerRef.Release, nil
	})
}

// _ is a type assertion
var _ SRPCAccessRpcServiceServer = ((*AccessRpcServiceServer)(nil))
