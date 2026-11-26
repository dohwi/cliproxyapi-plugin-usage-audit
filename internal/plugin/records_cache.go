package plugin

import (
	"sync"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
)

// recordsCacheSize bounds the number of distinct paginated records queries
// cached per store version. Covers initial loads, filter toggles, and back/forth
// page navigation without re-scanning when no new records arrived.
const recordsCacheSize = 16

type recordsPageResult struct {
	page       []store.LogRecord
	nextCursor string
	totalCount int
	hasMore    bool
}

type recordsCacheEntry struct {
	key string
	res *recordsPageResult
}

// recordsCache is a tiny LRU of paginated query results keyed by store version
// plus query parameters. Stored results must be treated as read-only.
type recordsCache struct {
	mu      sync.Mutex
	entries []recordsCacheEntry // most recently used last
}

func (c *recordsCache) get(key string) (*recordsPageResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, e := range c.entries {
		if e.key == key {
			copy(c.entries[i:], c.entries[i+1:])
			c.entries[len(c.entries)-1] = e
			return e.res, true
		}
	}
	return nil, false
}

func (c *recordsCache) put(key string, res *recordsPageResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, e := range c.entries {
		if e.key == key {
			c.entries = append(c.entries[:i], c.entries[i+1:]...)
			break
		}
	}
	if len(c.entries) >= recordsCacheSize {
		c.entries = append(c.entries[:0], c.entries[1:]...)
	}
	c.entries = append(c.entries, recordsCacheEntry{key, res})
}
