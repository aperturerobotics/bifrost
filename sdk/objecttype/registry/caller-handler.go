package s4wave_objecttype_registry

import (
	"context"

	resource_server "github.com/s4wave/spacewave/bldr/resource/server"
	"github.com/s4wave/spacewave/net/peer"
	sdk_world "github.com/s4wave/spacewave/sdk/world"
	"github.com/s4wave/spacewave/sdk/world/objecttype"
)

// CallerHandler restores the authenticated native caller from the engine lent
// by the registry. Every Resource it publishes retains that identity.
type CallerHandler struct {
	// handler serves typed requests under the restored caller context.
	handler SRPCObjectTypeHandlerServiceServer
}

// NewCallerHandler binds a typed handler to the registry's caller identity.
func NewCallerHandler(handler SRPCObjectTypeHandlerServiceServer) *CallerHandler {
	return &CallerHandler{handler: handler}
}

// InvokeObjectType restores the granting mount's caller before invoking the
// handler. A transport-authenticated remote peer takes precedence over a lent
// engine, so a remote request cannot replace its identity with an attachment.
func (h *CallerHandler) InvokeObjectType(ctx context.Context, req *InvokeObjectTypeRequest) (*InvokeObjectTypeResponse, error) {
	// Read the caller from the trusted engine attachment for native dispatch.
	caller := objecttype.CallerFromContext(ctx)
	if caller.PeerID == "" && req.GetAttachedEngineResourceId() != 0 {
		resources, err := resource_server.MustGetResourceClientContext(ctx)
		if err != nil {
			return nil, err
		}
		client, err := resources.GetAttachedResource(req.GetAttachedEngineResourceId())
		if err != nil {
			return nil, err
		}
		info, err := sdk_world.NewSRPCEngineResourceServiceClient(client).GetEngineInfo(ctx, &sdk_world.GetEngineInfoRequest{})
		if err != nil {
			return nil, err
		}
		if info.GetSessionPeerId() != "" {
			_, err = peer.IDB58Decode(info.GetSessionPeerId())
			if err != nil {
				return nil, err
			}
			caller.Participant = info.GetCallerParticipant()
			caller.PeerID = info.GetSessionPeerId()
		}
	}

	// Carry the identity into the handler and all independently retained children.
	ctx = bindCallerResources(ctx, caller)
	return h.handler.InvokeObjectType(ctx, req)
}

// _ is a type assertion.
var _ SRPCObjectTypeHandlerServiceServer = (*CallerHandler)(nil)
