package plugin

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/queue"
	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
)

func TestThinkingAndServiceTierEndToEnd(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "usage.jsonl"), store.DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	q := queue.New(64)
	ps := &pluginState{st: st, q: q}
	ps.workerWg.Add(1)
	go ps.consumeLoop()
	stateMu.Lock()
	prev := state
	state = ps
	stateMu.Unlock()
	t.Cleanup(func() {
		ps.shutdown()
		stateMu.Lock()
		state = prev
		stateMu.Unlock()
	})

	t1 := time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 11, 24, 10, 1, 0, 0, time.UTC)
	t3 := time.Date(2026, 11, 24, 10, 2, 0, 0, time.UTC)

	payloads := []map[string]any{
		{
			"Provider":             "claude",
			"ExecutorType":         "claude",
			"Model":                "claude-opus-5-5",
			"ReasoningEffort":      "high",
			"ServiceTier":          "fast",
			"EffectiveServiceTier": "fast",
			"ResponseServiceTier":  "fast",
			"Generate":             true,
			"RequestedAt":          t1.Format(time.RFC3339),
			"Detail": map[string]any{
				"InputTokens": 100, "OutputTokens": 80, "ReasoningTokens": 45, "TotalTokens": 180,
			},
		},
		{
			"Provider":             "claude",
			"ExecutorType":         "claude",
			"Model":                "claude-haiku-5-5",
			"ServiceTier":          "fast",
			"EffectiveServiceTier": "fast",
			"ResponseServiceTier":  "standard", // fallback
			"Generate":             true,
			"RequestedAt":          t2.Format(time.RFC3339),
			"Detail": map[string]any{
				"InputTokens": 50, "OutputTokens": 20, "TotalTokens": 70,
			},
		},
		{
			"Provider":             "codex",
			"ExecutorType":         "codex",
			"Model":                "gpt-5",
			"ReasoningEffort":      "medium",
			"ServiceTier":          "priority",
			"EffectiveServiceTier": "priority",
			"ResponseServiceTier":  "priority",
			"Generate":             true,
			"RequestedAt":          t3.Format(time.RFC3339),
			"Detail": map[string]any{
				"InputTokens": 200, "OutputTokens": 100, "ReasoningTokens": 60, "TotalTokens": 300,
			},
		},
	}

	for _, p := range payloads {
		raw, _ := json.Marshal(p)
		handleUsage(raw)
	}

	// Wait for consumer to persist all 3 records.
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		recs, err := st.ReadAllView()
		if err == nil && len(recs) == 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Check dashboard /data/records response contains all 4 fields.
	envRaw := handleGetDashboardRecords(managementHandleRequest{Method: "GET", Path: mgmtBase + "/data/records"})
	var env struct {
		OK     bool               `json:"ok"`
		Result managementResponse `json:"result"`
	}
	if err := json.Unmarshal(envRaw, &env); err != nil || !env.OK || env.Result.StatusCode != 200 {
		t.Fatalf("unexpected envelope: %s err=%v", string(envRaw), err)
	}
	var body struct {
		Records []resourceRecord `json:"records"`
	}
	if err := json.Unmarshal(env.Result.Body, &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(body.Records))
	}
	// Newest first: t3 (codex priority), t2 (claude fallback), t1 (claude fast + high)
	if body.Records[0].ReasoningEffort != "medium" || body.Records[0].ResponseServiceTier != "priority" || body.Records[0].EffectiveServiceTier != "priority" {
		t.Fatalf("record 0 mismatch: %+v", body.Records[0])
	}
	if body.Records[1].ServiceTier != "fast" || body.Records[1].ResponseServiceTier != "standard" {
		t.Fatalf("record 1 fallback mismatch: %+v", body.Records[1])
	}
	if body.Records[2].ReasoningEffort != "high" || body.Records[2].ResponseServiceTier != "fast" || body.Records[2].Reasoning != 45 {
		t.Fatalf("record 2 mismatch: %+v", body.Records[2])
	}
}

func TestSelectRecordsPageByNumberDeepAndBlockSkipping(t *testing.T) {
	// Create 3000 records across 3 blocks (> maxBoundedPageHeap = 2048) with ties and out-of-order entries.
	base := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	const n = 3000
	recs := make([]store.LogRecord, n)
	for i := 0; i < n; i++ {
		// Introduce timestamp ties every 3 records and out-of-order jumps every 37 records.
		ts := base.Add(time.Duration(i/3) * time.Minute)
		if i%37 == 0 {
			ts = base.Add(time.Duration(n-i) * time.Minute)
		}
		recs[i] = store.LogRecord{
			Sequence:    uint64(i + 1),
			T:           ts,
			RequestedAt: ts,
			Provider:    []string{"claude", "codex", "gemini"}[i%3],
			Model:       fmt.Sprintf("model-%d", i%5),
			Failed:      i%11 == 0,
		}
	}
	// Build blocks matching store.SnapshotBlockSize.
	var blocks []store.BlockMeta
	for start := 0; start < n; start += store.SnapshotBlockSize {
		end := min(start+store.SnapshotBlockSize, n)
		minTS, maxTS := recs[start].EffectiveTimestamp(), recs[start].EffectiveTimestamp()
		for i := start + 1; i < end; i++ {
			ts := recs[i].EffectiveTimestamp()
			if ts.Before(minTS) {
				minTS = ts
			}
			if ts.After(maxTS) {
				maxTS = ts
			}
		}
		blocks = append(blocks, store.BlockMeta{Start: start, End: end, MinTS: minTS, MaxTS: maxTS})
	}

	from := base.Add(100 * time.Minute)
	to := base.Add(850 * time.Minute)
	filterCases := []store.FilterOptions{
		{},
		{From: &from, To: &to},
		{Provider: "claude"},
		{From: &from, To: &to, Provider: "codex", Status: "success"},
	}
	pageCases := []struct{ page, limit int }{
		{1, 20},
		{2, 20},
		{50, 20},
		{110, 20}, // start=2180 -> k=2201 > maxBoundedPageHeap (exercises 3-way Quickselect!)
		{150, 20}, // exact end boundary
		{151, 20}, // out-of-range
		{1, 1},
	}

	for fi, opts := range filterCases {
		wantAll := store.FilterRecords(recs, opts)
		for _, pc := range pageCases {
			got, total, hasMore := selectRecordsPageByNumberBlocks(recs, blocks, opts, pc.page, pc.limit)
			if total != len(wantAll) {
				t.Fatalf("filter %d page %d: total=%d want=%d", fi, pc.page, total, len(wantAll))
			}
			start := (pc.page - 1) * pc.limit
			if start >= len(wantAll) {
				if len(got) != 0 || hasMore {
					t.Fatalf("filter %d page %d: expected empty, got len=%d hasMore=%v", fi, pc.page, len(got), hasMore)
				}
				continue
			}
			end := min(start+pc.limit, len(wantAll))
			wantMore := end < len(wantAll)
			if hasMore != wantMore {
				t.Fatalf("filter %d page %d: hasMore=%v want=%v", fi, pc.page, hasMore, wantMore)
			}
			if len(got) != end-start {
				t.Fatalf("filter %d page %d: len=%d want=%d", fi, pc.page, len(got), end-start)
			}
			for i := range got {
				if got[i].Sequence != wantAll[start+i].Sequence {
					t.Fatalf("filter %d page %d idx %d: seq=%d want=%d", fi, pc.page, i, got[i].Sequence, wantAll[start+i].Sequence)
				}
			}
		}
	}
}

func TestRecordsCacheHitAndInvalidation(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "usage.jsonl"), store.DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	base := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	if err := st.Append(store.LogRecord{T: base, RequestedAt: base, Provider: "claude", Model: "m1"}); err != nil {
		t.Fatal(err)
	}
	ps := &pluginState{st: st, q: queue.New(8)}
	defer ps.q.Close()

	opts := store.FilterOptions{}
	p1, _, total1, _, err := cachedRecordsPageQuery(ps, opts, "resource", "", 1, 20)
	if err != nil || total1 != 1 || len(p1) != 1 {
		t.Fatalf("first query: total=%d len=%d err=%v", total1, len(p1), err)
	}
	hitsBefore := st.Health().CacheHits
	// Second identical call should hit ps.records without even calling SnapshotView.
	p2, _, total2, _, err := cachedRecordsPageQuery(ps, opts, "resource", "", 1, 20)
	if err != nil || total2 != 1 || len(p2) != 1 {
		t.Fatalf("cached query: total=%d len=%d err=%v", total2, len(p2), err)
	}
	if st.Health().CacheHits != hitsBefore {
		t.Fatalf("expected recordsCache hit to skip store snapshot: hits %d -> %d", hitsBefore, st.Health().CacheHits)
	}

	// Append invalidates version and returns the new record.
	if err := st.Append(store.LogRecord{T: base.Add(time.Minute), RequestedAt: base.Add(time.Minute), Provider: "codex", Model: "m2"}); err != nil {
		t.Fatal(err)
	}
	p3, _, total3, _, err := cachedRecordsPageQuery(ps, opts, "resource", "", 1, 20)
	if err != nil || total3 != 2 || len(p3) != 2 || p3[0].Model != "m2" {
		t.Fatalf("after append: total=%d len=%d first=%+v err=%v", total3, len(p3), p3, err)
	}
}

func TestDashboardRecordsExposeAuthRoutingWithoutSecrets(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "usage.jsonl"), store.DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	q := queue.New(64)
	ps := &pluginState{st: st, q: q}
	ps.workerWg.Add(1)
	go ps.consumeLoop()
	stateMu.Lock()
	prev := state
	state = ps
	stateMu.Unlock()
	t.Cleanup(func() {
		ps.shutdown()
		stateMu.Lock()
		state = prev
		stateMu.Unlock()
	})

	t1 := time.Date(2026, 11, 24, 11, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 11, 24, 11, 1, 0, 0, time.UTC)
	t3 := time.Date(2026, 11, 24, 11, 2, 0, 0, time.UTC)
	payloads := []map[string]any{
		{
			"Provider":    "codex",
			"Model":       "gpt-5",
			"AuthID":      "secret-auth-id-1.json",
			"AuthIndex":   "0",
			"AuthType":    "oauth",
			"APIKey":      "sk-client-secret-should-never-appear",
			"Source":      "secret-source-email@example.com",
			"Generate":    true,
			"RequestedAt": t1.Format(time.RFC3339),
			"Detail":      map[string]any{"InputTokens": 10, "OutputTokens": 5, "TotalTokens": 15},
		},
		{
			"Provider":    "claude",
			"Model":       "claude-opus-5-5",
			"AuthID":      "secret-auth-id-2",
			"AuthIndex":   "idx-claude-api",
			"AuthType":    "api_key",
			"APIKey":      "sk-ant-raw-upstream-secret",
			"Source":      "sk-ant-raw-upstream-secret",
			"Generate":    true,
			"RequestedAt": t2.Format(time.RFC3339),
			"Detail":      map[string]any{"InputTokens": 20, "OutputTokens": 10, "TotalTokens": 30},
		},
		{
			"Provider":    "openai",
			"Model":       "gpt-4.1",
			"Generate":    true,
			"RequestedAt": t3.Format(time.RFC3339),
			"Detail":      map[string]any{"InputTokens": 5, "OutputTokens": 5, "TotalTokens": 10},
		},
	}
	for _, p := range payloads {
		raw, _ := json.Marshal(p)
		handleUsage(raw)
	}
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		recs, err := st.ReadAllView()
		if err == nil && len(recs) == 3 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	envRaw := handleGetDashboardRecords(managementHandleRequest{Method: "GET", Path: mgmtBase + "/data/records"})
	var env struct {
		OK     bool               `json:"ok"`
		Result managementResponse `json:"result"`
	}
	if err := json.Unmarshal(envRaw, &env); err != nil || !env.OK || env.Result.StatusCode != 200 {
		t.Fatalf("unexpected envelope: %s err=%v", string(envRaw), err)
	}
	rawBody := string(env.Result.Body)
	for _, forbidden := range []string{
		"secret-auth-id-1.json",
		"secret-auth-id-2",
		"sk-client-secret-should-never-appear",
		"sk-ant-raw-upstream-secret",
		"secret-source-email@example.com",
	} {
		if strings.Contains(rawBody, forbidden) {
			t.Fatalf("/data/records leaked forbidden field %q in %s", forbidden, rawBody)
		}
	}
	var rawBodyObj struct {
		Records []map[string]any `json:"records"`
	}
	if err := json.Unmarshal(env.Result.Body, &rawBodyObj); err != nil {
		t.Fatal(err)
	}
	if len(rawBodyObj.Records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(rawBodyObj.Records))
	}
	// Empty auth_index / auth_type must be omitted from JSON via omitempty.
	if _, exists := rawBodyObj.Records[0]["auth_index"]; exists {
		t.Fatalf("empty auth_index should be omitted from JSON: %+v", rawBodyObj.Records[0])
	}
	if _, exists := rawBodyObj.Records[0]["auth_type"]; exists {
		t.Fatalf("empty auth_type should be omitted from JSON: %+v", rawBodyObj.Records[0])
	}
	var body struct {
		Records []resourceRecord `json:"records"`
	}
	if err := json.Unmarshal(env.Result.Body, &body); err != nil {
		t.Fatal(err)
	}
	if body.Records[1].AuthIndex != "idx-claude-api" || body.Records[1].AuthType != "api_key" {
		t.Fatalf("record 1 auth mismatch: %+v", body.Records[1])
	}
	if body.Records[2].AuthIndex != "0" || body.Records[2].AuthType != "oauth" {
		t.Fatalf("record 2 auth mismatch (zero string must survive): %+v", body.Records[2])
	}

	// /data/summary must remain auth_index-free.
	sumRaw := handleGetDashboardSummary(managementHandleRequest{Method: "GET", Path: mgmtBase + "/data/summary"})
	if strings.Contains(string(sumRaw), "idx-claude-api") || strings.Contains(string(sumRaw), `"auth_index"`) {
		t.Fatalf("/data/summary must not expose auth_index: %s", string(sumRaw))
	}
}
