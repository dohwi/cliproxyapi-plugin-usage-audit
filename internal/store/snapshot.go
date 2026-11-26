package store

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"
	"unsafe"
)

// maxSnapshotCacheBytes bounds the decoded history kept in memory (estimate is
// conservative: it counts every string field as if it were not interned). A
// history above this is re-decoded from disk on each request. Default 512 MiB
// accommodates the full default retention (5 × 64 MiB JSONL files) without
// hitting a cold-decode cliff on every request.
var maxSnapshotCacheBytes uint64 = 512 << 20

// SnapshotBlockSize is the number of records grouped into one zone-map block
// for O(1) time-range skipping in summary and pagination queries.
const SnapshotBlockSize = 1024

// BlockMeta records exact timestamp bounds for Snapshot.Records[Start:End].
type BlockMeta struct {
	Start int
	End   int
	MinTS time.Time
	MaxTS time.Time
}

var generationID atomic.Uint64

// Health contains only counters and a fixed error category, never filesystem
// paths, raw decoder input, credentials, or other error details.
type Health struct {
	WriteFailures    uint64    `json:"write_failures"`
	FlushFailures    uint64    `json:"flush_failures"`
	RotationFailures uint64    `json:"rotation_failures"`
	DecodeFailures   uint64    `json:"decode_failures"`
	CorruptLines     uint64    `json:"corrupt_lines"`
	TruncatedLines   uint64    `json:"truncated_lines"`
	ReadFailures     uint64    `json:"read_failures"`
	SnapshotDecodes  uint64    `json:"snapshot_decodes"`
	CacheHits        uint64    `json:"cache_hits"`
	DecodeMode       string    `json:"decode_mode"`
	Degraded         bool      `json:"degraded"`
	LastError        string    `json:"last_error,omitempty"`
	LastErrorAt      time.Time `json:"last_error_at,omitempty"`
}

func (s *Store) Health() Health {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.health
}
func (s *Store) failureLocked(category string) {
	s.health.LastError, s.health.LastErrorAt = category, time.Now().UTC()
}

// Snapshot owns its slice (or aliases a read-only view via SnapshotView).
// Sequence is stable across appends, including legacy JSONL with no sequence,
// but invalidated by retention, clear or external edits.
// Only one bounded decoded cache is retained; callers cannot mutate it.
type Snapshot struct {
	Records    []LogRecord
	Generation string
	Boundary   uint64
	Blocks     []BlockMeta
}
type fileStamp struct {
	path string
	info os.FileInfo
}
type cachedFile struct {
	info    os.FileInfo
	records []LogRecord
}

func (s *Store) newGenerationLocked() {
	s.generation = fmt.Sprintf("%d-%d", time.Now().UnixNano(), generationID.Add(1))
	s.cache, s.cacheBlocks, s.manifest, s.cacheValid, s.blocksShared, s.cacheBytes = nil, nil, nil, false, false, 0
	s.lastObserve = time.Time{}
}

// recordCost estimates the retained heap size of one decoded record.
func recordCost(r *LogRecord) uint64 {
	return uint64(unsafe.Sizeof(*r)) + uint64(len(r.Provider)+len(r.ExecutorType)+len(r.Model)+len(r.Alias)+len(r.SessionID)+len(r.ParentSessionID)+len(r.AuthIndex)+len(r.AuthType)+len(r.Source)+len(r.ReasoningEffort)+len(r.ServiceTier)+len(r.EffectiveServiceTier)+len(r.ResponseServiceTier))
}

func buildBlocks(records []LogRecord) []BlockMeta {
	if len(records) == 0 {
		return nil
	}
	numBlocks := (len(records) + SnapshotBlockSize - 1) / SnapshotBlockSize
	blocks := make([]BlockMeta, 0, numBlocks)
	for start := 0; start < len(records); start += SnapshotBlockSize {
		end := start + SnapshotBlockSize
		if end > len(records) {
			end = len(records)
		}
		minTS := records[start].EffectiveTimestamp()
		maxTS := minTS
		for i := start + 1; i < end; i++ {
			ts := records[i].EffectiveTimestamp()
			if ts.Before(minTS) {
				minTS = ts
			}
			if ts.After(maxTS) {
				maxTS = ts
			}
		}
		blocks = append(blocks, BlockMeta{Start: start, End: end, MinTS: minTS, MaxTS: maxTS})
	}
	return blocks
}

func appendBlock(blocks []BlockMeta, idx int, ts time.Time, shared bool) ([]BlockMeta, bool) {
	if shared && len(blocks) > 0 {
		cp := make([]BlockMeta, len(blocks), len(blocks)+1)
		copy(cp, blocks)
		blocks = cp
		shared = false
	}
	if len(blocks) == 0 || blocks[len(blocks)-1].End-blocks[len(blocks)-1].Start >= SnapshotBlockSize {
		blocks = append(blocks, BlockMeta{Start: idx, End: idx + 1, MinTS: ts, MaxTS: ts})
		return blocks, shared
	}
	last := &blocks[len(blocks)-1]
	last.End = idx + 1
	if ts.Before(last.MinTS) {
		last.MinTS = ts
	}
	if ts.After(last.MaxTS) {
		last.MaxTS = ts
	}
	return blocks, shared
}
func (s *Store) statManifestLocked() ([]fileStamp, error) {
	paths, err := logPaths(s.path)
	if err != nil {
		return nil, err
	}
	stamps := make([]fileStamp, 0, len(paths))
	for _, p := range paths {
		info, err := os.Stat(p)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		stamps = append(stamps, fileStamp{p, info})
	}
	return stamps, nil
}
func sameManifest(a, b []fileStamp) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].path != b[i].path || !os.SameFile(a[i].info, b[i].info) || a[i].info.Size() != b[i].info.Size() || !a[i].info.ModTime().Equal(b[i].info.ModTime()) {
			return false
		}
	}
	return true
}
func (s *Store) observeManifestLocked() error {
	m, err := s.statManifestLocked()
	if err != nil {
		return err
	}
	if s.manifest != nil && !sameManifest(s.manifest, m) {
		s.newGenerationLocked()
	}
	s.manifest = m
	return nil
}
func (s *Store) refreshManifestLocked() error {
	m, err := s.statManifestLocked()
	if err == nil {
		s.manifest = m
	}
	return err
}

// Snapshot returns a snapshot the caller owns and may mutate.
func (s *Store) Snapshot() (Snapshot, error) { return s.snapshot(true) }

// SnapshotView returns a read-only snapshot that may alias the internal cache,
// avoiding an O(n) copy per request. Callers MUST NOT modify Records. This is
// safe against concurrent Append because elements below Boundary are never
// rewritten: appends only write past the end or into a fresh backing array.
func (s *Store) SnapshotView() (Snapshot, error) { return s.snapshot(false) }

func (s *Store) snapshot(own bool) (snap Snapshot, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	previousErrorAt := s.health.LastErrorAt
	defer func() {
		if err != nil {
			s.health.ReadFailures++
			// Preserve the more specific flush/decode failure when available.
			if s.health.LastErrorAt.Equal(previousErrorAt) {
				s.failureLocked("read")
			}
		}
	}()
	if s.closed {
		return snap, os.ErrClosed
	}
	if err = s.observeManifestLocked(); err != nil {
		return snap, err
	}
	if s.bw != nil && s.bw.Buffered() > 0 {
		if err = s.flushLocked(); err != nil {
			return snap, err
		}
		if err = s.refreshManifestLocked(); err != nil {
			return snap, err
		}
	}
	if s.cacheValid {
		s.health.CacheHits++
		if !own {
			n := len(s.cache)
			nb := len(s.cacheBlocks)
			s.blocksShared = true
			return Snapshot{s.cache[:n:n], s.generation, uint64(n), s.cacheBlocks[:nb:nb]}, nil
		}
		return Snapshot{
			Records:    append([]LogRecord(nil), s.cache...),
			Generation: s.generation,
			Boundary:   uint64(len(s.cache)),
			Blocks:     append([]BlockMeta(nil), s.cacheBlocks...),
		}, nil
	}
	s.health.SnapshotDecodes++
	records, nextCachedFiles, report, err := s.readAllCachedLocked()
	if report.corruptLines > 0 {
		s.health.DecodeFailures++
		s.health.CorruptLines += report.corruptLines
		s.health.TruncatedLines += report.truncatedLines
		s.health.Degraded = true
		s.failureLocked("corrupt_jsonl")
	} else if err == nil && s.health.Degraded && s.health.LastError == "corrupt_jsonl" {
		// A complete decode that found no corruption means the damaged data was
		// rotated away, cleared or repaired. Counters stay as history, but the
		// store is no longer degraded and the stale error category is dropped.
		s.health.Degraded = false
		s.health.LastError = ""
		s.health.LastErrorAt = time.Time{}
	}
	if err != nil {
		return snap, err
	}
	// Detect out-of-process changes during decoding rather than caching a mixed view.
	after, err := s.statManifestLocked()
	if err != nil {
		return snap, err
	}
	if !sameManifest(s.manifest, after) {
		s.newGenerationLocked()
		return snap, errors.New("store: files changed during snapshot; retry")
	}
	var size uint64
	for i := range records {
		r := &records[i]
		r.Sequence = uint64(i + 1)
		// Legacy API key and source fields must not enter either API responses or cache.
		r.APIKeyID, r.Source = "", ""
		size += recordCost(r)
	}
	blocks := buildBlocks(records)
	if size <= maxSnapshotCacheBytes {
		s.cacheBytes = size
		s.cacheValid = true
		s.cacheBlocks = blocks
		s.cachedFiles = nextCachedFiles
		if own {
			// The caller may mutate its copy, so the cache needs its own backing array.
			s.cache = make([]LogRecord, len(records), len(records)+min(len(records)/8+256, 8192))
			copy(s.cache, records)
			s.blocksShared = false
			return Snapshot{records, s.generation, uint64(len(records)), append([]BlockMeta(nil), blocks...)}, nil
		}
		// Keep extra capacity on s.cache so subsequent Append calls do not reallocate
		// and copy the entire history, while clamping the returned view to [:n:n] so
		// callers can only read [0:n) and never alias appended tail elements.
		n := len(records)
		if cap(records) == n && n > 0 {
			grown := make([]LogRecord, n, n+min(n/8+256, 8192))
			copy(grown, records)
			records = grown
		}
		s.cache = records
		s.blocksShared = true
		nb := len(blocks)
		return Snapshot{s.cache[:n:n], s.generation, uint64(n), s.cacheBlocks[:nb:nb]}, nil
	}
	s.cachedFiles = nil
	return Snapshot{records, s.generation, uint64(len(records)), blocks}, nil
}
