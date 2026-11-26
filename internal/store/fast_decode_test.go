package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFastDecodeEquivalenceWithStdlib(t *testing.T) {
	now := time.Date(2026, 11, 24, 15, 8, 12, 123456789, time.UTC)
	nonUTC := time.Date(2026, 11, 24, 23, 59, 59, 0, time.FixedZone("KST", 9*3600))
	cases := []struct {
		name string
		raw  string
	}{
		{
			name: "standard_full_record",
			raw: func() string {
				b, _ := json.Marshal(LogRecord{
					T: now, RequestedAt: now.Add(-time.Second),
					Provider: "claude", ExecutorType: "claude", Model: "claude-opus-5-5", Alias: "opus",
					SessionID: "sess-1", ParentSessionID: "parent-1", AuthIndex: "auth-0", AuthType: "oauth",
					ReasoningEffort: "high", ServiceTier: "fast", EffectiveServiceTier: "fast", ResponseServiceTier: "fast",
					Generate: true, LatencyMs: 1234.5, TTFTMs: 210.25, Failed: false,
					InputTokens: 1000, OutputTokens: 500, ReasoningTokens: 300,
					CachedTokens: 200, CacheReadTokens: 150, CacheCreationTokens: 50, TotalTokens: 1700,
				})
				return string(b)
			}(),
		},
		{
			name: "empty_object",
			raw:  `{}`,
		},
		{
			name: "escaped_and_unicode_strings",
			raw:  `{"t":"2026-11-24T15:08:12Z","model":"claude\"test\\v1","provider":"한글프로바이더","reasoning_effort":"max","response_service_tier":"standard","latency_ms":1.5e2,"failed":true,"failure_status":429}`,
		},
		{
			name: "non_utc_timestamp",
			raw: func() string {
				b, _ := json.Marshal(LogRecord{T: nonUTC, RequestedAt: nonUTC, Model: "gpt", Generate: false})
				return string(b)
			}(),
		},
		{
			name: "unknown_field_fallback",
			raw:  `{"t":"2026-11-24T15:08:12Z","model":"m1","extra_nested":{"a":[1,2,3]},"input_tokens":42}`,
		},
		{
			name: "negative_zero_and_large_int",
			raw:  `{"t":"2026-11-24T15:08:12Z","latency_ms":-0.0,"input_tokens":9223372036854775807}`,
		},
	}

	intern := make(map[string]string)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var want, got LogRecord
			wantErr := json.Unmarshal([]byte(tc.raw), &want)
			ok := decodeLogRecordFast([]byte(tc.raw), &got, intern)
			if (wantErr == nil) != ok {
				t.Fatalf("ok=%v wantErr=%v", ok, wantErr)
			}
			if ok && !recordsEqual(got, want) {
				t.Fatalf("mismatch:\ngot:  %+v\nwant: %+v", got, want)
			}
		})
	}

	invalidLines := []string{
		`[]`,
		`"just a string"`,
		`123`,
		`{"t":"2026-02-30T00:00:00Z"}`,
		`{"input_tokens":01}`,
		`{"latency_ms":01.5}`,
		`{"t":"2026-11-24T15:08:12Z"}{"t":"2026-11-24T15:08:12Z"}`,
		`{"t":"2026-11-24T15:08:12Z"`,
	}
	for _, raw := range invalidLines {
		var got LogRecord
		if decodeLogRecordFast([]byte(raw), &got, intern) {
			t.Fatalf("expected invalid line to fail: %q", raw)
		}
	}
}

func recordsEqual(a, b LogRecord) bool {
	if !a.T.Equal(b.T) || !a.RequestedAt.Equal(b.RequestedAt) {
		return false
	}
	a.T, b.T = time.Time{}, time.Time{}
	a.RequestedAt, b.RequestedAt = time.Time{}, time.Time{}
	return a == b
}

func TestLongLineExceedingReadBuffer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	longModel := strings.Repeat("m", 300*1024)
	rec := LogRecord{
		T:                    time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC),
		Model:                longModel,
		EffectiveServiceTier: "priority",
		ResponseServiceTier:  "fast",
		InputTokens:          10,
	}
	b, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	recs, err := ReadAll(path)
	if err != nil || len(recs) != 1 {
		t.Fatalf("ReadAll: len=%d err=%v", len(recs), err)
	}
	if recs[0].Model != longModel || recs[0].EffectiveServiceTier != "priority" || recs[0].ResponseServiceTier != "fast" {
		t.Fatalf("unexpected decoded record: len(model)=%d eff=%q resp=%q", len(recs[0].Model), recs[0].EffectiveServiceTier, recs[0].ResponseServiceTier)
	}
}

func TestSnapshotBlocksAndConcurrentAppendRace(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "usage.jsonl"), DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	base := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	// Seed across >2 blocks including out-of-order timestamps.
	const initial = SnapshotBlockSize*2 + 15
	for i := 0; i < initial; i++ {
		ts := base.Add(time.Duration(i) * time.Minute)
		if i == SnapshotBlockSize+5 {
			ts = base.Add(-24 * time.Hour) // backdated outlier in block 1
		}
		if err := st.Append(LogRecord{
			T: ts, RequestedAt: ts, Provider: "claude", Model: "claude-opus-5-5",
			ReasoningEffort: "high", ResponseServiceTier: "fast", InputTokens: 10, TotalTokens: 10,
		}); err != nil {
			t.Fatal(err)
		}
	}
	snap, err := st.SnapshotView()
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Blocks) != 3 {
		t.Fatalf("expected 3 blocks, got %d", len(snap.Blocks))
	}
	if !snap.Blocks[1].MinTS.Equal(base.Add(-24 * time.Hour)) {
		t.Fatalf("block 1 MinTS missed backdated outlier: %v", snap.Blocks[1].MinTS)
	}

	// Concurrent SnapshotView + Append must be completely data-race-free.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 300; i++ {
			ts := base.Add(time.Duration(initial+i) * time.Minute)
			if err := st.Append(LogRecord{T: ts, RequestedAt: ts, Provider: "codex", Model: "gpt-5"}); err != nil {
				t.Errorf("append %d: %v", i, err)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 150; i++ {
			v, err := st.SnapshotView()
			if err != nil {
				t.Errorf("snapshot %d: %v", i, err)
				return
			}
			// Read every block & boundary element to exercise -race.
			total := 0
			for _, b := range v.Blocks {
				if b.Start < 0 || b.End > len(v.Records) || b.Start >= b.End {
					t.Errorf("invalid block bounds %+v for len=%d", b, len(v.Records))
					return
				}
				_ = v.Records[b.Start].Model
				_ = v.Records[b.End-1].Model
				total += b.End - b.Start
			}
			if total != len(v.Records) {
				t.Errorf("block coverage mismatch: %d != %d", total, len(v.Records))
				return
			}
		}
	}()
	wg.Wait()
}

func TestScrubChunkBoundarySecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	var buf bytes.Buffer
	line := `{"t":"2026-11-24T00:00:00Z","model":"m","input_tokens":1}` + "\n"
	// Pad close to 256 KiB so the dirty record straddles the chunk boundary.
	for buf.Len() < 256*1024-12 {
		buf.WriteString(line)
	}
	buf.WriteString(`{"t":"2026-11-24T00:00:01Z","api_key_id":"sk-secret","model":"m"}` + "\n")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	n, err := ScrubLegacySecrets(path)
	if err != nil || n != 1 {
		t.Fatalf("ScrubLegacySecrets across chunk boundary: n=%d err=%v", n, err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(raw, []byte("sk-secret")) || bytes.Contains(raw, []byte(`"api_key_id"`)) {
		t.Fatal("legacy secret survived chunk-boundary scrub")
	}
}

func TestRotatedSegmentCachePreservedAndCleared(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "usage.jsonl")
	st, err := Open(path, 250, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	base := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 6; i++ {
		ts := base.Add(time.Duration(i) * time.Second)
		if err := st.Append(LogRecord{T: ts, RequestedAt: ts, Provider: "claude", Model: fmt.Sprintf("m-%d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	snap1, err := st.SnapshotView()
	if err != nil || len(snap1.Records) == 0 {
		t.Fatalf("snap1: len=%d err=%v", len(snap1.Records), err)
	}
	st.mu.Lock()
	cachedCount := len(st.cachedFiles)
	st.mu.Unlock()
	if cachedCount == 0 {
		t.Fatal("expected rotated file segments to be cached")
	}
	if err := st.Clear(); err != nil {
		t.Fatal(err)
	}
	st.mu.Lock()
	afterClear := len(st.cachedFiles)
	st.mu.Unlock()
	if afterClear != 0 {
		t.Fatalf("expected cachedFiles cleared after Clear(), got %d", afterClear)
	}
}
