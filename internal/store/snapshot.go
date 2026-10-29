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
// history above this is re-decoded from disk on each request.
var maxSnapshotCacheBytes uint64 = 128 << 20

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

// Snapshot owns its slice. Sequence is stable across appends, including legacy
// JSONL with no sequence, but invalidated by retention, clear or external edits.
// Only one bounded decoded cache is retained; callers cannot mutate it.
type Snapshot struct {
	Records    []LogRecord
	Generation string
	Boundary   uint64
}
type fileStamp struct {
	path string
	info os.FileInfo
}

func (s *Store) newGenerationLocked() {
	s.generation = fmt.Sprintf("%d-%d", time.Now().UnixNano(), generationID.Add(1))
	s.cache, s.manifest, s.cacheValid, s.cacheBytes = nil, nil, false, 0
	s.lastObserve = time.Time{}
}

// recordCost estimates the retained heap size of one decoded record.
func recordCost(r *LogRecord) uint64 {
	return uint64(unsafe.Sizeof(*r)) + uint64(len(r.Provider)+len(r.ExecutorType)+len(r.Model)+len(r.Alias)+len(r.SessionID)+len(r.ParentSessionID)+len(r.AuthIndex)+len(r.AuthType)+len(r.Source)+len(r.ReasoningEffort)+len(r.ServiceTier))
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
	if err = s.flushLocked(); err != nil {
		return snap, err
	}
	if err = s.refreshManifestLocked(); err != nil {
		return snap, err
	}
	if s.cacheValid {
		s.health.CacheHits++
		if !own {
			n := len(s.cache)
			return Snapshot{s.cache[:n:n], s.generation, uint64(n)}, nil
		}
		return Snapshot{append([]LogRecord(nil), s.cache...), s.generation, uint64(len(s.cache))}, nil
	}
	s.health.SnapshotDecodes++
	records, report, err := readAll(s.path, s.strictDecode)
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
		// Legacy API key fields must not enter either API responses or cache.
		r.APIKeyID = ""
		size += recordCost(r)
	}
	if size <= maxSnapshotCacheBytes {
		s.cacheBytes = size
		s.cacheValid = true
		if own {
			// The caller may mutate its copy, so the cache needs its own backing array.
			s.cache = make([]LogRecord, len(records))
			copy(s.cache, records)
		} else {
			// Read-only callers share the freshly decoded slice with the cache:
			// no second O(n) allocation on a cold build. Capacity is clamped so a
			// later Append can never write into the array a view may still hold.
			n := len(records)
			s.cache = records[:n:n]
			return Snapshot{s.cache, s.generation, uint64(n)}, nil
		}
	}
	return Snapshot{records, s.generation, uint64(len(records))}, nil
}
