package block_gc

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/s4wave/spacewave/db/block"
	block_mock "github.com/s4wave/spacewave/db/block/mock"
	block_store "github.com/s4wave/spacewave/db/block/store"
	"github.com/s4wave/spacewave/db/packfile/writer"
	"github.com/s4wave/spacewave/net/hash"
)

// TestGCStoreOpsGetStoredBlockUntracked checks that stored bytes without graph
// edges are not reported as a known leaf.
func TestGCStoreOpsGetStoredBlockUntracked(t *testing.T) {
	// Store a child and its parent without GC tracking.
	env := newGCTestEnv(t)
	child, _, err := block.PutBlock(env.ctx, env.rawStore, block_mock.NewExample("untracked child"))
	if err != nil {
		t.Fatal(err)
	}
	parent := &block_mock.Root{ExampleSubBlock: &block_mock.SubBlock{ExamplePtr: child}}
	ref, _, err := block.PutBlock(env.ctx, env.rawStore, parent)
	if err != nil {
		t.Fatal(err)
	}

	// Confirm that the stored parent has no graph edge in either direction.
	outgoing, err := env.refGraph.GetOutgoingRefs(env.ctx, BlockIRI(ref))
	if err != nil {
		t.Fatal(err)
	}
	incoming, err := env.refGraph.GetIncomingRefs(env.ctx, BlockIRI(ref))
	if err != nil || len(outgoing) != 0 || len(incoming) != 0 {
		t.Fatalf("untracked graph: outgoing=%v incoming=%v err=%v", outgoing, incoming, err)
	}

	// Preserve the payload while reporting that its references are unknown.
	stored, err := env.gcStore.GetStoredBlock(env.ctx, ref)
	if err != nil || stored == nil || stored.RefsKnown || len(stored.Refs) != 0 {
		t.Fatalf("untracked block = %+v/%v, want unknown refs", stored, err)
	}
	decoded := &block_mock.Root{}
	if err := decoded.UnmarshalVT(stored.Data); err != nil {
		t.Fatal(err)
	}
	refs, err := block.ExtractBlockRefs(decoded)
	if err != nil || len(refs) != 1 || !refs[0].EqualsRef(child) {
		t.Fatalf("payload refs = %v/%v, want [%v]", refs, err, child)
	}

	// Opaque graph consumers must report unknown refs instead of truncating the DAG.
	destination := newGCTestEnv(t)
	if err := block.CopyGraph(env.ctx, env.gcStore, destination.gcStore, ref, nil); !errors.Is(err, block.ErrRefsUnknown) {
		t.Fatalf("untracked copy = %v, want ErrRefsUnknown", err)
	}
	if err := block.WalkGraph(env.ctx, env.gcStore, []*block.BlockRef{ref}, 1, nil); !errors.Is(err, block.ErrRefsUnknown) {
		t.Fatalf("untracked walk = %v, want ErrRefsUnknown", err)
	}
	yielded := false
	_, err = writer.PackBlocks(io.Discard, func() (*hash.Hash, *block.StoredBlock, error) {
		if yielded {
			return nil, nil, nil
		}
		yielded = true
		return ref.GetHash(), stored, nil
	})
	if !errors.Is(err, block.ErrRefsUnknown) {
		t.Fatalf("untracked pack = %v, want ErrRefsUnknown", err)
	}

	// Read-through serves the untracked bytes without filling a false leaf.
	readThrough := block_store.NewStoreReadThrough(
		func() block.StoreOps { return destination.gcStore },
		func() block.StoreOps { return env.gcStore },
		block_store.CacheFill,
	)
	got, err := readThrough.GetStoredBlock(env.ctx, ref)
	if err != nil || got == nil || got.RefsKnown {
		t.Fatalf("untracked read-through = %+v/%v, want unknown refs", got, err)
	}
	if destination.blockExists(t, ref) {
		t.Fatal("untracked block was cached as a leaf")
	}

	// A tracked re-put exposes its supplied refs even before the graph is repaired.
	if _, _, err := block.PutBlock(env.ctx, env.gcStore, decoded); err != nil {
		t.Fatal(err)
	}
	stored, err = env.gcStore.GetStoredBlock(env.ctx, ref)
	if err != nil || stored == nil || !stored.RefsKnown || len(stored.Refs) != 1 || !stored.Refs[0].EqualsRef(child) {
		t.Fatalf("buffered re-put = %+v/%v, want child refs", stored, err)
	}
}

// TestGCStoreOpsGetStoredBlockLeaf checks leaf knowledge before and after its
// ownership edges reach the graph.
func TestGCStoreOpsGetStoredBlockLeaf(t *testing.T) {
	// Exercise parent ownership and unreferenced cache staging independently.
	for _, parent := range []string{"", BucketIRI("leaf-owner")} {
		t.Run(parent, func(t *testing.T) {
			// Store a leaf with its ownership still buffered.
			env := newGCTestEnvWithParent(t, parent)
			ref, _, err := env.gcStore.PutBlock(env.ctx, []byte("tracked leaf"), nil)
			if err != nil {
				t.Fatal(err)
			}

			// Require a known leaf on both sides of the flush.
			for _, flushed := range []bool{false, true} {
				if flushed {
					env.flush(t)
				}
				stored, err := env.gcStore.GetStoredBlock(env.ctx, ref)
				if err != nil || stored == nil || !stored.RefsKnown || len(stored.Refs) != 0 {
					t.Fatalf("flushed=%v: leaf = %+v/%v", flushed, stored, err)
				}
			}
		})
	}
}

// stallGraph holds ApplyRefBatch until it is released, then applies the batch
// to the graph it wraps.
type stallGraph struct {
	*RefGraph
	// applying is closed when ApplyRefBatch starts.
	applying chan struct{}
	// release lets ApplyRefBatch apply the batch.
	release chan struct{}
	// read is closed when the first incoming lookup has answered.
	read     chan struct{}
	readOnce sync.Once
}

// ApplyRefBatch applies the batch after it is released.
func (g *stallGraph) ApplyRefBatch(ctx context.Context, adds, removes []RefEdge) error {
	close(g.applying)
	<-g.release
	return g.RefGraph.ApplyRefBatch(ctx, adds, removes)
}

// GetIncomingRefs reads the wrapped graph and reports the first lookup.
func (g *stallGraph) GetIncomingRefs(ctx context.Context, node string) ([]string, error) {
	sources, err := g.RefGraph.GetIncomingRefs(ctx, node)
	g.readOnce.Do(func() { close(g.read) })
	return sources, err
}

// TestGCStoreOpsGetStoredBlockDuringFlush checks that a block whose edges a
// running flush has taken but not applied is not reported as untracked.
func TestGCStoreOpsGetStoredBlockDuringFlush(t *testing.T) {
	// Store a leaf and start a flush that stalls while applying its edges.
	env := newGCTestEnv(t)
	graph := &stallGraph{
		RefGraph: env.refGraph,
		applying: make(chan struct{}),
		release:  make(chan struct{}),
		read:     make(chan struct{}),
	}
	env.gcStore = NewGCStoreOps(env.rawStore, graph)
	ref, _, err := env.gcStore.PutBlock(env.ctx, []byte("flushing leaf"), nil)
	if err != nil {
		t.Fatal(err)
	}
	flushed := make(chan error, 1)
	go func() { flushed <- env.gcStore.FlushPending(env.ctx) }()
	<-graph.applying

	// Read the block while the flush holds its edges, and finish the flush
	// once the read has looked for them.
	type result struct {
		stored *block.StoredBlock
		err    error
	}
	read := make(chan result, 1)
	go func() {
		stored, err := env.gcStore.GetStoredBlock(env.ctx, ref)
		read <- result{stored, err}
	}()
	<-graph.read
	close(graph.release)
	if err := <-flushed; err != nil {
		t.Fatal(err)
	}

	// The read waits out the flush and finds the applied edge.
	got := <-read
	if got.err != nil || got.stored == nil || !got.stored.RefsKnown {
		t.Fatalf("block during flush = %+v/%v, want known refs", got.stored, got.err)
	}
}
