package kvtx_deferred

import (
	"bytes"

	"github.com/s4wave/spacewave/db/kvtx"
	"github.com/tidwall/btree"
)

// source is where the merge iterator's current entry comes from.
type source int

const (
	// srcNone is no current entry.
	srcNone source = iota
	// srcHeld is a held write.
	srcHeld
	// srcUnder is the underlying iterator.
	srcUnder
)

// mergeIterator merges a snapshot of held writes into a sorted forward
// iterator over the underlying keys. A held write replaces the underlying entry
// with its key, and a held delete hides it.
type mergeIterator struct {
	// under is the underlying iterator.
	under kvtx.Iterator
	// held is the snapshot of the held writes.
	held *btree.BTreeG[*entry]
	// heldIt iterates held.
	heldIt btree.IterG[*entry]
	// heldOK is set while heldIt is on a write with the prefix.
	heldOK bool
	// prefix bounds the held writes.
	prefix []byte
	// started is set once the iterator is positioned.
	started bool
	// src is the source of the current entry.
	src source
}

// newMergeIterator constructs a merge iterator.
func newMergeIterator(under kvtx.Iterator, held *btree.BTreeG[*entry], prefix []byte) *mergeIterator {
	return &mergeIterator{under: under, held: held, heldIt: held.Iter(), prefix: prefix}
}

// heldOn reports whether ok left heldIt on a write with the prefix.
func (m *mergeIterator) heldOn(ok bool) bool {
	return ok && bytes.HasPrefix(m.heldIt.Item().key, m.prefix)
}

// settle finds the next visible entry at or after both positions.
func (m *mergeIterator) settle() bool {
	for {
		// Stop when both sources are done or the underlying one failed.
		underOK := m.under.Valid()
		if m.under.Err() != nil || (!m.heldOK && !underOK) {
			m.src = srcNone
			return false
		}

		// Take the held write when it is first, stepping past the underlying
		// entry it replaces and past a delete.
		if m.heldOK {
			e := m.heldIt.Item()
			c := -1
			if underOK {
				c = bytes.Compare(e.key, m.under.Key())
			}
			if c <= 0 {
				if c == 0 {
					m.under.Next()
				}
				if e.del {
					m.heldOK = m.heldOn(m.heldIt.Next())
					continue
				}
				m.src = srcHeld
				return true
			}
		}
		m.src = srcUnder
		return true
	}
}

// Err returns any error that has closed the iterator.
func (m *mergeIterator) Err() error {
	return m.under.Err()
}

// Valid returns if the iterator points to a valid entry.
func (m *mergeIterator) Valid() bool {
	return m.src != srcNone && m.under.Err() == nil
}

// Key returns the current entry key, or nil if not valid.
func (m *mergeIterator) Key() []byte {
	switch {
	case !m.Valid():
		return nil
	case m.src == srcHeld:
		return m.heldIt.Item().key
	default:
		return m.under.Key()
	}
}

// Value returns the current entry value, or nil if not valid.
func (m *mergeIterator) Value() ([]byte, error) {
	switch {
	case !m.Valid():
		return nil, m.Err()
	case m.src == srcHeld:
		return m.heldIt.Item().value, nil
	default:
		return m.under.Value()
	}
}

// ValueCopy copies the value to the given byte slice and returns it.
func (m *mergeIterator) ValueCopy(buf []byte) ([]byte, error) {
	switch {
	case !m.Valid():
		return nil, m.Err()
	case m.src == srcHeld:
		return append(buf[:0], m.heldIt.Item().value...), nil
	default:
		return m.under.ValueCopy(buf)
	}
}

// Next advances to the next entry and returns Valid.
func (m *mergeIterator) Next() bool {
	if !m.started {
		return m.Seek(nil) == nil && m.Valid()
	}
	switch m.src {
	case srcHeld:
		m.heldOK = m.heldOn(m.heldIt.Next())
	case srcUnder:
		m.under.Next()
	default:
		return false
	}
	return m.settle()
}

// Seek moves the iterator to the first key >= k.
func (m *mergeIterator) Seek(k []byte) error {
	// Seek the underlying iterator first, leaving no entry when it fails.
	if err := m.under.Seek(k); err != nil {
		m.src = srcNone
		return err
	}

	// Seek the held writes to the same key, the prefix standing for nil.
	if k == nil {
		k = m.prefix
	}
	m.started = true
	m.heldOK = m.heldOn(m.heldIt.Seek(&entry{key: k}))
	m.settle()
	return m.under.Err()
}

// Close closes the iterator.
func (m *mergeIterator) Close() {
	m.heldIt.Release()
	m.under.Close()
	m.src = srcNone
}

// _ is a type assertion
var _ kvtx.Iterator = (*mergeIterator)(nil)
