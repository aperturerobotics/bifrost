package bldr_manifest_world

import (
	"context"
	"io"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/aperturerobotics/cayley"
	"github.com/s4wave/spacewave/db/world"
	world_testbed "github.com/s4wave/spacewave/db/world/testbed"
	world_types "github.com/s4wave/spacewave/db/world/types"
)

// TestManifestListingWaitChange checks that manifest listings wait past unrelated
// graph writes and rerun for changes to the manifest and type edges they read.
func TestManifestListingWaitChange(t *testing.T) {
	// Exercise each listing with additions and removals of relevant edges.
	tests := []struct {
		// name describes the listing and changed edge.
		name string
		// list is the production listing interface under test.
		list func(context.Context, world.WorldState, ...string) ([]string, error)
		// edge is the relevant quad to add or remove.
		edge world.GraphQuad
		// remove reports whether the relevant edge starts present.
		remove bool
	}{
		{"manifest/add-link", ListManifests, NewManifestQuad("listing/bundle", "listing/next", "app"), false},
		{"manifest/remove-link", ListManifests, NewManifestQuad("listing/bundle", "listing/next", "app"), true},
		{"manifest/add-type", ListManifests, world.NewGraphQuadWithKeys("listing/candidate", world_types.TypePred.String(), world_types.BuildTypeObjectKey(ManifestTypeID), ""), false},
		{"manifest/remove-type", ListManifests, world.NewGraphQuadWithKeys("listing/candidate", world_types.TypePred.String(), world_types.BuildTypeObjectKey(ManifestRefTypeID), ""), true},
		{"candidate/add-link", ListManifestCandidates, NewManifestQuad("listing/bundle", "listing/next", "app"), false},
		{"candidate/remove-link", ListManifestCandidates, NewManifestQuad("listing/bundle", "listing/next", "app"), true},
		{"empty/add-link", ListManifestCandidates, NewManifestQuad("listing/empty", "listing/next", "app"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a real World with an intermediate bundle and an untyped candidate.
			ctx := t.Context()
			tb := world_testbed.MustDefault(t, ctx)
			ws := tb.WorldState
			for _, key := range []string{"listing/store", "listing/bundle", "listing/candidate", "listing/next", "listing/unrelated", "listing/empty", world_types.BuildTypeObjectKey(ManifestTypeID), world_types.BuildTypeObjectKey(ManifestRefTypeID)} {
				obj, err := ws.CreateObject(ctx, key, nil)
				world.ReleaseObjectState(obj)
				if err != nil {
					t.Fatal(err)
				}
			}
			for _, edge := range []world.GraphQuad{
				NewManifestQuad("listing/store", "listing/bundle", ""),
				NewManifestQuad("listing/bundle", "listing/candidate", "app"),
				world.NewGraphQuadWithKeys("listing/next", world_types.TypePred.String(), world_types.BuildTypeObjectKey(ManifestTypeID), ""),
			} {
				if err := ws.SetGraphQuad(ctx, edge); err != nil {
					t.Fatal(err)
				}
			}
			if tt.remove {
				if err := ws.SetGraphQuad(ctx, tt.edge); err != nil {
					t.Fatal(err)
				}
			}

			// Record the listing's reads before any subsequent graph writes.
			startKey := "listing/store"
			if tt.name == "empty/add-link" {
				startKey = "listing/empty"
			}
			seqno, err := ws.GetSeqno(ctx)
			if err != nil {
				t.Fatal(err)
			}
			reads := world.NewReadSet()
			before, err := tt.list(ctx, reads.Track(ws), startKey)
			if err != nil {
				t.Fatal(err)
			}

			// Change unrelated manifest and type edges, and another predicate on a read node.
			for _, edge := range []world.GraphQuad{
				NewManifestQuad("listing/unrelated", "listing/next", "app"),
				world.NewGraphQuadWithKeys("listing/unrelated", world_types.TypePred.String(), world_types.BuildTypeObjectKey(ManifestTypeID), ""),
				world.NewGraphQuadWithKeys("listing/bundle", "<other>", "listing/next", ""),
			} {
				if err := ws.SetGraphQuad(ctx, edge); err != nil {
					t.Fatal(err)
				}
			}

			// Read the real change feed before a relevant write can join this report.
			waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			err = world.WatchChanges(waitCtx, ws, seqno, func(_ uint64, changes *world.ChangeSet) error {
				if reads.Touched(changes) {
					t.Error("unrelated World changes would rerun the listing")
				}
				return io.EOF
			})
			if err != io.EOF {
				t.Fatalf("read unrelated changes: %v", err)
			}

			// Change a relevant edge and record exactly which commit should wake the view.
			if tt.remove {
				err = ws.DeleteGraphQuad(ctx, tt.edge)
			} else {
				err = ws.SetGraphQuad(ctx, tt.edge)
			}
			if err != nil {
				t.Fatal(err)
			}
			want, err := ws.GetSeqno(ctx)
			if err != nil {
				t.Fatal(err)
			}

			// Use the same read-set wait as production watches, with a bounded deadline.
			got, err := reads.WaitChange(waitCtx, ws, seqno)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Fatalf("listing woke at seqno %d, want relevant edge commit %d", got, want)
			}

			// Rerun the listing and check that the relevant edge changed its result.
			after, err := tt.list(ctx, ws, startKey)
			if err != nil {
				t.Fatal(err)
			}
			if slices.Equal(before, after) {
				t.Fatalf("listing stayed %v after relevant edge change", after)
			}
		})
	}
}

// TestManifestListingTraversal checks that tracked listings retain the Cayley
// path's depth limit, cycle handling, type filtering and sorted, distinct output.
func TestManifestListingTraversal(t *testing.T) {
	// Seed a wide graph with a cycle, converging edges and a path beyond fifty hops.
	ctx := t.Context()
	tb := world_testbed.MustDefault(t, ctx)
	err := world.ExecTransaction(ctx, tb.Engine, true, func(ctx context.Context, ws world.WorldState) error {
		for i := range 140 {
			key := "traversal/" + strconv.Itoa(i)
			obj, err := ws.CreateObject(ctx, key, nil)
			world.ReleaseObjectState(obj)
			if err != nil {
				return err
			}
			if err := world_types.SetObjectType(ctx, ws, key, ManifestTypeID); err != nil {
				return err
			}
		}
		for i := 1; i < 140; i++ {
			src := 0
			if i < 55 {
				src = i - 1
			}
			if err := ws.SetGraphQuad(ctx, NewManifestQuad("traversal/"+strconv.Itoa(src), "traversal/"+strconv.Itoa(i), "app")); err != nil {
				return err
			}
		}
		for _, edge := range []world.GraphQuad{
			NewManifestQuad("traversal/1", "traversal/0", ""),
			NewManifestQuad("traversal/55", "traversal/2", ""),
			NewManifestQuad("traversal/0", "traversal/55", "other"),
		} {
			if err := ws.SetGraphQuad(ctx, edge); err != nil {
				return err
			}
		}
		return world_types.SetObjectType(ctx, ws, "traversal/3", ManifestBundleTypeID)
	})
	if err != nil {
		t.Fatal(err)
	}

	// Compare both production listings with their existing Cayley path contracts.
	for _, tt := range []struct {
		// list collects a production listing.
		list func(context.Context, world.WorldState, ...string) ([]string, error)
		// path selects the equivalent Cayley traversal.
		path func(*cayley.Path) *cayley.Path
	}{
		{ListManifestCandidates, NewListManifestCandidatePath},
		{ListManifests, NewListManifestPath},
	} {
		want, err := world.CollectPathWithKeys(ctx, tb.WorldState, []string{"traversal/0", "traversal/0"}, func(p *cayley.Path) (*cayley.Path, error) {
			return tt.path(p), nil
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := tt.list(ctx, tb.WorldState, "traversal/0", "traversal/0")
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(got, want) {
			t.Fatalf("listing = %v, want path results %v", got, want)
		}
	}
}
