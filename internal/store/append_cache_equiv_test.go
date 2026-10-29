package store

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// roundTrip is the reference: exactly what a disk decode of Append's line yields.
func roundTrip(t *testing.T, rec LogRecord) LogRecord {
	t.Helper()
	rec.APIKeyID = ""
	line, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	var out LogRecord
	if err := json.Unmarshal(line, &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func randString(r *rand.Rand) string {
	pool := []string{"", "codex", "한글 모델", "a\"b\\c", "line\nbreak\ttab", "<html>&amp;", "\u2028\u2029", "emoji😀", "nul\x00byte",
		"bad\xffutf8", "\xc3\x28", "\xe2\x82", "run\xff\xfe\xfdend", strings.Repeat("x", 300)}
	return pool[r.Intn(len(pool))]
}

func randTime(r *rand.Rand) time.Time {
	switch r.Intn(6) {
	case 0:
		return time.Time{}
	case 1:
		return time.Now() // carries a monotonic reading and Local
	case 2:
		return time.Now().UTC()
	case 3:
		return time.Unix(r.Int63n(4e9), r.Int63n(1e9)).UTC()
	case 4:
		return time.Date(2026, 3, 8, 2, 30, 0, r.Intn(1e9), time.FixedZone("KST", 9*3600))
	default:
		loc, err := time.LoadLocation("America/New_York")
		if err != nil {
			return time.Unix(0, 0).UTC()
		}
		return time.Unix(r.Int63n(4e9), r.Int63n(1e9)).In(loc)
	}
}

// Property: the record Append puts in the cache is identical (==) to the one a
// fresh decode of the same line yields, for UTC records; non-UTC records take
// the exact decoder fallback and compare equal as instants and fields.
func TestNormalizeLikeDecodeMatchesRoundTrip(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	fast := 0
	for i := 0; i < 20000; i++ {
		rec := LogRecord{
			T: randTime(r), RequestedAt: randTime(r),
			Provider: randString(r), ExecutorType: randString(r), Model: randString(r), Alias: randString(r),
			APIKeyID: "secret", SessionID: randString(r), ParentSessionID: randString(r), AuthIndex: randString(r),
			AuthType: randString(r), Source: randString(r), ReasoningEffort: randString(r), ServiceTier: randString(r),
			Generate: r.Intn(2) == 0, Failed: r.Intn(2) == 0, FailureStatus: r.Intn(600),
			LatencyMs:   []float64{0, math.Copysign(0, -1), 1.5, 123456.789012345, 1e-9, 1e21}[r.Intn(6)],
			TTFTMs:      []float64{0, math.Copysign(0, -1), 0.1, 3e7}[r.Intn(4)],
			InputTokens: r.Int63n(1 << 40), OutputTokens: -r.Int63n(5), ReasoningTokens: 0,
			CachedTokens: r.Int63(), CacheReadTokens: 1, CacheCreationTokens: 2, TotalTokens: math.MaxInt64,
		}
		want := roundTrip(t, rec)
		got, ok := normalizeLikeDecode(rec)
		got.APIKeyID = ""
		if !ok {
			if rec.T.Location() == time.UTC && rec.RequestedAt.Location() == time.UTC {
				t.Fatalf("UTC record must take the fast path")
			}
			continue
		}
		fast++
		if got != want {
			t.Fatalf("iteration %d mismatch\n got: %+v\nwant: %+v", i, got, want)
		}
		for _, s := range []string{got.Provider, got.Model, got.Alias, got.SessionID} {
			if !utf8.ValidString(s) {
				t.Fatal("invalid UTF-8 survived normalization")
			}
		}
	}
	if fast < 1000 {
		t.Fatalf("fast path barely exercised: %d", fast)
	}
}

// End to end: a cache extended by Append equals a cold decode of the same file,
// field for field, including non-UTC timestamps (compared as instants).
func TestAppendExtendedCacheEqualsDiskDecode(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	st, err := Open(path, DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Append(LogRecord{Model: "seed"}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.SnapshotView(); err != nil { // arm cache extension
		t.Fatal(err)
	}
	for i := 0; i < 500; i++ {
		rec := LogRecord{T: randTime(r), RequestedAt: randTime(r), Provider: randString(r), Model: randString(r),
			SessionID: randString(r), Generate: r.Intn(2) == 0, LatencyMs: r.Float64() * 1000, TotalTokens: r.Int63n(1e6)}
		if err := st.Append(rec); err != nil {
			t.Fatal(err)
		}
	}
	view, err := st.SnapshotView()
	if err != nil {
		t.Fatal(err)
	}
	decodes := st.Health().SnapshotDecodes
	if err := st.Flush(); err != nil {
		t.Fatal(err)
	}
	disk, err := ReadAll(path)
	if err != nil || len(disk) != len(view.Records) {
		t.Fatal(err, len(disk), len(view.Records))
	}
	for i := range disk {
		a, b := disk[i], view.Records[i]
		if !a.T.Equal(b.T) || !a.RequestedAt.Equal(b.RequestedAt) {
			t.Fatalf("time differs at %d", i)
		}
		a.T, b.T, a.RequestedAt, b.RequestedAt = time.Time{}, time.Time{}, time.Time{}, time.Time{}
		a.Sequence, a.APIKeyID = b.Sequence, ""
		if a != b {
			t.Fatalf("cache diverges from disk at %d:\n%+v\n%+v", i, a, b)
		}
	}
	if st.Health().SnapshotDecodes != decodes {
		t.Fatal("cache was rebuilt from disk")
	}
}

// Append may skip the full manifest check inside observeInterval, but Snapshot
// must still notice an out-of-band edit immediately.
func TestSnapshotDetectsExternalEditDespiteAppendThrottle(t *testing.T) {
	old := observeInterval
	observeInterval = time.Hour
	t.Cleanup(func() { observeInterval = old })
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	st, err := Open(path, DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	for i := 0; i < 3; i++ {
		if err := st.Append(LogRecord{Model: "m"}); err != nil {
			t.Fatal(err)
		}
	}
	first, err := st.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Append(LogRecord{Model: "m"}); err != nil { // throttled: no full check
		t.Fatal(err)
	}
	if err := os.Rename(path, path+".1"); err != nil { // out-of-band rotation
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{\"model\":\"external\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	second, err := st.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if second.Generation == first.Generation {
		t.Fatal("snapshot missed an out-of-band rotation while Append was throttled")
	}
}

// Appends in a single throttle window must not change the generation or force a decode.
func TestAppendThrottleKeepsGenerationAndCache(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "usage.jsonl"), DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Append(LogRecord{Model: "a"}); err != nil {
		t.Fatal(err)
	}
	s1, _ := st.SnapshotView()
	for i := 0; i < 2000; i++ { // crosses several 64 KiB buffer spills and 256-line flushes
		if err := st.Append(LogRecord{Model: "b", Provider: strings.Repeat("p", 40)}); err != nil {
			t.Fatal(err)
		}
	}
	s2, err := st.SnapshotView()
	if err != nil {
		t.Fatal(err)
	}
	if s2.Generation != s1.Generation || len(s2.Records) != 2001 || st.Health().SnapshotDecodes != 1 {
		t.Fatalf("own writes disturbed generation/cache: %d decodes", st.Health().SnapshotDecodes)
	}
}

// An out-of-band change made inside the throttle window must not be swallowed by
// the manifest refresh that follows Append's own file writes: the flush point
// observes first, so the next snapshot still sees a new generation.
func TestExternalChangeNotSwallowedByOwnFlush(t *testing.T) {
	old := observeInterval
	observeInterval = time.Hour
	t.Cleanup(func() { observeInterval = old })
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	st, err := Open(path, DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.Append(LogRecord{Model: "m"}); err != nil {
		t.Fatal(err)
	}
	first, err := st.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".1", []byte("{\"model\":\"external\"}\n"), 0o600); err != nil { // new rotation file appears
		t.Fatal(err)
	}
	for i := 0; i < defaultFlushEvery+1; i++ {
		if err := st.Append(LogRecord{Model: "m"}); err != nil {
			t.Fatal(err)
		}
	}
	second, err := st.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if second.Generation == first.Generation {
		t.Fatal("external change swallowed by Append's own manifest refresh")
	}
}
