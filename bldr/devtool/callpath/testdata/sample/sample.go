// Package sample is the call graph the callpath tests search.
package sample

import "sync"

// Store holds a counter behind a mutex.
type Store struct {
	mu sync.Mutex
	n  int
}

// Op is applied to a Store.
type Op interface {
	Apply(s *Store)
}

// AddOp increments the counter under the lock.
type AddOp struct{}

// Apply increments the counter.
func (AddOp) Apply(s *Store) { s.add() }

// ReadOp reads the counter without the lock.
type ReadOp struct{}

// Apply reads the counter.
func (ReadOp) Apply(s *Store) { _ = s.read() }

func (s *Store) add() {
	s.mu.Lock()
	s.n++
	s.mu.Unlock()
}

func (s *Store) read() int { return s.n }

// Run applies op to s.
func Run(s *Store, op Op) { op.Apply(s) }
