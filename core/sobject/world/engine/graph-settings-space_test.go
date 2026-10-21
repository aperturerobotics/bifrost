package sobject_world_engine_test

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/aperturerobotics/cayley/graph/kv"
	"github.com/aperturerobotics/cayley/quad"
	sobject_world_engine "github.com/s4wave/spacewave/core/sobject/world/engine"
	bucket_lookup "github.com/s4wave/spacewave/db/bucket/lookup"
	"github.com/s4wave/spacewave/db/world"
	world_block "github.com/s4wave/spacewave/db/world/block"
)

// TestStepGraphSettingsChangesSpaceWorld moves a Space World to the predicate
// index through its engine while committing links between steps, checks the
// lookups see every link, then returns it to the default settings.
func TestStepGraphSettingsChangesSpaceWorld(t *testing.T) {
	// Write a graph on a Space World.
	ctx := t.Context()
	sw := newSpaceWorld(ctx, t)
	commitSpaceGraphLinks(ctx, t, sw.eng, 0, 8)
	target, err := world_block.NewGraphSettings(append(slices.Clone(kv.DefaultQuadIndexes), kv.QuadIndex{
		Dirs: []quad.Direction{quad.Predicate, quad.Label, quad.Subject},
	}))
	if err != nil {
		t.Fatal(err)
	}

	// Build the index four links per step, committing a link after each step.
	err = world.AccessEngine(ctx, sw.eng, func(ctx context.Context, eng world.Engine) error {
		ge := eng.(sobject_world_engine.GraphSettingsEngine)
		n := 8
		for steps := 0; ; steps++ {
			done, err := ge.StepGraphSettings(ctx, target, 4)
			if err != nil || done {
				t.Logf("built in %d steps", steps)
				return err
			}
			commitSpaceGraphLinks(ctx, t, sw.eng, n, n+1)
			n++
			if steps > 100 {
				t.Fatal("graph settings change did not finish")
			}
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := spaceGraphSettings(ctx, t, sw.eng); !got.EqualVT(target) {
		t.Fatalf("graph settings after the build = %v, want %v", got, target)
	}
	built := lookupSpaceGraph(ctx, t, sw.eng)
	if want := spaceGraphLinks(t, sw.eng); !slices.Equal(built, want) || len(built) == 0 {
		t.Fatalf("lookups after the build = %v, want %v", built, want)
	}

	// Return to the default settings in one call.
	err = world.AccessEngine(ctx, sw.eng, func(ctx context.Context, eng world.Engine) error {
		_, err := world_block.ApplyGraphSettings(ctx, eng.(sobject_world_engine.GraphSettingsEngine), nil, world_block.DefaultGraphSettingsStepMax)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := spaceGraphSettings(ctx, t, sw.eng); got != nil {
		t.Fatalf("graph settings after the drop = %v, want default", got)
	}
	if got := lookupSpaceGraph(ctx, t, sw.eng); !slices.Equal(got, built) {
		t.Fatalf("lookups after the drop = %v, want %v", got, built)
	}
}

// spaceGraphSettings returns the graph settings of the World root.
func spaceGraphSettings(ctx context.Context, t *testing.T, eng world.Engine) *world_block.GraphSettings {
	// Read the root through the engine's World state access.
	t.Helper()
	var settings *world_block.GraphSettings
	err := eng.AccessWorldState(ctx, nil, func(bls *bucket_lookup.Cursor) error {
		// Unmarshal the root block and keep its settings.
		_, bcs := bls.BuildTransaction(nil)
		wi, err := bcs.Unmarshal(ctx, world_block.NewWorldBlock)
		if err != nil {
			return err
		}
		settings = wi.(*world_block.World).GetGraphSettings()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return settings
}

// commitSpaceGraphLinks creates objects lo to hi, each linked to every
// earlier object, in one commit.
func commitSpaceGraphLinks(ctx context.Context, t *testing.T, eng world.Engine, lo, hi int) {
	// Open a write transaction.
	t.Helper()
	tx, err := eng.NewTransaction(ctx, true)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()

	// Create the objects and links and commit.
	for i := lo; i < hi; i++ {
		obj, err := tx.CreateObject(ctx, fmt.Sprintf("o/%d", i), nil)
		if err != nil {
			t.Fatal(err)
		}
		world.ReleaseObjectState(obj)
		for j := range i {
			q := world.NewGraphQuadWithKeys(fmt.Sprintf("o/%d", i), fmt.Sprintf("<rel-%d>", (i+j)%3), fmt.Sprintf("o/%d", j), "")
			if err := tx.SetGraphQuad(ctx, q); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}

// spaceGraphLinks lists every link of the World by subject.
func spaceGraphLinks(t *testing.T, eng world.Engine) []string {
	t.Helper()
	return lookupSpaceGraphBy(t.Context(), t, eng, func(i int) world.GraphQuad {
		return world.NewGraphQuadWithKeys(fmt.Sprintf("o/%d", i), "", "", "")
	}, 64)
}

// lookupSpaceGraph lists every link of the World by predicate.
func lookupSpaceGraph(ctx context.Context, t *testing.T, eng world.Engine) []string {
	t.Helper()
	return lookupSpaceGraphBy(ctx, t, eng, func(p int) world.GraphQuad {
		return world.NewGraphQuad("", fmt.Sprintf("<rel-%d>", p), "", "")
	}, 3)
}

// lookupSpaceGraphBy runs the lookups pattern returns for 0 to n in one read
// transaction and returns the sorted links.
func lookupSpaceGraphBy(ctx context.Context, t *testing.T, eng world.Engine, pattern func(int) world.GraphQuad, n int) []string {
	// Run every lookup in one read transaction.
	t.Helper()
	tx, err := eng.NewTransaction(ctx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Discard()
	var out []string
	for i := range n {
		quads, err := tx.LookupGraphQuads(ctx, pattern(i), 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, q := range quads {
			out = append(out, q.GetSubject()+" "+q.GetPredicate()+" "+q.GetObj())
		}
	}
	slices.Sort(out)
	return out
}
