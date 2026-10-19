package flowgraph_run_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	forge_target "github.com/s4wave/spacewave/forge/target"
	forge_task "github.com/s4wave/spacewave/forge/task"
	s4wave_flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
	flowgraph_run "github.com/s4wave/spacewave/sdk/flowgraph/run"
)

const (
	// sourceID is the event-source node of the delivery tests.
	sourceID = "src"
	// eventObjectKey is the World object the test events happened on.
	eventObjectKey = "events/room"
)

// eventGraph creates a Flowgraph with an event-source node feeding one Step.
func (e *env) eventGraph() {
	source := &s4wave_flowgraph.FlowgraphNode{
		TypeId: "test/event-source",
		Ports:  []*s4wave_flowgraph.FlowgraphPort{port("events", s4wave_flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_OUTPUT)},
	}
	e.createGraph(map[string]*s4wave_flowgraph.FlowgraphNode{
		sourceID: source,
		"a":      e.step([]string{"in"}, [2]string{"out", "none"}, true, nil),
	}, map[string]*s4wave_flowgraph.FlowgraphConnection{
		"src-a": connect(sourceID, "events", "a", "in"),
	})
}

// deliver delivers events from the source node in one World transaction.
func (e *env) deliver(cursor, next string, events ...*flowgraph_run.DeliveredEvent) flowgraph_run.DeliverEventsResult {
	var result flowgraph_run.DeliverEventsResult
	e.exec(func(ctx context.Context, tx world.WorldState) error {
		var err error
		result, err = flowgraph_run.DeliverEvents(ctx, tx, runKey, &flowgraph_run.DeliverEventsRequest{
			NodeId:     sourceID,
			Cursor:     []byte(cursor),
			NextCursor: []byte(next),
			Events:     events,
		})
		return err
	})
	return result
}

// event builds a delivered event on the test object.
func event(id, key string) *flowgraph_run.DeliveredEvent {
	return &flowgraph_run.DeliveredEvent{ObjectKey: eventObjectKey, EventId: id, Key: key}
}

// eventIDs returns the event ID of each event arrival, in order.
func eventIDs(run *s4wave_flowgraph.FlowgraphRun) []string {
	var ids []string
	for _, arrival := range run.GetArrivals() {
		ids = append(ids, arrival.GetEvent().GetEventId())
	}
	return ids
}

// requireEventIDs fails unless the run's arrivals are the given events.
func requireEventIDs(t *testing.T, run *s4wave_flowgraph.FlowgraphRun, want ...string) {
	t.Helper()
	got := eventIDs(run)
	if len(got) != len(want) {
		t.Fatalf("arrivals = %v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("arrivals = %v, want %v", got, want)
		}
	}
}

// TestDeliverEventsCursor applies a delivery once and rejects a delivery that
// read a cursor another reader has already moved.
func TestDeliverEventsCursor(t *testing.T) {
	// Start a run whose source node feeds a Step, with no controller running.
	e := newEnv(t)
	e.eventGraph()
	e.startRun()

	// Deliver two events from the initial cursor.
	if result := e.deliver("", "c2", event("e1", "k1"), event("e2", "")); result != flowgraph_run.DeliverEventsResult_DELIVER_EVENTS_RESULT_DELIVERED {
		t.Fatalf("first delivery = %s", result)
	}
	delivered := e.readRun()
	requireEventIDs(t, delivered, "e1", "e2")
	if got := string(delivered.GetCursor()[sourceID]); got != "c2" {
		t.Fatalf("cursor = %q, want c2", got)
	}
	first := delivered.GetArrivals()[0]
	if first.GetNodeId() != "a" || first.GetPort() != "in" || first.GetKey() != "k1" || first.GetEvent().GetObjectKey() != eventObjectKey {
		t.Fatalf("first arrival = %v", first)
	}

	// A second reader that read the initial cursor is stale and changes nothing.
	if result := e.deliver("", "c3", event("e1", "k1"), event("e2", ""), event("e3", "")); result != flowgraph_run.DeliverEventsResult_DELIVER_EVENTS_RESULT_STALE_CURSOR {
		t.Fatalf("stale delivery = %s", result)
	}
	stale := e.readRun()
	if !stale.EqualVT(delivered) {
		t.Fatalf("stale delivery changed the run: %v, was %v", stale, delivered)
	}
	requireEventIDs(t, stale, "e1", "e2")

	// Resuming from the stored cursor delivers the next event once.
	if result := e.deliver("c2", "c3", event("e3", "")); result != flowgraph_run.DeliverEventsResult_DELIVER_EVENTS_RESULT_DELIVERED {
		t.Fatalf("resumed delivery = %s", result)
	}
	resumed := e.readRun()
	requireEventIDs(t, resumed, "e1", "e2", "e3")
	if !bytes.Equal(resumed.GetCursor()[sourceID], []byte("c3")) {
		t.Fatalf("cursor = %q, want c3", resumed.GetCursor()[sourceID])
	}
}

// TestEventArrivalInput starts a Step on an event arrival and reads the
// Task's input as a WORLD_OBJECT input of the event's object.
func TestEventArrivalInput(t *testing.T) {
	// Create the event's object and a run holding one delivered event.
	e := newEnv(t)
	e.eventGraph()
	e.exec(func(ctx context.Context, tx world.WorldState) error {
		obj, _, err := world.CreateWorldObject(ctx, tx, eventObjectKey, func(cursor *block.Cursor) error {
			cursor.SetBlock(&s4wave_flowgraph.FlowgraphVisits{}, true)
			return nil
		})
		world.ReleaseObjectState(obj)
		return err
	})
	e.startRun()
	e.deliver("", "c1", event("e1", ""))

	// The Step runs once on the event.
	e.startWorker()
	e.startController()
	run := e.waitRun(s4wave_flowgraph.FlowgraphRunState_FLOWGRAPH_RUN_STATE_COMPLETE)
	requireTrace(t, run, "a.out")
	inputs := run.GetActivations()[0].GetInputs()
	if len(inputs) != 1 || inputs[0].GetEvent().GetEventId() != "e1" {
		t.Fatalf("activation inputs = %v, want event e1", inputs)
	}

	// The Task reads the event's object through its input port.
	target, _, err := forge_task.LookupTaskTarget(e.ctx, e.tb.WorldState, run.GetActivations()[0].GetTaskKey())
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range target.GetInputs() {
		if in.GetName() != "in" {
			continue
		}
		if in.GetInputType() != forge_target.InputType_InputType_WORLD_OBJECT || in.GetWorldObject().GetObjectKey() != eventObjectKey {
			t.Fatalf("input in = %v, want WORLD_OBJECT of %s", in, eventObjectKey)
		}
		return
	}
	t.Fatal("activation has no input in")
}
