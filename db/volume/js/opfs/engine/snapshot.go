package engine

import (
	"bytes"
	"context"
	"slices"
	"sort"
)

// snapshot pins immutable files across one read operation.
type snapshot struct {
	// engine supplies validated immutable file reads.
	engine *Engine
	// root fixes the operation's complete committed generation.
	root *Root
	// release relinquishes the cross-runtime reclamation lock.
	release func()
}

// snapshot acquires file protection before resolving the durable root.
func (e *Engine) snapshot(ctx context.Context) (*snapshot, error) {
	release, err := e.protect(ctx)
	if err != nil {
		return nil, err
	}
	root, err := e.loadRoot(ctx)
	if err != nil {
		release()
		return nil, err
	}
	return &snapshot{engine: e, root: root, release: release}, nil
}

// get resolves a full key through one partition's bounded newest-first runs.
func (s *snapshot) get(ctx context.Context, key []byte) ([]byte, bool, error) {
	name := s.root.Catalogue
	for {
		page, err := s.engine.readCatalogue(ctx, name)
		if err != nil {
			return nil, false, err
		}
		if len(page.Children) != 0 {
			i := sort.Search(len(page.Children), func(i int) bool { return bytes.Compare(page.Children[i].Lower, key) > 0 }) - 1
			if i < 0 {
				return nil, false, nil
			}
			name = page.Children[i].File
			continue
		}
		i := sort.Search(len(page.Partitions), func(i int) bool { return bytes.Compare(page.Partitions[i].Lower, key) > 0 }) - 1
		if i < 0 {
			return nil, false, nil
		}
		runs := page.Partitions[i].Runs
		for _, run := range slices.Backward(runs) {
			run, err := s.engine.readRun(ctx, run)
			if err != nil {
				return nil, false, err
			}
			i := sort.Search(len(run.Records), func(i int) bool { return bytes.Compare(run.Records[i].Key, key) >= 0 })
			if i < len(run.Records) && bytes.Equal(run.Records[i].Key, key) {
				record := run.Records[i]
				return record.Value, !record.Deleted, nil
			}
		}
		return nil, false, nil
	}
}

// Get returns a caller-owned value and the generation that supplied it.
func (e *Engine) Get(ctx context.Context, key []byte) ([]byte, bool, uint64, error) {
	s, err := e.snapshot(ctx)
	if err != nil {
		return nil, false, 0, err
	}
	defer s.release()
	value, found, err := s.get(ctx, key)
	return bytes.Clone(value), found, s.root.Revision, err
}
