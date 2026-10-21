package world_block

import (
	"context"

	"github.com/aperturerobotics/cayley/graph"
	cayley_kv "github.com/aperturerobotics/cayley/graph/kv"
	"github.com/pkg/errors"
	"github.com/s4wave/spacewave/db/tx"
)

// NewGraphSettings builds the settings holding indexes. The default indexes
// are the empty settings.
func NewGraphSettings(indexes []cayley_kv.QuadIndex) (*GraphSettings, error) {
	if len(indexes) == 0 || cayley_kv.CompareQuadIndexes(indexes, cayley_kv.DefaultQuadIndexes) {
		return nil, nil
	}
	metas, err := cayley_kv.QuadIndexMetas(indexes)
	if err != nil {
		return nil, err
	}
	return &GraphSettings{Indexes: metas}, nil
}

// QuadIndexes returns the quad indexes with their states, or nil for the
// default indexes.
func (s *GraphSettings) QuadIndexes() ([]cayley_kv.QuadIndex, error) {
	if len(s.GetIndexes()) == 0 {
		return nil, nil
	}
	return cayley_kv.QuadIndexesFromMetas(s.GetIndexes())
}

// ValidateInitial checks the settings a new World may start with: every index
// must be complete, since a World starts with no change in progress.
func (s *GraphSettings) ValidateInitial() error {
	indexes, err := s.QuadIndexes()
	if err != nil {
		return err
	}
	for _, index := range indexes {
		if len(index.Dirs) == 0 {
			return errors.New("graph settings: index has no directions")
		}
		if index.State != cayley_kv.QuadIndexState_COMPLETE {
			return errors.Errorf("graph settings: initial index is %s", index.State)
		}
	}
	return nil
}

// graphOptions returns the options that open the graph store under the
// settings without reading them from the store.
func (s *GraphSettings) graphOptions() (graph.Options, error) {
	indexes, err := s.QuadIndexes()
	if err != nil {
		return nil, err
	}
	if indexes == nil {
		// The default indexes: the store reads no layout.
		return graph.Options{cayley_kv.OptAssumeDefaultIdx: true}, nil
	}
	return graph.Options{cayley_kv.OptTrustQuadIndexes: indexes}, nil
}

// StepGraphSettings advances the World's graph toward the target settings by
// at most max links. It reports whether the graph matches target and whether
// this call changed the World. Expects rmtx to be locked.
//
// The first call records the change in the graph store and later calls
// continue it; a different change in progress is an error. Each call is a
// function of the World and its arguments only, so every replay of the same
// calls reaches the same World.
func (t *WorldState) StepGraphSettings(ctx context.Context, target *GraphSettings, max int) (done, changed bool, err error) {
	// Reject a read-only state and an invalid target.
	if !t.write {
		return false, false, tx.ErrNotWrite
	}
	if err := target.ValidateInitial(); err != nil {
		return false, false, err
	}
	indexes, err := target.QuadIndexes()
	if err != nil {
		return false, false, err
	}
	if indexes == nil {
		indexes = cayley_kv.DefaultQuadIndexes
	}

	// Skip a graph that already matches.
	qs, ok := t.graphHd.QuadStore.(*cayley_kv.QuadStore)
	if !ok {
		return false, false, errors.New("world graph is not a cayley kv store")
	}
	if cayley_kv.CompareQuadIndexes(qs.Layout(), indexes) {
		return true, false, nil
	}

	// Begin or continue the change and advance it one step.
	if err := qs.BeginIndexChange(ctx, indexes); err != nil {
		return false, false, err
	}
	done, err = qs.StepIndexChange(ctx, max)
	if err != nil {
		return false, false, err
	}

	// Record the graph's new layout in the root, which opens it next time.
	r, err := t.GetRoot(ctx)
	if err != nil {
		return false, false, err
	}
	r.GraphSettings, err = NewGraphSettings(qs.Layout())
	if err != nil {
		return false, false, err
	}
	t.bcs.MarkDirty()
	return done, true, nil
}
