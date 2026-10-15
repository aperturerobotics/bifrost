package s4db

import "testing"

// TestCacheEvictsLeavesFirst keeps inner pages within their share of a shard
// cached while leaves stream through it.
func TestCacheEvictsLeavesFirst(t *testing.T) {
	// Cache inner pages in one shard up to half its budget.
	c := newCache(8 * pageSize * cacheShards)
	s := c.shard(0)
	var pages []uint64
	for p := uint64(0); len(pages) < 8*4; p++ {
		if c.shard(p) == s {
			pages = append(pages, p)
		}
	}
	inner, leaves := pages[:4], pages[4:]
	for _, p := range inner {
		c.put(p, &node{cost: pageSize})
	}

	// Read the inner pages once, then stream leaves through the shard.
	for _, p := range inner {
		c.get(p)
	}
	for _, p := range leaves {
		c.put(p, &node{leaf: true, cost: pageSize})
	}
	for _, p := range inner {
		if c.get(p) == nil {
			t.Fatalf("inner page %d evicted while leaves remained", p)
		}
	}
}
