package plugin

import "sync"

// summaryCacheSize bounds the number of distinct summary queries kept. The
// dashboard polls one query at a time; a few slots cover the management route,
// the dashboard route and a quick filter change back and forth.
const summaryCacheSize = 8

// summaryCache is a tiny LRU of computed summary aggregates. Keys embed the
// store version, so entries for an older version are never returned; they just
// age out. The zero value is ready to use.
type summaryCache struct {
	mu      sync.Mutex
	entries []summaryCacheEntry // most recently used last
}

type summaryCacheEntry struct {
	key string
	agg *summaryAggregate
}

func (c *summaryCache) get(key string) (*summaryAggregate, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, e := range c.entries {
		if e.key == key {
			copy(c.entries[i:], c.entries[i+1:])
			c.entries[len(c.entries)-1] = e
			return e.agg, true
		}
	}
	return nil, false
}

func (c *summaryCache) put(key string, agg *summaryAggregate) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, e := range c.entries {
		if e.key == key {
			c.entries = append(c.entries[:i], c.entries[i+1:]...)
			break
		}
	}
	if len(c.entries) >= summaryCacheSize {
		c.entries = append(c.entries[:0], c.entries[1:]...)
	}
	c.entries = append(c.entries, summaryCacheEntry{key, agg})
}
