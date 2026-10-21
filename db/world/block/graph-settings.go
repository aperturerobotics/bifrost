package world_block

import (
	"github.com/aperturerobotics/cayley/graph"
	cayley_kv "github.com/aperturerobotics/cayley/graph/kv"
	"github.com/pkg/errors"
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
