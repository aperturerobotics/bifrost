package flowgraph_run_test

import (
	"context"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/world"
	forge_job "github.com/s4wave/spacewave/forge/job"
	forge_task "github.com/s4wave/spacewave/forge/task"
	s4wave_flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
	flowgraph_run "github.com/s4wave/spacewave/sdk/flowgraph/run"
)

// standingKey is the key of the test Flowgraph's standing run.
var standingKey = flowgraph_run.StandingRunKey(graphKey)

// standingRun finds or creates the test Flowgraph's standing run in one World
// transaction and returns its key.
func (e *env) standingRun() string {
	var key string
	e.exec(func(ctx context.Context, tx world.WorldState) error {
		var err error
		key, err = flowgraph_run.StandingRun(ctx, tx, e.sender, graphKey, clusterKey, nil)
		return err
	})
	return key
}

// readStanding reads the standing run, or nil before it exists.
func (e *env) readStanding() *s4wave_flowgraph.FlowgraphRun {
	obj, found, err := e.tb.WorldState.GetObject(e.ctx, standingKey)
	defer world.ReleaseObjectState(obj)
	if err != nil {
		e.t.Fatal(err)
	}
	if !found {
		return nil
	}
	var run *s4wave_flowgraph.FlowgraphRun
	_, _, err = world.AccessObjectState(e.ctx, obj, false, func(cursor *block.Cursor) error {
		var err error
		run, err = s4wave_flowgraph.UnmarshalFlowgraphRun(e.ctx, cursor)
		return err
	})
	if err != nil {
		e.t.Fatal(err)
	}
	return run
}

// deliverStanding delivers events from the source node to the standing run.
func (e *env) deliverStanding(cursor, next string, events ...*flowgraph_run.DeliveredEvent) flowgraph_run.DeliverEventsResult {
	var result flowgraph_run.DeliverEventsResult
	e.exec(func(ctx context.Context, tx world.WorldState) error {
		var err error
		result, err = flowgraph_run.DeliverEvents(ctx, tx, standingKey, &flowgraph_run.DeliverEventsRequest{
			NodeId:     sourceID,
			Cursor:     []byte(cursor),
			NextCursor: []byte(next),
			Events:     events,
		})
		return err
	})
	return result
}

// waitStanding waits until both Steps have run the given number of times and
// the standing run holds no activation, and returns the run.
func (e *env) waitStanding(visits uint32) *s4wave_flowgraph.FlowgraphRun {
	e.t.Helper()
	var run *s4wave_flowgraph.FlowgraphRun
	e.waitFor("standing run visits", func() bool {
		run = e.readStanding()
		return run.GetVisits()["a"].GetCount() == visits && run.GetVisits()["b"].GetCount() == visits && len(run.GetActivations()) == 0
	})
	return run
}

// TestStandingRun delivers events across two Steps of a standing run and
// checks that the run keeps no done activation, never completes, and leaves
// the Job's Tasks in place.
func TestStandingRun(t *testing.T) {
	// Create a graph of an event source feeding a Step that feeds a second Step.
	e := newEnv(t)
	source := &s4wave_flowgraph.FlowgraphNode{
		TypeId: "test/event-source",
		Ports:  []*s4wave_flowgraph.FlowgraphPort{port("events", s4wave_flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_OUTPUT)},
	}
	e.createGraph(map[string]*s4wave_flowgraph.FlowgraphNode{
		sourceID: source,
		"a":      e.step([]string{"in"}, [2]string{"out", "none"}, true, nil),
		"b":      e.step([]string{"in"}, [2]string{"done", "none"}, true, nil),
	}, map[string]*s4wave_flowgraph.FlowgraphConnection{
		"src-a": connect(sourceID, "events", "a", "in"),
		"a-b":   connect("a", "out", "b", "in"),
	})
	e.exec(func(ctx context.Context, tx world.WorldState) error {
		obj, _, err := world.CreateWorldObject(ctx, tx, eventObjectKey, func(cursor *block.Cursor) error {
			cursor.SetBlock(&s4wave_flowgraph.FlowgraphVisits{}, true)
			return nil
		})
		world.ReleaseObjectState(obj)
		return err
	})

	// A first find creates the run, and a second finds the same one.
	if _, found, err := flowgraph_run.FindStandingRun(e.ctx, e.tb.WorldState, graphKey); err != nil || found {
		t.Fatalf("find before create = %v, %v; want not found", found, err)
	}
	key := e.standingRun()
	if key != standingKey {
		t.Fatalf("standing run key = %q, want %q", key, standingKey)
	}
	created := e.readStanding()
	if !created.GetStanding() || created.GetState() != s4wave_flowgraph.FlowgraphRunState_FLOWGRAPH_RUN_STATE_RUNNING {
		t.Fatalf("created run = %v, want a running standing run", created)
	}
	if again := e.standingRun(); again != key {
		t.Fatalf("second find = %q, want %q", again, key)
	}
	if !e.readStanding().EqualVT(created) {
		t.Fatal("second find changed the run")
	}
	e.exec(func(ctx context.Context, tx world.WorldState) error {
		if err := flowgraph_run.StartRun(ctx, tx, e.sender, "other-run", graphKey, clusterKey, nil, flowgraph_run.Standing()); err == nil {
			t.Error("standing StartRun under another key succeeded")
		}
		return nil
	})

	// Run the standing run and deliver an event, which crosses both Steps.
	e.startWorker()
	_, ref, err := flowgraph_run.StartControllerWithConfig(e.ctx, e.tb.Bus, flowgraph_run.NewConfig(e.tb.EngineID, key))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(ref.Release)
	if result := e.deliverStanding("", "c1", event("e1", "")); result != flowgraph_run.DeliverEventsResult_DELIVER_EVENTS_RESULT_DELIVERED {
		t.Fatalf("first delivery = %s", result)
	}
	run := e.waitStanding(1)
	if run.GetState() != s4wave_flowgraph.FlowgraphRunState_FLOWGRAPH_RUN_STATE_RUNNING {
		t.Fatalf("standing run state = %s, want running with nothing waiting", run.GetState())
	}

	// A second event crosses both Steps again, and the run still holds no activation.
	if result := e.deliverStanding("c1", "c2", event("e2", "")); result != flowgraph_run.DeliverEventsResult_DELIVER_EVENTS_RESULT_DELIVERED {
		t.Fatalf("second delivery = %s", result)
	}
	run = e.waitStanding(2)
	if run.GetState() != s4wave_flowgraph.FlowgraphRunState_FLOWGRAPH_RUN_STATE_RUNNING || len(run.GetArrivals()) != 0 {
		t.Fatalf("standing run = %v, want running with no arrivals", run)
	}
	if got := string(run.GetCursor()[sourceID]); got != "c2" {
		t.Fatalf("cursor = %q, want c2", got)
	}

	// The Job keeps all four Tasks, each complete and under its own key.
	taskKeys, err := forge_job.ListJobTasks(e.ctx, e.tb.WorldState, flowgraph_run.JobKey(key))
	if err != nil {
		t.Fatal(err)
	}
	if len(taskKeys) != 4 {
		t.Fatalf("job tasks = %v, want 4", taskKeys)
	}
	for _, taskKey := range taskKeys {
		task, err := forge_task.LookupTaskBody(e.ctx, e.tb.WorldState, taskKey)
		if err != nil {
			t.Fatal(err)
		}
		if !task.IsComplete() {
			t.Fatalf("task %s is not complete", taskKey)
		}
	}
}
