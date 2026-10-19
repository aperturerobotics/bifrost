//go:build !js

package s4wave_flowgraph_test

import (
	"context"
	"testing"
	"time"

	"github.com/aperturerobotics/starpc/srpc"
	resource "github.com/s4wave/spacewave/bldr/resource"
	resource_client "github.com/s4wave/spacewave/bldr/resource/client"
	resource_server "github.com/s4wave/spacewave/bldr/resource/server"
	resource_objecttype_registry "github.com/s4wave/spacewave/core/resource/objecttype/registry"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/bucket"
	"github.com/s4wave/spacewave/db/world"
	world_block_engine "github.com/s4wave/spacewave/db/world/block/engine"
	world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	world_types "github.com/s4wave/spacewave/db/world/types"
	s4wave_device "github.com/s4wave/spacewave/sdk/device"
	flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
	s4wave_objecttype_registry "github.com/s4wave/spacewave/sdk/objecttype/registry"
	"github.com/sirupsen/logrus"
)

// boundType is the object type the binding test registers with two ports.
const boundType = "test/ported"

// registerPortedType registers boundType with an input and an output port for
// the life of the test.
func registerPortedType(t *testing.T, ctx context.Context, tb *world_testbed.Testbed) {
	t.Helper()

	// Serve the registry through an in-process resource client.
	registry := resource_objecttype_registry.NewObjectTypeRegistryResource(nil)
	mux := srpc.NewMux()
	if err := resource_server.NewResourceServer(registry.GetMux()).Register(mux); err != nil {
		t.Fatal(err)
	}
	resources, err := resource_client.NewClient(ctx, resource.NewSRPCResourceServiceClient(srpc.NewClient(srpc.NewServerPipe(srpc.NewServer(mux)))))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(resources.Release)
	rootRef := resources.AccessRootResource()
	t.Cleanup(rootRef.Release)
	rootClient, err := rootRef.GetClient()
	if err != nil {
		t.Fatal(err)
	}

	// Resolve the registry's lookups on the World's bus.
	le := logrus.NewEntry(logrus.New())
	releaseBridge, err := tb.Bus.AddController(ctx, resource_objecttype_registry.NewBridgeController(le, tb.Bus, registry), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(releaseBridge)

	// Register the type.
	ports := []*s4wave_objecttype_registry.ObjectTypePort{
		{Name: "in", Direction: s4wave_objecttype_registry.ObjectTypePortDirection_OBJECT_TYPE_PORT_DIRECTION_INPUT, TypeId: "test/stream"},
		{Name: "out", Direction: s4wave_objecttype_registry.ObjectTypePortDirection_OBJECT_TYPE_PORT_DIRECTION_OUTPUT, TypeId: "test/event"},
	}
	registration, err := s4wave_objecttype_registry.NewSRPCObjectTypeRegistryResourceServiceClient(rootClient).RegisterObjectType(ctx, &s4wave_objecttype_registry.RegisterObjectTypeRequest{
		TypeId:   boundType,
		PluginId: "test-plugin",
		Metadata: &s4wave_objecttype_registry.ObjectTypeMetadata{DisplayName: "Ported", Ports: ports},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(resources.CreateResourceReference(registration.GetResourceId()).Release)
}

// createTypedObject creates an object of a type in the World.
func createTypedObject(t *testing.T, ctx context.Context, ws world.WorldState, key, typeID string) {
	t.Helper()
	obj, _, err := world.CreateWorldObject(ctx, ws, key, func(cursor *block.Cursor) error {
		cursor.SetBlock(&s4wave_device.Device{PeerId: key, Label: key}, true)
		return nil
	})
	world.ReleaseObjectState(obj)
	if err != nil {
		t.Fatal(err)
	}
	if err := world_types.SetObjectType(ctx, ws, key, typeID); err != nil {
		t.Fatal(err)
	}
}

// retype changes an object's type in one committed transaction, so a watcher
// never reads the object without a type.
func retype(t *testing.T, ctx context.Context, engine world.Engine, key, typeID string) {
	t.Helper()
	tx, err := engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()
	if err := world_types.SetObjectType(ctx, tx, key, typeID); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

// bindingWatch reads watched snapshots.
type bindingWatch struct {
	t      *testing.T
	stream flowgraph.SRPCFlowgraphResourceService_WatchFlowgraphClient
}

// next reads the next snapshot and returns the binding of the node.
func (w *bindingWatch) next(nodeID string) *flowgraph.FlowgraphBinding {
	w.t.Helper()
	msg, err := w.stream.Recv()
	if err != nil {
		w.t.Fatalf("watch: %v", err)
	}
	return msg.GetSnapshot().GetBindings()[nodeID]
}

// requireBinding checks a binding's object, type and state.
func requireBinding(t *testing.T, binding *flowgraph.FlowgraphBinding, objectKey, typeID string, state flowgraph.FlowgraphBindingState) {
	t.Helper()
	if binding.GetObjectKey() != objectKey || binding.GetTypeId() != typeID || binding.GetState() != state {
		t.Fatalf("binding = %v, want %s %q %s", binding, objectKey, typeID, state)
	}
}

// requirePorts checks that a binding declares the ports of boundType.
func requirePorts(t *testing.T, binding *flowgraph.FlowgraphBinding) {
	t.Helper()
	ports := binding.GetPorts()
	if len(ports) != 2 ||
		ports[0].GetName() != "in" || ports[0].GetDirection() != flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_INPUT || ports[0].GetTypeId() != "test/stream" ||
		ports[1].GetName() != "out" || ports[1].GetDirection() != flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_OUTPUT || ports[1].GetTypeId() != "test/event" {
		t.Fatalf("binding ports = %v", ports)
	}
}

// TestFlowgraphBindsObject binds nodes to World objects and reads each node's
// derived state through the snapshot and the watch.
func TestFlowgraphBindsObject(t *testing.T) {
	// Start a World, register a type with two ports and serve a graph.
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	tb := world_testbed.MustDefault(t, ctx)
	registerPortedType(t, ctx, tb)
	ws := tb.WorldState
	createTypedObject(t, ctx, ws, "thing/one", boundType)
	createTypedObject(t, ctx, ws, "thing/plain", "test/unregistered")
	if _, _, err := ws.ApplyWorldOp(ctx, &flowgraph.CreateFlowgraphOp{ObjectKey: "flowgraph/first", Name: "first"}, ""); err != nil {
		t.Fatal(err)
	}
	serve := func(ws world.WorldState, engine world.Engine) flowgraph.SRPCFlowgraphResourceServiceClient {
		mux := srpc.NewMux()
		resource := flowgraph.NewFlowgraphResource(ws, engine, tb.Bus, "", "flowgraph/first")
		if err := flowgraph.SRPCRegisterFlowgraphResourceService(mux, resource); err != nil {
			t.Fatal(err)
		}
		return flowgraph.NewSRPCFlowgraphResourceServiceClient(srpc.NewClient(srpc.NewServerPipe(srpc.NewServer(mux))))
	}
	service := serve(ws, tb.Engine)

	// Bind a node to each object and to one that does not exist.
	updated, err := service.UpdateFlowgraph(ctx, &flowgraph.UpdateFlowgraphRequest{SetNodes: map[string]*flowgraph.FlowgraphNode{
		"one":     {TypeId: "object", ObjectKey: "thing/one"},
		"plain":   {TypeId: "object", ObjectKey: "thing/plain"},
		"missing": {TypeId: "object", ObjectKey: "thing/missing"},
		"free":    {TypeId: "object"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	requireBinding(t, updated.GetSnapshot().GetBindings()["one"], "thing/one", boundType, flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_BOUND)
	requirePorts(t, updated.GetSnapshot().GetBindings()["one"])

	// Read the same bindings back, with each unbound or unavailable state.
	read, err := service.GetFlowgraph(ctx, &flowgraph.GetFlowgraphRequest{})
	if err != nil {
		t.Fatal(err)
	}
	bindings := read.GetSnapshot().GetBindings()
	requireBinding(t, bindings["one"], "thing/one", boundType, flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_BOUND)
	requirePorts(t, bindings["one"])
	requireBinding(t, bindings["plain"], "thing/plain", "test/unregistered", flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_UNREGISTERED)
	requireBinding(t, bindings["missing"], "thing/missing", "", flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_UNAVAILABLE)
	if len(bindings["plain"].GetPorts()) != 0 || len(bindings["missing"].GetPorts()) != 0 {
		t.Fatal("an unregistered or absent object declared ports")
	}
	if _, ok := bindings["free"]; ok || len(bindings) != 3 {
		t.Fatalf("bindings = %v, want only the three bound nodes", bindings)
	}

	// Watch the graph: the initial snapshot is bound; deleting the object makes
	// the node unavailable and recreating it makes the node bound again.
	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()
	stream, err := service.WatchFlowgraph(watchCtx, &flowgraph.WatchFlowgraphRequest{})
	if err != nil {
		t.Fatal(err)
	}
	watch := &bindingWatch{t: t, stream: stream}
	requireBinding(t, watch.next("one"), "thing/one", boundType, flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_BOUND)
	if deleted, err := ws.DeleteObject(ctx, "thing/one"); err != nil || !deleted {
		t.Fatalf("delete object = %v, %v", deleted, err)
	}
	requireBinding(t, watch.next("one"), "thing/one", "", flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_UNAVAILABLE)
	tx, err := tb.Engine.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()
	createTypedObject(t, ctx, tx, "thing/one", boundType)
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	requireBinding(t, watch.next("one"), "thing/one", boundType, flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_BOUND)

	// A retyped object reads as unregistered, and its original type restores it.
	retype(t, ctx, tb.Engine, "thing/one", "test/other")
	requireBinding(t, watch.next("one"), "thing/one", "test/other", flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_UNREGISTERED)
	retype(t, ctx, tb.Engine, "thing/one", boundType)
	requireBinding(t, watch.next("one"), "thing/one", boundType, flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_BOUND)
	watchCancel()

	// A World reopened from the same store reads the same binding.
	reopenedID := tb.EngineID + "-reopened"
	transformConf, err := world_testbed.NewEngineTransformConfig(tb.EngineBucketID)
	if err != nil {
		t.Fatal(err)
	}
	initRef := &bucket.ObjectRef{BucketId: tb.EngineBucketID, TransformConf: transformConf}
	conf := world_block_engine.NewConfig(reopenedID, tb.EngineVolumeID, tb.EngineBucketID, tb.EngineObjectStoreID, initRef, nil, false)
	engine, engineRef, err := world.StartEngine(ctx, tb.Bus, conf, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer engineRef.Release()
	reopened, err := serve(world.NewEngineWorldState(engine, true), engine).GetFlowgraph(ctx, &flowgraph.GetFlowgraphRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got := reopened.GetSnapshot().GetState().GetNodes()["one"].GetObjectKey(); got != "thing/one" {
		t.Fatalf("reopened node object key = %q", got)
	}
	requireBinding(t, reopened.GetSnapshot().GetBindings()["one"], "thing/one", boundType, flowgraph.FlowgraphBindingState_FLOWGRAPH_BINDING_STATE_BOUND)
	requirePorts(t, reopened.GetSnapshot().GetBindings()["one"])
}
