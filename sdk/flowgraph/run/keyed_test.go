package flowgraph_run_test

import (
	"bytes"
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/aperturerobotics/controllerbus/config"
	"github.com/aperturerobotics/controllerbus/controller"
	"github.com/aperturerobotics/controllerbus/directive"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/block"
	"github.com/s4wave/spacewave/db/block/blob"
	"github.com/s4wave/spacewave/db/world"
	forge_target "github.com/s4wave/spacewave/forge/target"
	target_json "github.com/s4wave/spacewave/forge/target/json"
	forge_value "github.com/s4wave/spacewave/forge/value"
	s4wave_flowgraph "github.com/s4wave/spacewave/sdk/flowgraph"
	flowgraph_run "github.com/s4wave/spacewave/sdk/flowgraph/run"
)

const (
	// liveStepID is the controller ID of the test Step that reads live inputs.
	liveStepID = "test/live-step"
	// keyedID is the keyed Step of the live tests.
	keyedID = "a"
	// otherSourceID is the second event-source node of the join test.
	otherSourceID = "src2"
)

// liveConfig is the test Step's config: the World engine and the run to read.
type liveConfig struct {
	*flowgraph_run.Config
}

// GetConfigID returns the test Step's config ID.
func (c *liveConfig) GetConfigID() string { return liveStepID }

// EqualsConfig checks if the other config is equal.
func (c *liveConfig) EqualsConfig(other config.Config) bool {
	oc, ok := other.(*liveConfig)
	return ok && c.EqualVT(oc.Config)
}

// liveSteps is the test Step factory. Each Step execution takes one token
// before it completes, so a test holds activations running. A Step with ack
// set reads its live inputs the way a code Step does: it echoes the event ID
// of each and acknowledges it. A Step without ack reads none.
type liveSteps struct {
	e *env
	// ack makes each execution echo and acknowledge its live inputs.
	ack bool
	// tokens releases one execution per token.
	tokens chan struct{}
	// echoed receives the event ID of each live input an execution read.
	echoed chan string
	// runKey is the run the executions belong to.
	runKey string
}

// GetConfigID returns the config ID.
func (f *liveSteps) GetConfigID() string { return liveStepID }

// GetControllerID returns the controller ID.
func (f *liveSteps) GetControllerID() string { return liveStepID }

// ConstructConfig constructs an empty config.
func (f *liveSteps) ConstructConfig() config.Config {
	return &liveConfig{Config: &flowgraph_run.Config{}}
}

// Construct constructs a Step execution.
func (f *liveSteps) Construct(context.Context, config.Config, controller.ConstructOpts) (controller.Controller, error) {
	return &liveStep{f: f}, nil
}

// GetVersion returns the version.
func (f *liveSteps) GetVersion() controller.Version { return flowgraph_run.Version }

// liveStep is one execution of the test Step.
type liveStep struct {
	f      *liveSteps
	handle forge_target.ExecControllerHandle
}

// GetControllerInfo returns information about the controller.
func (s *liveStep) GetControllerInfo() *controller.Info {
	return controller.NewInfo(liveStepID, flowgraph_run.Version, "test Step that reads live inputs")
}

// InitForgeExecController keeps the handle.
func (s *liveStep) InitForgeExecController(_ context.Context, _ forge_target.InputMap, handle forge_target.ExecControllerHandle) error {
	s.handle = handle
	return nil
}

// activation returns the Task key of the run's one running activation of the
// keyed Step: the identity a production Step receives with its start.
func (s *liveStep) activation() (string, error) {
	run := s.f.e.readStanding()
	for _, activation := range run.GetActivations() {
		if !activation.GetDone() && activation.GetNodeId() == keyedID {
			return activation.GetTaskKey(), nil
		}
	}
	return "", errors.New("keyed Step has no running activation")
}

// Execute reads live inputs while it waits for a token, then sets its output.
func (s *liveStep) Execute(ctx context.Context) error {
	// Echo and acknowledge each live input until the token arrives.
	e := s.f.e
	watchCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	watched := make(chan error, 1)
	if s.f.ack {
		id, err := s.activation()
		if err != nil {
			return err
		}
		go func() {
			watched <- flowgraph_run.WatchLiveInputs(watchCtx, e.tb.WorldState, s.f.runKey, id, func(inputs []*s4wave_flowgraph.FlowgraphLiveInput) error {
				for _, input := range inputs {
					select {
					case s.f.echoed <- input.GetArrival().GetEvent().GetEventId():
					case <-watchCtx.Done():
						return watchCtx.Err()
					}
				}
				last := inputs[len(inputs)-1].GetSeq()
				return world.ExecTransaction(watchCtx, e.tb.Engine, true, func(ctx context.Context, tx world.WorldState) error {
					return flowgraph_run.AckLiveInputs(ctx, tx, s.f.runKey, id, last)
				})
			})
		}()
	}
	select {
	case <-s.f.tokens:
	case <-ctx.Done():
		return ctx.Err()
	}
	cancel()
	if s.f.ack {
		if err := <-watched; !errors.Is(err, context.Canceled) {
			return err
		}
	}

	// Choose the first output port.
	seed := forge_value.NewValueWithBlockRef("", nil)
	val, err := forge_target.AccessValue(ctx, s.handle, seed, func(bcs *block.Cursor) error {
		_, err := blob.BuildBlob(ctx, 2, bytes.NewReader([]byte("ok")), bcs, nil)
		return err
	})
	if err != nil {
		return err
	}
	val.Name = "out"
	return s.handle.SetOutputs(ctx, forge_value.ValueSlice{val}, true)
}

// HandleDirective resolves no directives.
func (s *liveStep) HandleDirective(context.Context, directive.Instance) ([]directive.Resolver, error) {
	return nil, nil
}

// Close releases nothing.
func (s *liveStep) Close() error { return nil }

// _ is a type assertion
var _ forge_target.ExecController = (*liveStep)(nil)

// liveGraph creates a Flowgraph of the event sources feeding one keyed Step
// that runs the test controller, with the standing run, its Worker and its
// controller started, and returns the test Step factory. The Step has one
// input port per source.
func (e *env) liveGraph(ack bool, sources ...string) *liveSteps {
	// Register the test Step.
	steps := &liveSteps{e: e, ack: ack, tokens: make(chan struct{}, 8), echoed: make(chan string, 8), runKey: standingKey}
	e.tb.StaticResolver.AddFactory(steps)
	target, err := target_json.ResolveYAML(e.ctx, e.tb.Bus, []byte(`
outputs:
  - name: out
    outputType: OutputType_EXEC
    execOutput: out
  - name: none
    outputType: OutputType_EXEC
    execOutput: none
exec:
  controller:
    id: `+liveStepID+`
    config:
      engineId: `+strconv.Quote(e.tb.EngineID)+`
      objectKey: `+strconv.Quote(standingKey)+`
`))
	if err != nil {
		e.t.Fatal(err)
	}

	// Connect each source to its own input port of the keyed Step.
	step := &s4wave_flowgraph.FlowgraphNode{
		TypeId: s4wave_flowgraph.StepNodeTypeID,
		Step:   &s4wave_flowgraph.FlowgraphStep{Target: target, Keyed: true},
		Ports: []*s4wave_flowgraph.FlowgraphPort{
			port("out", s4wave_flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_OUTPUT),
			port("none", s4wave_flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_OUTPUT),
		},
	}
	nodes := map[string]*s4wave_flowgraph.FlowgraphNode{keyedID: step}
	connections := make(map[string]*s4wave_flowgraph.FlowgraphConnection)
	for _, source := range sources {
		nodes[source] = &s4wave_flowgraph.FlowgraphNode{
			TypeId: "test/event-source",
			Ports:  []*s4wave_flowgraph.FlowgraphPort{port("events", s4wave_flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_OUTPUT)},
		}
		step.Ports = append(step.Ports, port("in-"+source, s4wave_flowgraph.FlowgraphPortDirection_FLOWGRAPH_PORT_DIRECTION_INPUT))
		connections[source+"-"+keyedID] = connect(source, "events", keyedID, "in-"+source)
	}
	e.createGraph(nodes, connections)

	// Create the event object, the standing run, and start both controllers.
	e.exec(func(ctx context.Context, tx world.WorldState) error {
		obj, _, err := world.CreateWorldObject(ctx, tx, eventObjectKey, func(cursor *block.Cursor) error {
			cursor.SetBlock(&s4wave_flowgraph.FlowgraphVisits{}, true)
			return nil
		})
		world.ReleaseObjectState(obj)
		return err
	})
	e.standingRun()
	e.startWorker()
	_, ref, err := flowgraph_run.StartControllerWithConfig(e.ctx, e.tb.Bus, flowgraph_run.NewConfig(e.tb.EngineID, standingKey))
	if err != nil {
		e.t.Fatal(err)
	}
	e.t.Cleanup(ref.Release)
	return steps
}

// deliverFrom delivers events from a source node to the standing run.
func (e *env) deliverFrom(source, cursor, next string, events ...*flowgraph_run.DeliveredEvent) {
	e.t.Helper()
	e.exec(func(ctx context.Context, tx world.WorldState) error {
		result, err := flowgraph_run.DeliverEvents(ctx, tx, standingKey, &flowgraph_run.DeliverEventsRequest{
			NodeId:     source,
			Cursor:     []byte(cursor),
			NextCursor: []byte(next),
			Events:     events,
		})
		if err == nil && result != flowgraph_run.DeliverEventsResult_DELIVER_EVENTS_RESULT_DELIVERED {
			e.t.Fatalf("delivery from %s = %s", source, result)
		}
		return err
	})
}

// waitStandingRun waits until the standing run satisfies done and returns it.
func (e *env) waitStandingRun(what string, done func(*s4wave_flowgraph.FlowgraphRun) bool) *s4wave_flowgraph.FlowgraphRun {
	e.t.Helper()
	var run *s4wave_flowgraph.FlowgraphRun
	e.waitFor(what, func() bool {
		run = e.readStanding()
		return done(run)
	})
	return run
}

// release lets n Step executions complete.
func (f *liveSteps) release(n int) {
	for range n {
		f.tokens <- struct{}{}
	}
}

// nextEcho waits for the next event ID a Step read live.
func (f *liveSteps) nextEcho() string {
	f.e.t.Helper()
	select {
	case id := <-f.echoed:
		return id
	case <-time.After(30 * time.Second):
		f.e.t.Fatal("timed out waiting for an echoed live input")
		return ""
	}
}

// eventID returns an arrival's event ID.
func eventID(arrival *s4wave_flowgraph.FlowgraphArrival) string {
	return arrival.GetEvent().GetEventId()
}

// requireDrained waits for the standing run to hold nothing and have started
// the Step the given number of times.
func (e *env) requireDrained(visits uint32) *s4wave_flowgraph.FlowgraphRun {
	e.t.Helper()
	run := e.waitStandingRun("drained run", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return run.GetVisits()[keyedID].GetCount() == visits && len(run.GetActivations()) == 0
	})
	if len(run.GetArrivals()) != 0 {
		e.t.Fatalf("drained run holds arrivals %v", eventIDs(run))
	}
	return run
}

// TestKeyedLiveDelivery sends a second arrival for a key while its activation
// runs, and reads it through the activation's inbox.
func TestKeyedLiveDelivery(t *testing.T) {
	// Start a keyed Step whose executions echo and acknowledge live inputs.
	e := newEnv(t)
	steps := e.liveGraph(true, sourceID)

	// The first arrival for key k starts an activation that records the key.
	e.deliverFrom(sourceID, "", "c1", event("e1", "k"))
	run := e.waitStandingRun("first activation", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return len(run.GetActivations()) == 1
	})
	first := run.GetActivations()[0]
	if first.GetKey() != "k" || len(first.GetInputs()) != 1 || eventID(first.GetInputs()[0]) != "e1" {
		t.Fatalf("first activation = %v, want key k with input e1", first)
	}

	// A second arrival for k reaches the running activation, which echoes it.
	e.deliverFrom(sourceID, "c1", "c2", event("e2", "k"))
	if got := steps.nextEcho(); got != "e2" {
		t.Fatalf("echoed %q, want e2", got)
	}

	// The acknowledgement empties the inbox, and no second activation started.
	run = e.waitStandingRun("acknowledged inbox", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return len(run.GetActivations()) == 1 && len(run.GetActivations()[0].GetInbox()) == 0 && len(run.GetArrivals()) == 0
	})
	if got := run.GetActivations()[0].GetLastInputSeq(); got != 1 {
		t.Fatalf("last input seq = %d, want 1", got)
	}
	if got := run.GetVisits()[keyedID].GetCount(); got != 1 {
		t.Fatalf("visits = %d, want 1", got)
	}

	// The activation ends without a successor, since it acknowledged e2.
	steps.release(1)
	e.requireDrained(1)
}

// TestKeyedActivationsRunTogether runs one activation per key at the same
// time and routes a repeat arrival to the activation of its key.
func TestKeyedActivationsRunTogether(t *testing.T) {
	// Start a keyed Step whose executions wait for a token.
	e := newEnv(t)
	steps := e.liveGraph(false, sourceID)

	// Arrivals for two keys run two activations at once.
	e.deliverFrom(sourceID, "", "c1", event("e1", "k1"), event("e2", "k2"))
	run := e.waitStandingRun("two activations", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return len(run.GetActivations()) == 2
	})
	keys := []string{run.GetActivations()[0].GetKey(), run.GetActivations()[1].GetKey()}
	if keys[0] != "k1" || keys[1] != "k2" || run.GetActivations()[0].GetDone() || run.GetActivations()[1].GetDone() {
		t.Fatalf("activations = %v, want k1 and k2 both running", run.GetActivations())
	}

	// A repeat arrival for k1 joins k1's inbox and starts no third activation.
	e.deliverFrom(sourceID, "c1", "c2", event("e3", "k1"))
	run = e.waitStandingRun("k1 inbox", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return len(run.GetActivations()) == 2 && len(run.GetActivations()[0].GetInbox()) == 1
	})
	input := run.GetActivations()[0].GetInbox()[0]
	if input.GetSeq() != 1 || eventID(input.GetArrival()) != "e3" || input.GetArrival().GetKey() != "k1" || len(run.GetActivations()[1].GetInbox()) != 0 {
		t.Fatalf("activations = %v, want e3 in k1's inbox only", run.GetActivations())
	}
	if len(run.GetArrivals()) != 0 || run.GetVisits()[keyedID].GetCount() != 2 {
		t.Fatalf("run = %v, want no arrivals and two visits", run)
	}

	// Both executions end; k1's unacknowledged e3 starts a third activation.
	steps.release(3)
	e.requireDrained(3)
}

// TestKeyedSuccessorReceivesUnacknowledgedInput starts a successor with an
// input its predecessor never acknowledged.
func TestKeyedSuccessorReceivesUnacknowledgedInput(t *testing.T) {
	// Start a keyed Step whose executions read no live inputs.
	e := newEnv(t)
	steps := e.liveGraph(false, sourceID)

	// Deliver a second arrival for k while the first activation runs.
	e.deliverFrom(sourceID, "", "c1", event("e1", "k"))
	e.waitStandingRun("first activation", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return len(run.GetActivations()) == 1
	})
	e.deliverFrom(sourceID, "c1", "c2", event("e2", "k"))
	e.waitStandingRun("inbox", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return len(run.GetActivations()) == 1 && len(run.GetActivations()[0].GetInbox()) == 1
	})

	// The first activation ends, and its successor starts with e2 as its input.
	steps.release(1)
	run := e.waitStandingRun("successor", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return run.GetVisits()[keyedID].GetCount() == 2 && len(run.GetActivations()) == 1
	})
	successor := run.GetActivations()[0]
	if successor.GetKey() != "k" || successor.GetDone() || len(successor.GetInbox()) != 0 ||
		len(successor.GetInputs()) != 1 || eventID(successor.GetInputs()[0]) != "e2" || successor.GetInputs()[0].GetKey() != "k" {
		t.Fatalf("successor = %v, want a running activation for k with input e2", successor)
	}
	if len(run.GetArrivals()) != 0 {
		t.Fatalf("run holds arrivals %v, want none", eventIDs(run))
	}

	// The successor ends without another one.
	steps.release(1)
	e.requireDrained(2)
}

// TestKeyedJoinPerKey joins the inputs of a keyed Step by key: arrivals at its
// two ports that carry different keys do not start it.
func TestKeyedJoinPerKey(t *testing.T) {
	// Start a keyed Step with a port for each of two sources.
	e := newEnv(t)
	steps := e.liveGraph(false, sourceID, otherSourceID)

	// Arrivals for k1 at one port and k2 at the other wait for a partner.
	e.deliverFrom(sourceID, "", "c1", event("e1", "k1"))
	e.deliverFrom(otherSourceID, "", "d1", event("e2", "k2"))
	e.waitStandingRun("both arrivals", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return len(run.GetArrivals()) == 2
	})
	if run := e.readStanding(); len(run.GetActivations()) != 0 {
		t.Fatalf("activations = %v, want none before a key has both ports", run.GetActivations())
	}

	// The partner for k1 starts the Step with both k1 arrivals, and k2 waits on.
	e.deliverFrom(otherSourceID, "d1", "d2", event("e3", "k1"))
	run := e.waitStandingRun("k1 activation", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return len(run.GetActivations()) == 1
	})
	activation := run.GetActivations()[0]
	if activation.GetKey() != "k1" || len(activation.GetInputs()) != 2 ||
		eventID(activation.GetInputs()[0]) != "e1" || eventID(activation.GetInputs()[1]) != "e3" {
		t.Fatalf("activation = %v, want key k1 with inputs e1 and e3", activation)
	}
	requireEventIDs(t, run, "e2")
	steps.release(1)
	e.waitStandingRun("drained k1", func(run *s4wave_flowgraph.FlowgraphRun) bool {
		return run.GetVisits()[keyedID].GetCount() == 1 && len(run.GetActivations()) == 0
	})
}
