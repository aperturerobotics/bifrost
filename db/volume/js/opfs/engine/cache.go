package engine

import "context"

// cachedMessage reuses immutable decoded metadata without repeating protobuf work.
func (e *Engine) cachedMessage(ctx context.Context, name string) (message, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	e.mtx.Lock()
	defer e.mtx.Unlock()
	if e.closed {
		return nil, ErrClosed
	}
	if elem := e.cache[name]; elem != nil {
		e.recency.MoveToFront(elem)
		return elem.Value.(*cacheEntry).parsed, nil
	}
	return nil, nil
}

// cacheMessage charges decoded metadata to the same byte and entry budgets.
// The conservative charge covers generated record objects and copied fields.
func (e *Engine) cacheMessage(name string, parsed message, decodedCharge int) {
	e.mtx.Lock()
	defer e.mtx.Unlock()
	if e.closed {
		return
	}
	elem := e.cache[name]
	if elem == nil {
		return
	}
	entry := elem.Value.(*cacheEntry)
	if entry.parsed != nil || entry.charge+decodedCharge > cacheByteLimit {
		return
	}
	e.recency.MoveToFront(elem)
	for e.cacheBytes+decodedCharge > cacheByteLimit {
		victim := e.recency.Back()
		old := victim.Value.(*cacheEntry)
		delete(e.cache, old.name)
		e.cacheBytes -= old.charge
		e.recency.Remove(victim)
	}
	entry.parsed = parsed
	entry.charge += decodedCharge
	e.cacheBytes += decodedCharge
}
