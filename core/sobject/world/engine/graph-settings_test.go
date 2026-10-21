package sobject_world_engine

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/aperturerobotics/cayley/graph/kv"
	"github.com/aperturerobotics/cayley/quad"
	"github.com/s4wave/spacewave/db/world"
	world_block "github.com/s4wave/spacewave/db/world/block"
	world_block_tx "github.com/s4wave/spacewave/db/world/block/tx"
	"github.com/sirupsen/logrus"
)

// TestGraphSettingsStepOpReplays writes a graph, moves it to the predicate
// index and back in small steps on two replays of the same operations, and
// checks both reach the same roots, the settings, and the same lookups.
func TestGraphSettingsStepOpReplays(t *testing.T) {
	// Write the same graph on two replays.
	ctx := context.Background()
	pid := newProcessTestPeerID(t)
	c, so, base := newProcessTestWorld(t, ctx)
	w1, w2 := newReplayWorld(c, so), newReplayWorld(c, so)

	// Create ten objects, each linked to every earlier one.
	var ops []*SOWorldOp
	for i := range 10 {
		ops = append(ops, transformTestCreateOp(t, base, fmt.Sprintf("o/%d", i)))
	}
	for i := range 10 {
		for j := range i {
			ops = append(ops, graphSettingsTestQuadOp(t, fmt.Sprintf("o/%d", i), fmt.Sprintf("<rel-%d>", (i+j)%3), fmt.Sprintf("o/%d", j)))
		}
	}

	// Replay the operations on both.
	h1, h2 := base, base
	var nonce uint64
	for _, op := range ops {
		nonce++
		h1 = processTransformTestOp(t, c, w1, pid, nonce, h1, op)
		h2 = processTransformTestOp(t, c, w2, pid, nonce, h2, op)
	}
	want := graphSettingsTestLookups(t, ctx, w1, h1)

	// Move to the predicate index and back, five links per step.
	target := graphSettingsTestPredicate(t)
	for _, phase := range []struct {
		target *world_block.GraphSettings
		want   *world_block.GraphSettings
	}{{target, target}, {nil, nil}} {
		op := &SOWorldOp{Body: &SOWorldOp_GraphSettingsStep{GraphSettingsStep: &GraphSettingsStepOp{Target: phase.target, Max: 5}}}
		steps := 0
		for {
			nonce++
			n1 := processTransformTestOp(t, c, w1, pid, nonce, h1, op)
			n2 := processTransformTestOp(t, c, w2, pid, nonce, h2, op)
			if !n1.GetHeadRef().EqualVT(n2.GetHeadRef()) {
				t.Fatalf("step %d: replays reached different heads", steps)
			}
			unchanged := n1.GetHeadRef().EqualVT(h1.GetHeadRef())
			h1, h2 = n1, n2
			if unchanged {
				break
			}
			steps++
			if steps > 100 {
				t.Fatal("graph settings change did not finish")
			}
		}
		t.Logf("changed in %d steps", steps)
		if steps < 2 {
			t.Fatalf("change took %d steps, want several", steps)
		}

		// The root holds the target settings and serves the same lookups.
		got := graphSettingsTestRead(t, ctx, w1, h1)
		if !got.EqualVT(phase.want) {
			t.Fatalf("graph settings = %v, want %v", got, phase.want)
		}
		if lookups := graphSettingsTestLookups(t, ctx, w1, h1); !slices.Equal(lookups, want) {
			t.Fatalf("lookups = %v, want %v", lookups, want)
		}
	}
	if err := w1.close(ctx); err != nil {
		t.Fatal(err.Error())
	}
	if err := w2.close(ctx); err != nil {
		t.Fatal(err.Error())
	}
}

// TestGraphSettingsStepOpRejects checks the steps replay rejects.
func TestGraphSettingsStepOpRejects(t *testing.T) {
	// Replay each operation on an initialized World, or an empty state.
	ctx := context.Background()
	pid := newProcessTestPeerID(t)
	c, so, headState := newProcessTestWorld(t, ctx)
	w := newReplayWorld(c, so)
	defer w.close(ctx)

	// Build a valid target and one with an index still building.
	target := graphSettingsTestPredicate(t)
	building := slices.Clone(kv.DefaultQuadIndexes)
	building[0].State = kv.QuadIndexState_BUILDING
	partial, err := world_block.NewGraphSettings(building)
	if err != nil {
		t.Fatal(err.Error())
	}
	for _, tc := range []struct {
		name  string
		state *InnerState
		step  *GraphSettingsStepOp
	}{
		{"uninitialized", &InnerState{}, &GraphSettingsStepOp{Target: target, Max: 5}},
		{"zero max", headState, &GraphSettingsStepOp{Target: target}},
		{"max over bound", headState, &GraphSettingsStepOp{Target: target, Max: maxGraphSettingsStep + 1}},
		{"incomplete target", headState, &GraphSettingsStepOp{Target: partial, Max: 5}},
	} {
		opData, err := (&SOWorldOp{Body: &SOWorldOp_GraphSettingsStep{GraphSettingsStep: tc.step}}).MarshalVT()
		if err != nil {
			t.Fatal(err.Error())
		}
		next, res, err := c.processOp(ctx, logrus.NewEntry(logrus.New()), w, opData, "test-op", pid, 1, 0, tc.state)
		if err != nil || next != nil || res.GetSuccess() {
			t.Fatalf("%s: process = %v, %v, %v; want rejection", tc.name, next, res, err)
		}
	}
}

// graphSettingsTestPredicate returns the default indexes plus the predicate
// index.
func graphSettingsTestPredicate(t *testing.T) *world_block.GraphSettings {
	t.Helper()
	settings, err := world_block.NewGraphSettings(append(slices.Clone(kv.DefaultQuadIndexes), kv.QuadIndex{
		Dirs: []quad.Direction{quad.Predicate, quad.Label, quad.Subject},
	}))
	if err != nil {
		t.Fatal(err.Error())
	}
	return settings
}

// graphSettingsTestQuadOp returns an operation linking subject to object.
func graphSettingsTestQuadOp(t *testing.T, subject, predicate, object string) *SOWorldOp {
	t.Helper()
	tx, err := world_block_tx.NewTxSetGraphQuad(world.GraphQuadToQuad(world.NewGraphQuadWithKeys(subject, predicate, object, "")))
	if err != nil {
		t.Fatal(err.Error())
	}
	return &SOWorldOp{Body: &SOWorldOp_ApplyTxOp{ApplyTxOp: &ApplyTxOp{Tx: tx}}}
}

// graphSettingsTestEngine returns the replay's engine on the World at
// headState.
func graphSettingsTestEngine(t *testing.T, ctx context.Context, w *replayWorld, headState *InnerState) *world_block.Engine {
	t.Helper()
	ws, err := w.bind(ctx, headState.GetHeadRef())
	if err != nil {
		t.Fatal(err.Error())
	}
	return ws.bengine
}

// graphSettingsTestRead returns the graph settings of the World at headState.
func graphSettingsTestRead(t *testing.T, ctx context.Context, w *replayWorld, headState *InnerState) *world_block.GraphSettings {
	// Read the root through a snapshot of the replay's engine.
	t.Helper()
	btx, err := graphSettingsTestEngine(t, ctx, w, headState).NewBlockEngineTransaction(ctx, false)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer btx.Discard()
	settings, err := btx.GetGraphSettings(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}
	return settings
}

// graphSettingsTestLookups looks up each predicate of the World at headState.
func graphSettingsTestLookups(t *testing.T, ctx context.Context, w *replayWorld, headState *InnerState) []string {
	// Look up each predicate in one read transaction.
	t.Helper()
	tx, err := graphSettingsTestEngine(t, ctx, w, headState).NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer tx.Discard()
	var out []string
	for p := range 3 {
		quads, err := tx.LookupGraphQuads(ctx, world.NewGraphQuad("", fmt.Sprintf("<rel-%d>", p), "", ""), 0)
		if err != nil {
			t.Fatal(err.Error())
		}
		for _, q := range quads {
			out = append(out, q.GetSubject()+" "+q.GetPredicate()+" "+q.GetObj())
		}
	}
	slices.Sort(out)
	return out
}
