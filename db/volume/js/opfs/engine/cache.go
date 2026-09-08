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

// cacheBytesLocked retains immutable bytes under the shared byte and entry budgets.
// The caller holds mtx and has checked that the engine remains open.
func (e *Engine) cacheBytesLocked(key string, data []byte) []byte {
	if elem := e.cache[key]; elem != nil {
		e.recency.MoveToFront(elem)
		return elem.Value.(*cacheEntry).data
	}
	charge := len(data) + len(key) + 192
	for e.cacheBytes+charge > cacheByteLimit || len(e.cache) >= cacheFileLimit {
		elem := e.recency.Back()
		entry := elem.Value.(*cacheEntry)
		delete(e.cache, entry.name)
		e.cacheBytes -= entry.charge
		e.recency.Remove(elem)
	}
	e.cache[key] = e.recency.PushFront(&cacheEntry{name: key, data: data, charge: charge})
	e.cacheBytes += charge
	return data
}
