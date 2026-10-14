package wasivm_plugin

import (
	"context"

	"github.com/aperturerobotics/controllerbus/bus"
	"github.com/aperturerobotics/starpc/srpc"
	"github.com/pkg/errors"
	resource_client "github.com/s4wave/spacewave/bldr/resource/client"
	resource_server "github.com/s4wave/spacewave/bldr/resource/server"
	"github.com/s4wave/spacewave/db/world"
	"github.com/s4wave/spacewave/net/peer"
	s4wave_objecttype_registry "github.com/s4wave/spacewave/sdk/objecttype/registry"
	s4wave_process "github.com/s4wave/spacewave/sdk/process"
	s4wave_vm "github.com/s4wave/spacewave/sdk/vm"
	s4wave_world "github.com/s4wave/spacewave/sdk/world"
	sdk_world_engine "github.com/s4wave/spacewave/sdk/world/engine"
	s4wave_worldop_registry "github.com/s4wave/spacewave/sdk/worldop/registry"
	"github.com/sirupsen/logrus"
)

// handler serves the WasiVM ObjectType and operations to the core plugin.
type handler struct {
	le *logrus.Entry
	// b mounts the CDN Spaces of published images.
	b bus.Bus
	// consoles holds the console of each running VM.
	consoles *consoleTable
}

// newHandler constructs a handler.
func newHandler(le *logrus.Entry, b bus.Bus) *handler {
	return &handler{le: le, b: b, consoles: newConsoleTable()}
}

// InvokeObjectType opens a WasiVM resource on the attached engine.
func (h *handler) InvokeObjectType(
	ctx context.Context,
	req *s4wave_objecttype_registry.InvokeObjectTypeRequest,
) (*s4wave_objecttype_registry.InvokeObjectTypeResponse, error) {
	// Require a WasiVM and the engine it lives in.
	if req.GetTypeId() != s4wave_vm.WasiVmTypeID {
		return nil, errors.Errorf("wasivm plugin: unhandled object type %s", req.GetTypeId())
	}
	if req.GetAttachedEngineResourceId() == 0 {
		return nil, errors.New("wasivm plugin: attached engine resource is required")
	}
	resourceCtx, err := resource_server.MustGetResourceClientContext(ctx)
	if err != nil {
		return nil, err
	}

	// Serve execution, the console and the display of the VM on the attached
	// engine.
	engine, err := sdk_world_engine.NewAttachedEngine(ctx, req.GetAttachedEngineResourceId())
	if err != nil {
		return nil, err
	}
	vm := &vmResource{
		le:        h.le.WithField("object-key", req.GetObjectKey()),
		b:         h.b,
		objectKey: req.GetObjectKey(),
		ws:        world.NewEngineWorldState(engine, true),
		consoles:  h.consoles,
	}
	mux := resource_server.NewResourceMux(func(mux srpc.Mux) error {
		if err := s4wave_process.SRPCRegisterPersistentExecutionService(mux, vm); err != nil {
			return err
		}
		if err := s4wave_vm.SRPCRegisterWasiVmConsoleService(mux, &vmConsole{consoles: h.consoles, objectKey: vm.objectKey}); err != nil {
			return err
		}
		return s4wave_vm.SRPCRegisterWasiVmDisplayService(mux, &vmDisplay{consoles: h.consoles, objectKey: vm.objectKey})
	})
	resourceID, err := resourceCtx.AddResource(mux, engine.Release)
	if err != nil {
		engine.Release()
		return nil, err
	}
	return &s4wave_objecttype_registry.InvokeObjectTypeResponse{ResourceId: resourceID}, nil
}

// ApplyWorldOp applies a WasiVM operation to the attached world state.
func (h *handler) ApplyWorldOp(
	ctx context.Context,
	req *s4wave_worldop_registry.ApplyWorldOpRequest,
) (*s4wave_worldop_registry.ApplyWorldOpResponse, error) {
	// Decode the operation.
	if req.GetAttachedWorldStateResourceId() == 0 {
		return nil, errors.New("wasivm plugin: attached world state resource is required")
	}
	op, err := unmarshalOp(ctx, req.GetOperationTypeId(), req.GetOpData())
	if err != nil {
		return nil, err
	}

	// Open the attached world state.
	resourceCtx, err := resource_server.MustGetResourceClientContext(ctx)
	if err != nil {
		return nil, err
	}
	client := sdk_world_engine.NewAttachedResourceClient(resourceCtx)
	ref := client.CreateResourceReference(req.GetAttachedWorldStateResourceId())
	ws, err := sdk_world_engine.NewSDKWorldState(client, ref, false)
	if err != nil {
		ref.Release()
		return nil, err
	}
	defer ws.Release()

	// Apply the operation to it.
	sysErr, err := op.ApplyWorldOp(ctx, h.le, ws, peer.ID(""))
	if err != nil {
		return nil, err
	}
	return &s4wave_worldop_registry.ApplyWorldOpResponse{SystemError: sysErr}, nil
}

// ApplyWorldObjectOp applies a WasiVM operation to the attached object.
func (h *handler) ApplyWorldObjectOp(
	ctx context.Context,
	req *s4wave_worldop_registry.ApplyWorldObjectOpRequest,
) (*s4wave_worldop_registry.ApplyWorldObjectOpResponse, error) {
	// Decode the operation.
	if req.GetAttachedObjectStateResourceId() == 0 {
		return nil, errors.New("wasivm plugin: attached object state resource is required")
	}
	op, err := unmarshalOp(ctx, req.GetOperationTypeId(), req.GetOpData())
	if err != nil {
		return nil, err
	}

	// Reference the attached object.
	resourceCtx, err := resource_server.MustGetResourceClientContext(ctx)
	if err != nil {
		return nil, err
	}
	client := sdk_world_engine.NewAttachedResourceClient(resourceCtx)
	ref := client.CreateResourceReference(req.GetAttachedObjectStateResourceId())

	// Open it, checking it is the object the request names.
	objectKey, err := attachedObjectKey(ctx, ref)
	if err == nil && objectKey != req.GetObjectKey() {
		err = errors.Errorf("wasivm plugin: attached object %q is not %q", objectKey, req.GetObjectKey())
	}
	if err != nil {
		ref.Release()
		return nil, err
	}
	os, err := sdk_world_engine.NewSDKObjectState(client, ref, objectKey, false)
	if err != nil {
		ref.Release()
		return nil, err
	}
	defer os.Release()

	// Apply the operation to it.
	sysErr, err := op.ApplyWorldObjectOp(ctx, h.le, os, peer.ID(""))
	if err != nil {
		return nil, err
	}
	return &s4wave_worldop_registry.ApplyWorldObjectOpResponse{SystemError: sysErr}, nil
}

// ValidateOp validates a WasiVM operation.
func (h *handler) ValidateOp(
	ctx context.Context,
	req *s4wave_worldop_registry.ValidateOpRequest,
) (*s4wave_worldop_registry.ValidateOpResponse, error) {
	op, err := unmarshalOp(ctx, req.GetOperationTypeId(), req.GetOpData())
	if err == nil {
		err = op.Validate()
	}
	if err != nil {
		return &s4wave_worldop_registry.ValidateOpResponse{Error: err.Error()}, nil
	}
	return &s4wave_worldop_registry.ValidateOpResponse{}, nil
}

// attachedObjectKey reads the key of the attached object ref.
func attachedObjectKey(ctx context.Context, ref resource_client.ResourceRef) (string, error) {
	// Ask the object state resource for its key.
	client, err := ref.GetClient()
	if err != nil {
		return "", err
	}
	resp, err := s4wave_world.NewSRPCObjectStateResourceServiceClient(client).GetKey(ctx, &s4wave_world.GetKeyRequest{})
	if err != nil {
		return "", err
	}
	return resp.GetObjectKey(), nil
}

// unmarshalOp decodes a WasiVM operation.
func unmarshalOp(ctx context.Context, opTypeID string, data []byte) (world.Operation, error) {
	// Construct the operation of the type, then decode its data.
	op, err := world.LookupOpSlice([]world.LookupOp{
		s4wave_vm.LookupCreateWasiVmOp,
		s4wave_vm.LookupSetWasiVmStateOp,
		s4wave_vm.LookupCreateWasiVmImageOp,
		s4wave_vm.LookupCreateWasiVmPackageOp,
	}).LookupOp(ctx, opTypeID)
	if err == nil && op == nil {
		err = errors.New("unhandled operation type")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "wasivm plugin: operation %s", opTypeID)
	}
	if err := op.UnmarshalBlock(data); err != nil {
		return nil, err
	}
	return op, nil
}

// _ is a type assertion.
var (
	_ s4wave_objecttype_registry.SRPCObjectTypeHandlerServiceServer = (*handler)(nil)
	_ s4wave_worldop_registry.SRPCWorldOpHandlerServiceServer       = (*handler)(nil)
)
