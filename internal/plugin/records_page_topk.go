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

func (h *recordPageHeap) compare(i, j int) int {
	if h.newer(i, j) {
		return -1
	}
	if h.newer(j, i) {
		return 1
	}
	return 0
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

func hasNonTimeFilters(opts *store.FilterOptions) bool {
	return opts.Provider != "" || opts.Model != "" || opts.Status != "" ||
		opts.Source != "" || opts.ExecutorType != "" || opts.Endpoint != "" || opts.AuthIndex != ""
}

func blockOutsideRange(blk *store.BlockMeta, opts *store.FilterOptions) bool {
	if opts.From != nil && blk.MaxTS.Before(*opts.From) {
		return true
	}
	if opts.To != nil && blk.MinTS.After(*opts.To) {
		return true
	}
	return false
}

func blockFullyInsideRange(blk *store.BlockMeta, opts *store.FilterOptions) bool {
	if opts.From != nil && blk.MinTS.Before(*opts.From) {
		return false
	}
	if opts.To != nil && blk.MaxTS.After(*opts.To) {
		return false
	}
	return true
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
	return selectRecordsPageWithTotalBlocks(records, nil, opts, cursor, limit)
}

func selectRecordsPageWithTotalBlocks(records []store.LogRecord, blocks []store.BlockMeta, opts store.FilterOptions, cursor *recordsCursor, limit int) (page []store.LogRecord, more, anchorFound bool, total int) {
	k := limit + 1
	h := recordPageHeap{records: records, indexes: make([]int, 0, min(k, len(records)))}
	anchorFound = cursor == nil
	timeOnly := !hasNonTimeFilters(&opts)

	scanRange := func(start, end int) {
		for i := end - 1; i >= start; i-- {
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
	}

	if len(blocks) == 0 {
		scanRange(0, len(records))
	} else {
		n := len(records)
		for bi := len(blocks) - 1; bi >= 0; bi-- {
			blk := &blocks[bi]
			if blk.Start >= n {
				continue
			}
			blkEnd := blk.End
			if blkEnd > n {
				blkEnd = n
			}
			if blockOutsideRange(blk, &opts) {
				continue
			}
			// Fast path: unfiltered time-range block strictly older than every
			// item already in the top-K heap only contributes to total count.
			if cursor == nil && timeOnly && blkEnd == blk.End &&
				len(h.indexes) == k && blockFullyInsideRange(blk, &opts) &&
				blk.MaxTS.Before(records[h.indexes[0]].EffectiveTimestamp()) {
				total += blkEnd - blk.Start
				continue
			}
			scanRange(blk.Start, blkEnd)
		}
	}

	if !anchorFound {
		return nil, false, false, 0
	}
	slices.SortFunc(h.indexes, h.compare)
	more = len(h.indexes) > limit
	page = make([]store.LogRecord, min(limit, len(h.indexes)))
	for i := range page {
		page[i] = records[h.indexes[i]]
	}
	return page, more, true, total
}

const maxBoundedPageHeap = 2048

// selectRecordsPageByNumber returns a 1-indexed page in the exact same newest-first
// order as selectRecordsPage, plus total matching count and has_more.
func selectRecordsPageByNumber(records []store.LogRecord, opts store.FilterOptions, pageNum, limit int) (page []store.LogRecord, total int, more bool) {
	return selectRecordsPageByNumberBlocks(records, nil, opts, pageNum, limit)
}

func selectRecordsPageByNumberBlocks(records []store.LogRecord, blocks []store.BlockMeta, opts store.FilterOptions, pageNum, limit int) (page []store.LogRecord, total int, more bool) {
	if len(records) == 0 || pageNum < 1 || limit < 1 {
		return []store.LogRecord{}, 0, false
	}
	start := (pageNum - 1) * limit
	k := start + limit + 1

	// Bounded top-K heap path for pages 1..~100 (at limit=20): O(N log k) time
	// and O(k) memory, plus O(1) full-block skipping for older blocks.
	if k <= maxBoundedPageHeap {
		h := recordPageHeap{records: records, indexes: make([]int, 0, min(k, len(records)))}
		timeOnly := !hasNonTimeFilters(&opts)

		scanRange := func(lo, hi int) {
			for i := hi - 1; i >= lo; i-- {
				if !store.MatchRecord(&records[i], &opts) {
					continue
				}
				total++
				if len(h.indexes) < k {
					h.push(i)
				} else if h.newer(i, h.indexes[0]) {
					h.replaceOldest(i)
				}
			}
		}

		if len(blocks) == 0 {
			scanRange(0, len(records))
		} else {
			n := len(records)
			for bi := len(blocks) - 1; bi >= 0; bi-- {
				blk := &blocks[bi]
				if blk.Start >= n {
					continue
				}
				blkEnd := blk.End
				if blkEnd > n {
					blkEnd = n
				}
				if blockOutsideRange(blk, &opts) {
					continue
				}
				if timeOnly && blkEnd == blk.End &&
					len(h.indexes) == k && blockFullyInsideRange(blk, &opts) &&
					blk.MaxTS.Before(records[h.indexes[0]].EffectiveTimestamp()) {
					total += blkEnd - blk.Start
					continue
				}
				scanRange(blk.Start, blkEnd)
			}
		}

		if total == 0 || start >= total {
			return []store.LogRecord{}, total, false
		}
		slices.SortFunc(h.indexes, h.compare)
		end := min(start+limit, len(h.indexes))
		more = start+limit < total
		page = make([]store.LogRecord, end-start)
		for i := start; i < end; i++ {
			page[i-start] = records[h.indexes[i]]
		}
		return page, total, more
	}

	// Deep-page path: collect matching indexes (skipping out-of-range blocks),
	// partition with 3-way Quickselect in O(M) average time, and sort only the
	// returned window [start:end].
	h := recordPageHeap{records: records}
	indexes := make([]int, 0, min(len(records), 256))
	if len(blocks) == 0 {
		for i := range records {
			if store.MatchRecord(&records[i], &opts) {
				indexes = append(indexes, i)
			}
		}
	} else {
		n := len(records)
		for bi := range blocks {
			blk := &blocks[bi]
			if blk.Start >= n {
				continue
			}
			blkEnd := blk.End
			if blkEnd > n {
				blkEnd = n
			}
			if blockOutsideRange(blk, &opts) {
				continue
			}
			for i := blk.Start; i < blkEnd; i++ {
				if store.MatchRecord(&records[i], &opts) {
					indexes = append(indexes, i)
				}
			}
		}
	}
	total = len(indexes)
	if total == 0 || start >= total {
		return []store.LogRecord{}, total, false
	}
	end := min(start+limit, total)
	nthElement(indexes, start, h.compare)
	if end < total {
		nthElement(indexes[start:], end-start, h.compare)
	}
	slices.SortFunc(indexes[start:end], h.compare)
	more = end < total
	page = make([]store.LogRecord, end-start)
	for i := start; i < end; i++ {
		page[i-start] = records[indexes[i]]
	}
	return page, total, more
}

// nthElement partitions a in-place so that a[k] is in its final sorted position,
// all elements before k are <= a[k], and all elements after k are >= a[k].
// Uses 3-way Dutch-National-Flag partitioning with median-of-three pivots so
// ties and already-sorted inputs stay O(n).
func nthElement(a []int, k int, cmp func(i, j int) int) {
	lo, hi := 0, len(a)-1
	for lo < hi {
		mid := lo + (hi-lo)/2
		if cmp(a[mid], a[lo]) < 0 {
			a[lo], a[mid] = a[mid], a[lo]
		}
		if cmp(a[hi], a[lo]) < 0 {
			a[lo], a[hi] = a[hi], a[lo]
		}
		if cmp(a[hi], a[mid]) < 0 {
			a[mid], a[hi] = a[hi], a[mid]
		}
		pivot := a[mid]
		lt, i, gt := lo, lo, hi
		for i <= gt {
			c := cmp(a[i], pivot)
			if c < 0 {
				a[lt], a[i] = a[i], a[lt]
				lt++
				i++
			} else if c > 0 {
				a[i], a[gt] = a[gt], a[i]
				gt--
			} else {
				i++
			}
		}
		if k < lt {
			hi = lt - 1
		} else if k > gt {
			lo = gt + 1
		} else {
			return
		}
	}
}
