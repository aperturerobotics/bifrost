//go:build !js

package spacewave_cli

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/aperturerobotics/cayley/graph"
	cayley_kv "github.com/aperturerobotics/cayley/graph/kv"
	"github.com/aperturerobotics/cayley/quad"
	"github.com/aperturerobotics/cli"
	block_gc "github.com/s4wave/spacewave/db/block/gc"
	kvtx_cayley "github.com/s4wave/spacewave/db/kvtx/cayley"
	kvtx_prefixer "github.com/s4wave/spacewave/db/kvtx/prefixer"
	volume_s4db "github.com/s4wave/spacewave/db/volume/s4db"
	"github.com/sirupsen/logrus"
)

// TestDebugVolumeReindex preserves a stopped volume's references across
// regeneration, reopening, and subsequent ownership changes.
func TestDebugVolumeReindex(t *testing.T) {
	// Create an isolated volume whose graph uses the former forward layout.
	ctx := t.Context()
	path := filepath.Join(t.TempDir(), "volume.s4wave")
	conf := &volume_s4db.Config{Path: path}
	vol, err := volume_s4db.NewVolume(ctx, logrus.NewEntry(logrus.New()), conf)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := vol.Close(); err != nil {
			t.Error(err)
		}
	})

	// Open the former graph layout over the volume's GC namespace.
	g, err := kvtx_cayley.NewGraph(ctx, kvtx_prefixer.NewPrefixer(vol.GetKvtxStore(), []byte("gc/")), graph.Options{cayley_kv.OptQuadIndexes: cayley_kv.DefaultQuadIndexes})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := g.Close(); err != nil {
			t.Error(err)
		}
	})
	edges := []block_gc.RefEdge{
		{Subject: block_gc.NodeGCRoot, Object: "bucket:test"},
		{Subject: "bucket:test", Object: "root:head"},
		{Subject: "root:head", Object: "retained"},
		{Subject: "retained", Object: "child"},
		{Subject: block_gc.NodeUnreferenced, Object: "orphan"},
	}
	deltas := make([]graph.Delta, len(edges))
	for i, edge := range edges {
		deltas[i] = graph.Delta{
			Quad: quad.MakeIRI(edge.Subject, block_gc.PredGCRef, edge.Object, ""), Action: graph.Add,
		}
	}
	if err := g.ApplyDeltas(ctx, deltas, graph.IgnoreOpts{}); err != nil {
		t.Fatal(err)
	}
	if err := vol.Close(); err != nil {
		t.Fatal(err)
	}

	// Refuse the former layout before any reindex command has run.
	oldVolume, err := volume_s4db.NewVolume(ctx, logrus.NewEntry(logrus.New()), conf)
	if oldVolume != nil {
		if closeErr := oldVolume.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	if err == nil {
		t.Fatal("opened an old-layout volume before reindexing")
	}
	for _, detail := range []string{"sp,ops", "spo,ops", "spacewave debug volume-reindex"} {
		if !strings.Contains(err.Error(), detail) {
			t.Fatalf("open error lacks %q: %v", detail, err)
		}
	}

	// Run the real offline command and require its completion output.
	var output bytes.Buffer
	app := &cli.App{Writer: &output, Commands: []*cli.Command{newDebugCommand(nil)}}
	if err := app.RunContext(ctx, []string{"spacewave", "debug", "volume-reindex", path}); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(strings.Split(output.String(), "\n"), "References Indexed:    5") {
		t.Fatalf("completion output = %q", output.String())
	}

	// Reopen and require both traversal directions and the orphan marker.
	vol, err = volume_s4db.NewVolume(ctx, logrus.NewEntry(logrus.New()), conf)
	if err != nil {
		t.Fatal(err)
	}
	rg := vol.GetRefGraph()
	for _, edge := range edges {
		outgoing, err := rg.GetOutgoingRefs(ctx, edge.Subject)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(outgoing, edge.Object) {
			t.Fatalf("missing outgoing edge %v", edge)
		}
		incoming, err := rg.GetIncomingRefs(ctx, edge.Object)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(incoming, edge.Subject) {
			t.Fatalf("missing incoming edge %v", edge)
		}
	}
	orphans, err := rg.GetUnreferencedNodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(orphans, []string{"orphan"}) {
		t.Fatalf("orphan markers = %v", orphans)
	}

	// Transfer the root through the same interface used by publication.
	if err := rg.ApplyRefBatch(ctx,
		[]block_gc.RefEdge{{Subject: "root:head", Object: "child"}},
		[]block_gc.RefEdge{{Subject: "root:head", Object: "retained"}},
	); err != nil {
		t.Fatal(err)
	}
	if found, err := rg.HasIncomingRefs(ctx, "retained"); err != nil || found {
		t.Fatalf("released root still owned: %v, %v", found, err)
	}
}

// TestDebugVolumeReindexUsage exposes offline requirements and rejects ambiguous targets.
func TestDebugVolumeReindexUsage(t *testing.T) {
	// Require the registered command's help to explain interruption recovery.
	var output bytes.Buffer
	app := &cli.App{Writer: &output, Commands: []*cli.Command{newDebugCommand(nil)}}
	if err := app.RunContext(t.Context(), []string{"spacewave", "debug", "volume-reindex", "--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "<volume-file>") || !strings.Contains(output.String(), "rerun") {
		t.Fatalf("offline usage = %q", output.String())
	}

	// Reject absent or multiple volume paths before trying to open any file.
	for _, args := range [][]string{nil, {"first", "second"}} {
		app := &cli.App{Writer: &output, Commands: []*cli.Command{newDebugCommand(nil)}}
		command := append([]string{"spacewave", "debug", "volume-reindex"}, args...)
		if err := app.RunContext(t.Context(), command); err == nil || err.Error() != "expected one volume file argument" {
			t.Fatalf("arguments %v: %v", args, err)
		}
	}
}
