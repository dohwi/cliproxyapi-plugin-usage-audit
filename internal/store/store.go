// Package store implements append-only JSONL persistence for usage records.
// Each line is a self-contained JSON object. No external dependencies.
//
// Safety contracts:
//   - Write/flush/rotation errors are returned to the caller for logging.
//   - File is opened with O_APPEND|O_CREATE|O_WRONLY.
//   - Rotation: when file exceeds maxBytes, it is renamed to .1 and a new file opened.
//   - Writes, snapshot reads, deletion, and close are serialized via mu.
package store

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	DefaultMaxBytes    = 64 * 1024 * 1024 // 64 MiB per file
	DefaultMaxRotation = 4                // keep .1 .. .4
	defaultFlushEvery  = 256              // lines between explicit flushes
)

// observeInterval bounds how often Append re-checks the retained files for
// out-of-band changes. Snapshot reads are never throttled.
var observeInterval = time.Second

// ErrCorruptJSONL is deliberately detail-free so strict-mode callers cannot
// accidentally expose a file name, line number, raw record, or secret.
var ErrCorruptJSONL = errors.New("store: corrupt JSONL record")

// Options controls administrative Store behavior. Resilient decoding is the
// operational default: malformed non-empty lines are skipped and reported via
// Health while later valid records remain queryable.
type Options struct {
	StrictDecoding bool
}

// LogRecord is the on-disk format. Never stores raw tokens or auth secrets.
type LogRecord struct {
	// Sequence is the append ordinal within a snapshot generation; never persisted.
	Sequence uint64 `json:"-"`
	// T is the UTC timestamp the record was ingested by the plugin.
	T time.Time `json:"t"`
	// Provider is the upstream provider key (e.g. "antigravity", "codex").
	Provider string `json:"provider,omitempty"`
	// ExecutorType is the concrete executor that handled the request.
	ExecutorType string `json:"executor_type,omitempty"`
	// Model is the upstream model name actually used.
	Model string `json:"model,omitempty"`
	// Alias is the user-facing model alias (may be empty).
	Alias string `json:"alias,omitempty"`
	// APIKeyID is retained for legacy decoding only. Append always clears it.
	APIKeyID string `json:"api_key_id,omitempty"`
	// SessionID identifies the client session.
	SessionID string `json:"session_id,omitempty"`
	// ParentSessionID identifies the parent session in a hierarchy.
	ParentSessionID string `json:"parent_session_id,omitempty"`
	// AuthIndex is the stable runtime credential index. Never a raw token.
	AuthIndex string `json:"auth_index,omitempty"`
	// AuthType is the credential type (e.g. "oauth").
	AuthType string `json:"auth_type,omitempty"`
	// Source identifies the request source/integration.
	Source string `json:"source,omitempty"`
	// ReasoningEffort is the requested reasoning effort level.
	ReasoningEffort string `json:"reasoning_effort,omitempty"`
	// ServiceTier is the requested or reported service tier.
	ServiceTier string `json:"service_tier,omitempty"`
	// Generate distinguishes actual generation (true) from token-count-only (false).
	// Stored explicitly: false ≠ missing.
	Generate bool `json:"generate"`
	// RequestedAt is when the host received the client request.
	RequestedAt time.Time `json:"requested_at,omitempty"`
	// LatencyMs is the total request latency in milliseconds.
	LatencyMs float64 `json:"latency_ms,omitempty"`
	// TTFTMs is time-to-first-token for streaming requests, in milliseconds.
	TTFTMs float64 `json:"ttft_ms,omitempty"`
	// Failed reports whether the upstream request failed.
	Failed bool `json:"failed"`
	// FailureStatus is the HTTP status code on failure (0 if not failed).
	FailureStatus int `json:"failure_status,omitempty"`
	// InputTokens is the prompt/input token count. 0 is valid (not missing).
	InputTokens int64 `json:"input_tokens"`
	// OutputTokens is the completion/output token count.
	OutputTokens int64 `json:"output_tokens"`
	// ReasoningTokens is the reasoning token count.
	ReasoningTokens int64 `json:"reasoning_tokens"`
	// CachedTokens is the total cached token count.
	CachedTokens int64 `json:"cached_tokens"`
	// CacheReadTokens is the cache read token count.
	CacheReadTokens int64 `json:"cache_read_tokens"`
	// CacheCreationTokens is the cache creation token count.
	CacheCreationTokens int64 `json:"cache_creation_tokens"`
	// TotalTokens is the total token count.
	TotalTokens int64 `json:"total_tokens"`
}

// Store serializes records to an append-only rotated JSONL file.
type Store struct {
	mu           sync.Mutex
	path         string
	maxBytes     int64
	maxRotate    int
	f            *os.File
	bw           *bufio.Writer
	written      int64 // bytes accepted for current file, including buffered writes
	needsNewline bool  // pre-existing active file has an unterminated final line
	closed       bool
	lineCount    int
	generation   string
	manifest     []fileStamp
	lastObserve  time.Time // last full manifest observation made by Append (throttle)
	cache        []LogRecord
	cacheValid   bool
	cacheBytes   uint64 // estimated retained size of cache; maintained incrementally by Append
	health       Health
	strictDecode bool
	// Per-store hooks permit deterministic filesystem fault injection.
	rename     func(string, string) error
	remove     func(string) error
	createTemp func(string, string) (*os.File, error)
}

// Open opens or creates the JSONL log file at path using resilient decoding.
func Open(path string, maxBytes int64, maxRotate int) (*Store, error) {
	return OpenWithOptions(path, maxBytes, maxRotate, Options{})
}

// OpenWithOptions opens or creates the JSONL log file with administrative
// decoding options. Strict decoding is opt-in and rejects any malformed line,
// including an unterminated truncated tail.
func OpenWithOptions(path string, maxBytes int64, maxRotate int, opts Options) (*Store, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	if maxRotate <= 0 {
		maxRotate = DefaultMaxRotation
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("store: mkdir %s: %w", filepath.Dir(path), err)
	}
	mode := "resilient"
	if opts.StrictDecoding {
		mode = "strict"
	}
	s := &Store{
		path: path, maxBytes: maxBytes, maxRotate: maxRotate,
		strictDecode: opts.StrictDecoding,
		health:       Health{DecodeMode: mode},
		rename:       os.Rename, remove: os.Remove, createTemp: os.CreateTemp,
	}
	s.newGenerationLocked()
	if err := s.openFile(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) openFile() error {
	f, err := openAppend(s.path, 0o640, true)
	if err != nil {
		return fmt.Errorf("store: open %s: %w", s.path, err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("store: stat %s: %w", s.path, err)
	}
	s.f = f
	s.bw = bufio.NewWriterSize(f, 64*1024)
	s.written = info.Size()
	s.needsNewline = false
	if info.Size() > 0 {
		last, readErr := lastFileByte(s.path, info.Size())
		if readErr != nil {
			_ = f.Close()
			s.f, s.bw = nil, nil
			return fmt.Errorf("store: inspect active tail: %w", readErr)
		}
		s.needsNewline = last != '\n'
	}
	s.lineCount = 0
	return nil
}

func lastFileByte(path string, size int64) (byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var b [1]byte
	_, err = f.ReadAt(b[:], size-1)
	return b[0], err
}

// Append serializes rec to a JSONL line and appends it.
// Errors are returned but callers (the queue consumer) should log-and-continue.
func (s *Store) Append(rec LogRecord) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	previousErrorAt := s.health.LastErrorAt
	defer func() {
		if err != nil {
			s.health.WriteFailures++
			if s.health.LastErrorAt.Equal(previousErrorAt) {
				s.failureLocked("write")
			}
		}
	}()
	if s.closed {
		return os.ErrClosed
	}
	// External-change detection on the write path is throttled: a full
	// ReadDir+Stat of every retained file per record dominated Append CPU. Reads
	// (snapshot) always run the full check, so a stale view is never served; the
	// only behavioural difference is that an out-of-band rotation/rename/edit may
	// go unnoticed by Append for up to observeInterval (the record is still
	// appended through the already-open descriptor).
	if s.manifest == nil || s.lastObserve.IsZero() || time.Since(s.lastObserve) >= observeInterval {
		if err := s.observeManifestLocked(); err != nil {
			return err
		}
		s.lastObserve = time.Now()
	}
	// Keep the decoded cache across a successful append instead of discarding it:
	// dropping it forced every dashboard load after any new record to re-decode
	// the entire retained history. Any failure path below leaves it invalid.
	keptCache, keptBytes, keptGeneration := s.cache, s.cacheBytes, s.generation
	extendCache := s.cacheValid
	s.cache, s.cacheValid, s.cacheBytes = nil, false, 0
	if s.f == nil {
		if err := s.openFile(); err != nil {
			return err
		}
	}
	// The host APIKey field may contain a secret, not a safe identifier.
	rec.APIKeyID = ""
	line, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("store: marshal: %w", err)
	}
	line = append(line, '\n')

	separatorBytes := int64(0)
	if s.needsNewline {
		separatorBytes = 1
	}
	if s.written > 0 && s.written+separatorBytes+int64(len(line)) > s.maxBytes {
		if err := s.rotateLocked(); err != nil {
			return err
		}
	}
	// Preserve an existing valid or malformed unterminated tail as its own
	// JSONL line. Without this separator, the first live append would be joined
	// to the tail and become unreadable too.
	if s.needsNewline {
		n, err := s.bw.Write([]byte{'\n'})
		s.written += int64(n)
		if err != nil {
			return fmt.Errorf("store: write separator: %w", err)
		}
		s.needsNewline = false
	}

	// refreshManifestLocked adopts whatever is on disk as "our" state, which would
	// silently swallow an out-of-band change made inside the throttle window. So
	// whenever this write is about to touch the file (buffer spill), do the full
	// observation first; pure buffered appends never reach the file.
	spills := len(line) > s.bw.Available()
	if spills {
		if err := s.observeManifestLocked(); err != nil {
			return err
		}
		s.lastObserve = time.Now()
		if s.generation != keptGeneration {
			extendCache = false
		}
	}
	buffered := s.bw.Buffered()
	n, err := s.bw.Write(line)
	s.written += int64(n)
	// Buffered bytes do not touch the file, so the recorded size/mtime stay valid.
	// Only when the bufio layer spilled to the file (buffer filled) did our own
	// write change the file stamp; refresh just then instead of on every record.
	if s.bw.Buffered() < buffered+len(line) || err != nil {
		if e := s.refreshManifestLocked(); e != nil {
			err = errors.Join(err, e)
		}
	}
	if err != nil {
		return fmt.Errorf("store: write: %w", err)
	}
	// Rotation (generation change) or any decode mismatch leaves the cache
	// invalid so the next Snapshot decodes from disk. The appended line is
	// round-tripped through the decoder so the cached record is identical to
	// what a fresh decode would yield.
	if extendCache && s.generation == keptGeneration {
		cached, ok := normalizeLikeDecode(rec)
		if !ok {
			// Rare shapes (non-UTC offsets) keep the exact decoder round trip.
			ok = json.Unmarshal(bytes.TrimSpace(line), &cached) == nil
		}
		if ok {
			cached.Sequence = uint64(len(keptCache) + 1)
			cached.APIKeyID = ""
			if total := keptBytes + recordCost(&cached); total <= maxSnapshotCacheBytes {
				s.cache, s.cacheBytes, s.cacheValid = append(keptCache, cached), total, true
			}
		}
	}
	s.lineCount++
	if s.lineCount%defaultFlushEvery == 0 {
		// Same reasoning as the spill path: look before our own flush changes the file.
		if err := s.observeManifestLocked(); err != nil {
			return err
		}
		s.lastObserve = time.Now()
		return errors.Join(s.flushLocked(), s.refreshManifestLocked())
	}
	return nil
}

func (s *Store) flushLocked() error {
	if s.bw != nil {
		if err := s.bw.Flush(); err != nil {
			s.health.FlushFailures++
			s.failureLocked("flush")
			return fmt.Errorf("store: flush: %w", err)
		}
	}
	return nil
}

// Flush flushes buffered writes to the OS (not fsync).
func (s *Store) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.observeManifestLocked(); err != nil {
		return err
	}
	return errors.Join(s.flushLocked(), s.refreshManifestLocked())
}

// Close flushes and closes the underlying file, returning both errors if needed.
// It is idempotent; subsequent Append calls return os.ErrClosed.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	err := s.flushLocked()
	if s.f != nil {
		err = errors.Join(err, s.f.Close())
	}
	s.f, s.bw = nil, nil
	return err
}

// ReadAll flushes and snapshots the complete retained history under the writer
// lock, preventing rename/delete races across individual files.
func (s *Store) ReadAll() ([]LogRecord, error) {
	snap, err := s.Snapshot()
	return snap.Records, err
}

// ReadAllView is ReadAll without the defensive copy. The result is read-only.
func (s *Store) ReadAllView() ([]LogRecord, error) {
	snap, err := s.SnapshotView()
	return snap.Records, err
}

// Clear removes all retained logs and creates a fresh active writer under the
// same lock as Append. Buffered records are deliberately discarded. Records
// appended after this operation are new writes. The plugin epoch barrier rejects
// pending upstream queue entries from before the clear. Filesystem failure may leave a partially cleared history and is
// reported; reopening is attempted even when removal fails.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return os.ErrClosed
	}
	s.newGenerationLocked()
	paths, err := logPaths(s.path)
	if err != nil {
		return err
	}
	if s.f != nil {
		err = s.f.Close()
	}
	s.f, s.bw = nil, nil
	for _, path := range paths {
		if e := os.Remove(path); e != nil && !os.IsNotExist(e) {
			err = errors.Join(err, fmt.Errorf("store: remove %s: %w", path, e))
		}
	}
	return errors.Join(err, s.openFile())
}

// logPaths returns numeric rotations oldest-first, then the active file. Scan
// all canonical positive suffixes, including rotations from an older config.
func logPaths(path string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Dir(path))
	if os.IsNotExist(err) {
		return []string{path}, nil
	}
	if err != nil {
		return nil, err
	}
	prefix := filepath.Base(path) + "."
	var rotations []int
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		suffix := strings.TrimPrefix(entry.Name(), prefix)
		n, err := strconv.Atoi(suffix)
		if err == nil && n > 0 && strconv.Itoa(n) == suffix {
			rotations = append(rotations, n)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(rotations)))
	paths := make([]string, 0, len(rotations)+1)
	for _, n := range rotations {
		paths = append(paths, fmt.Sprintf("%s.%d", path, n))
	}
	return append(paths, path), nil
}

// ReadAll reads active and rotated JSONL logs in append order, oldest file first.
// Call Store.ReadAll instead when a live writer exists; this path-only helper
// cannot synchronize with a Store. The default is resilient: malformed
// non-empty lines, including an unterminated truncated final object, are skipped
// while later valid records are returned. Use ReadAllStrict for validation jobs.
func ReadAll(path string) ([]LogRecord, error) {
	records, _, err := readAll(path, false)
	return records, err
}

// ReadAllStrict rejects the first malformed non-empty line. Its corruption
// error is intentionally detail-free; it never includes file/line/raw content.
func ReadAllStrict(path string) ([]LogRecord, error) {
	records, _, err := readAll(path, true)
	return records, err
}

type decodeReport struct {
	corruptLines   uint64
	truncatedLines uint64
}

func (r *decodeReport) add(other decodeReport) {
	r.corruptLines += other.corruptLines
	r.truncatedLines += other.truncatedLines
}

func readAll(path string, strict bool) ([]LogRecord, decodeReport, error) {
	paths, err := logPaths(path)
	if err != nil {
		return nil, decodeReport{}, err
	}
	var out []LogRecord
	var report decodeReport
	for _, name := range paths {
		recs, fileReport, err := readFile(name, strict)
		report.add(fileReport)
		if err != nil {
			return nil, report, err
		}
		out = append(out, recs...)
	}
	internRecordStrings(out)
	return out, report, nil
}

// internRecordStrings makes repeated low-cardinality strings (provider, model,
// auth index, session id, ...) share one backing allocation. A decoded history
// otherwise holds one copy per record per field, which dominates cache size
// and GC scan time.
func internRecordStrings(recs []LogRecord) {
	seen := make(map[string]string, 256)
	in := func(p *string) {
		if *p == "" {
			return
		}
		if v, ok := seen[*p]; ok {
			*p = v
			return
		}
		seen[*p] = *p
	}
	for i := range recs {
		r := &recs[i]
		in(&r.Provider)
		in(&r.ExecutorType)
		in(&r.Model)
		in(&r.Alias)
		in(&r.SessionID)
		in(&r.ParentSessionID)
		in(&r.AuthIndex)
		in(&r.AuthType)
		in(&r.Source)
		in(&r.ReasoningEffort)
		in(&r.ServiceTier)
	}
}

func readFile(path string, strict bool) ([]LogRecord, decodeReport, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, decodeReport{}, nil
	}
	if err != nil {
		return nil, decodeReport{}, err
	}
	recs, report, decodeErr := decode(f, strict)
	return recs, report, errors.Join(decodeErr, f.Close())
}

// ReadTail reads the last n records from a file path.
func ReadTail(path string, n int) ([]LogRecord, error) {
	if n <= 0 {
		return []LogRecord{}, nil
	}
	all, err := ReadAll(path)
	if err != nil || len(all) <= n {
		return all, err
	}
	return all[len(all)-n:], nil
}

// decode parses JSONL records from r according to the decoding policy.
// In resilient mode (the operational default):
//   - Malformed, non-object, or multi-object lines are skipped on a per-line basis;
//     subsequent valid records continue to be decoded across active and rotated files.
//   - Final truncated line policy:
//     1. A valid JSON object without a trailing newline at EOF is accepted as a valid record
//     (neither corrupt_lines nor truncated_lines is incremented).
//     2. An unterminated truncated JSON object at EOF (starts with '{' but interrupted before
//     syntactic completion, returning unexpected EOF) is skipped and counted in both
//     corrupt_lines and truncated_lines.
//     3. A newline-terminated malformed line anywhere in the file is classified as ordinary
//     corruption and counted in corrupt_lines only, not truncated_lines.
//     4. Non-object or plain corrupt lines at EOF without a newline are counted in corrupt_lines only.
//   - In strict mode, the first malformed or non-object line immediately aborts decoding
//     with ErrCorruptJSONL without exposing corrupt content, file paths, or line numbers.
func decode(r io.Reader, strict bool) ([]LogRecord, decodeReport, error) {
	var out []LogRecord
	var report decodeReport
	reader := bufio.NewReader(r)
	for {
		line, readErr := reader.ReadBytes('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return nil, report, readErr
		}
		data := bytes.TrimSpace(line)
		if len(data) > 0 {
			var rec LogRecord
			unmarshalErr := json.Unmarshal(data, &rec)
			validObject := unmarshalErr == nil && len(data) >= 2 && data[0] == '{' && data[len(data)-1] == '}'
			if !validObject {
				report.corruptLines++
				if errors.Is(readErr, io.EOF) && isTruncatedJSONObject(data) {
					report.truncatedLines++
				}
				if strict {
					return nil, report, ErrCorruptJSONL
				}
			} else {
				out = append(out, rec)
			}
		}
		if errors.Is(readErr, io.EOF) {
			return out, report, nil
		}
	}
}

// isTruncatedJSONObject recognizes an unterminated final JSON object at EOF. A
// newline-terminated malformed line is ordinary corruption, and syntactically
// complete type errors or non-object payloads are not classified as truncation.
func isTruncatedJSONObject(data []byte) bool {
	if len(data) == 0 || data[0] != '{' {
		return false
	}
	var raw json.RawMessage
	err := json.NewDecoder(bytes.NewReader(data)).Decode(&raw)
	return errors.Is(err, io.ErrUnexpectedEOF)
}

// EffectiveTimestamp is the request time, falling back to ingestion time for
// legacy records and hosts that omit RequestedAt.
func (r LogRecord) EffectiveTimestamp() time.Time {
	if !r.RequestedAt.IsZero() {
		return r.RequestedAt
	}
	return r.T
}

// FilterOptions specifies query filters for LogRecords.
type FilterOptions struct {
	From         *time.Time
	To           *time.Time
	Provider     string
	Model        string
	Status       string // "success" or "failed"
	Source       string
	ExecutorType string
	// Endpoint matches either ExecutorType OR Source (OR semantics), mirroring
	// the by_endpoint aggregation key derivation in the summary handler.
	Endpoint  string
	AuthIndex string
}

// FilterRecords filters records and returns matching records in newest-first (descending) order.
// Newest-first order is stable: records with identical timestamps preserve reverse-ingest order.
func FilterRecords(records []LogRecord, opts FilterOptions) []LogRecord {
	// Do not pre-size to len(records): LogRecord is ~400B and typical filters
	// (today, one provider) keep a small fraction of the history.
	var filtered []LogRecord
	for i := range records {
		if MatchRecord(&records[i], &opts) {
			filtered = append(filtered, records[i])
		}
	}

	// Reverse to newest-first (from append order)
	n := len(filtered)
	for i := 0; i < n/2; i++ {
		filtered[i], filtered[n-1-i] = filtered[n-1-i], filtered[i]
	}
	newer := func(i, j int) bool {
		return filtered[i].EffectiveTimestamp().After(filtered[j].EffectiveTimestamp())
	}
	// Append order is nearly chronological, so the reversed slice is usually
	// already sorted; skip the O(n log n) stable sort in that case.
	if !sort.SliceIsSorted(filtered, newer) {
		sort.SliceStable(filtered, newer)
	}
	return filtered
}

// MatchRecord reports whether r satisfies opts. It takes pointers so hot loops
// can scan a snapshot without copying each ~400-byte record.
func MatchRecord(r *LogRecord, opts *FilterOptions) bool {
	if opts.From != nil || opts.To != nil {
		ts := r.EffectiveTimestamp()
		if opts.From != nil && ts.Before(*opts.From) {
			return false
		}
		if opts.To != nil && ts.After(*opts.To) {
			return false
		}
	}
	if opts.Provider != "" && r.Provider != opts.Provider {
		return false
	}
	if opts.Model != "" && r.Model != opts.Model {
		return false
	}
	if opts.Source != "" && r.Source != opts.Source {
		return false
	}
	if opts.ExecutorType != "" && r.ExecutorType != opts.ExecutorType {
		return false
	}
	if opts.Endpoint != "" && r.ExecutorType != opts.Endpoint && r.Source != opts.Endpoint {
		return false
	}
	if opts.AuthIndex != "" && r.AuthIndex != opts.AuthIndex {
		return false
	}
	if opts.Status == "success" && r.Failed {
		return false
	}
	if opts.Status == "failed" && !r.Failed {
		return false
	}
	return true
}

// normalizeLikeDecode returns the record exactly as decode() would yield it
// after json.Marshal + json.Unmarshal of rec, without the encode/decode work.
// It reports false when equality cannot be guaranteed cheaply (a time carrying a
// non-UTC location), in which case the caller falls back to the real round trip.
// Differences it reproduces: monotonic clock readings are dropped, UTC times lose
// their *Location pointer, invalid UTF-8 bytes become U+FFFD one by one, and
// negative zero (omitted by omitempty) decodes as +0.
func normalizeLikeDecode(r LogRecord) (LogRecord, bool) {
	if r.T.Location() != time.UTC || r.RequestedAt.Location() != time.UTC {
		return r, false
	}
	r.T, r.RequestedAt = r.T.Round(0).UTC(), r.RequestedAt.Round(0).UTC()
	for _, p := range []*string{&r.Provider, &r.ExecutorType, &r.Model, &r.Alias, &r.APIKeyID, &r.SessionID,
		&r.ParentSessionID, &r.AuthIndex, &r.AuthType, &r.Source, &r.ReasoningEffort, &r.ServiceTier} {
		*p = validUTF8Like(*p)
	}
	if r.LatencyMs == 0 {
		r.LatencyMs = 0
	}
	if r.TTFTMs == 0 {
		r.TTFTMs = 0
	}
	return r, true
}

// validUTF8Like mirrors encoding/json, which writes U+FFFD for every single
// invalid byte (strings.ToValidUTF8 would collapse runs, so it is not used).
func validUTF8Like(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			b.WriteString("\uFFFD")
		} else {
			b.WriteString(s[i : i+size])
		}
		i += size
	}
	return b.String()
}
