package plugin

import (
	"slices"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
)

// recordPageHeap keeps the oldest selected record at the root. Only indexes
// into the read-only snapshot are retained, never copies of full LogRecords.
type recordPageHeap struct {
	records []store.LogRecord
	indexes []int
}

func (h *recordPageHeap) newer(i, j int) bool {
	a, b := &h.records[i], &h.records[j]
	if cmp := a.EffectiveTimestamp().Compare(b.EffectiveTimestamp()); cmp != 0 {
		return cmp > 0
	}
	// Snapshot sequences are append ordinals. Descending sequence reproduces
	// FilterRecords' reverse-append order followed by a stable timestamp sort.
	return a.Sequence > b.Sequence
}

func (h *recordPageHeap) push(index int) {
	h.indexes = append(h.indexes, index)
	for child := len(h.indexes) - 1; child > 0; {
		parent := (child - 1) / 2
		if !h.newer(h.indexes[parent], h.indexes[child]) {
			break
		}
		h.indexes[parent], h.indexes[child] = h.indexes[child], h.indexes[parent]
		child = parent
	}
}

func (h *recordPageHeap) replaceOldest(index int) {
	h.indexes[0] = index
	for parent := 0; ; {
		child := 2*parent + 1
		if child >= len(h.indexes) {
			return
		}
		if right := child + 1; right < len(h.indexes) && h.newer(h.indexes[child], h.indexes[right]) {
			child = right
		}
		if !h.newer(h.indexes[parent], h.indexes[child]) {
			return
		}
		h.indexes[parent], h.indexes[child] = h.indexes[child], h.indexes[parent]
		parent = child
	}
}

// selectRecordsPage scans the entire original frontier: append order need not
// follow request timestamps. Handlers validate limit (1..500) before calling.
// Keeping limit+1 indexes determines has_more in O(N log(limit+1)) time and
// O(limit+1) memory; only the returned page is copied.
func selectRecordsPage(records []store.LogRecord, opts store.FilterOptions, cursor *recordsCursor, limit int) (page []store.LogRecord, more, anchorFound bool) {
	page, more, anchorFound, _ = selectRecordsPageWithTotal(records, opts, cursor, limit)
	return page, more, anchorFound
}

func selectRecordsPageWithTotal(records []store.LogRecord, opts store.FilterOptions, cursor *recordsCursor, limit int) (page []store.LogRecord, more, anchorFound bool, total int) {
	k := limit + 1
	h := recordPageHeap{records: records, indexes: make([]int, 0, min(k, len(records)))}
	anchorFound = cursor == nil
	for i := range records {
		r := &records[i]
		if !store.MatchRecord(r, &opts) {
			continue
		}
		total++
		if cursor != nil {
			ts := r.EffectiveTimestamp()
			if r.Sequence == cursor.Sequence && ts.Equal(cursor.Timestamp) {
				anchorFound = true
			}
			if ts.After(cursor.Timestamp) || (ts.Equal(cursor.Timestamp) && r.Sequence >= cursor.Sequence) {
				continue
			}
		}
		if len(h.indexes) < k {
			h.push(i)
		} else if h.newer(i, h.indexes[0]) {
			h.replaceOldest(i)
		}
	}
	if !anchorFound {
		return nil, false, false, 0
	}
	slices.SortFunc(h.indexes, func(i, j int) int {
		if h.newer(i, j) {
			return -1
		}
		if h.newer(j, i) {
			return 1
		}
		return 0
	})
	more = len(h.indexes) > limit
	page = make([]store.LogRecord, min(limit, len(h.indexes)))
	for i := range page {
		page[i] = records[h.indexes[i]]
	}
	return page, more, true, total
}

// selectRecordsPageByNumber returns a 1-indexed page in the exact same newest-first
// order as selectRecordsPage, plus total matching count and has_more.
func selectRecordsPageByNumber(records []store.LogRecord, opts store.FilterOptions, pageNum, limit int) (page []store.LogRecord, total int, more bool) {
	h := recordPageHeap{records: records}
	indexes := make([]int, 0, min(len(records), 64))
	for i := range records {
		if store.MatchRecord(&records[i], &opts) {
			indexes = append(indexes, i)
		}
	}
	total = len(indexes)
	if total == 0 || pageNum < 1 || limit < 1 {
		return []store.LogRecord{}, total, false
	}
	start := (pageNum - 1) * limit
	if start >= total {
		return []store.LogRecord{}, total, false
	}
	slices.SortFunc(indexes, func(i, j int) int {
		if h.newer(i, j) {
			return -1
		}
		if h.newer(j, i) {
			return 1
		}
		return 0
	})
	end := min(start+limit, total)
	more = end < total
	page = make([]store.LogRecord, end-start)
	for i := start; i < end; i++ {
		page[i-start] = records[indexes[i]]
	}
	return page, total, more
}
