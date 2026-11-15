package plugin

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
)

func summaryTotal(t *testing.T, req managementHandleRequest) int {
	t.Helper()
	resp := responseForTest(t, handleGetDashboardSummary(req))
	if resp.StatusCode != 200 {
		t.Fatalf("summary: %d %s", resp.StatusCode, resp.Body)
	}
	var body struct {
		TotalCount int `json:"total_count"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatal(err)
	}
	return body.TotalCount
}

// Repeated polls with no change reuse the computed aggregate (no snapshot read at
// all), and every kind of change invalidates it: append, clear, rotation and an
// out-of-band edit of the log file.
func TestSummaryCacheReuseAndInvalidation(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "usage.jsonl"), 600, 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	installTestState(t, st)
	now := time.Now().UTC()
	appendTestRecord(t, st, store.LogRecord{T: now, Model: "a"})
	// The dashboard always sends from/to. (Without "to" the resource route defaults
	// to the current instant, so each poll is a different query and is not reused.)
	req := managementHandleRequest{Path: mgmtBase + "/data/summary", Query: benchQuery(now.Add(-time.Hour), now.Add(time.Hour))}

	if n := summaryTotal(t, req); n != 1 {
		t.Fatalf("total = %d", n)
	}
	before := st.Health()
	first := handleGetDashboardSummary(req)
	second := handleGetDashboardSummary(req)
	after := st.Health()
	if after.CacheHits != before.CacheHits || after.SnapshotDecodes != before.SnapshotDecodes {
		t.Fatalf("unchanged poll read the snapshot again: before %+v after %+v", before, after)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("cached summary differs between identical polls")
	}
	// Another query (different filter / scope) is its own entry.
	other := managementHandleRequest{Path: mgmtBase + "/summary", Query: map[string][]string{"model": {"zzz"}}}
	if resp := responseForTest(t, handleGetSummary(other)); !bytes.Contains(resp.Body, []byte(`"total_count":0`)) {
		t.Fatalf("filtered summary: %s", resp.Body)
	}

	appendTestRecord(t, st, store.LogRecord{T: now, Model: "b"})
	if n := summaryTotal(t, req); n != 2 {
		t.Fatalf("after append total = %d, want 2", n)
	}
	if resp := responseForTest(t, handleDeleteRecords()); resp.StatusCode != 200 {
		t.Fatal(string(resp.Body))
	}
	if n := summaryTotal(t, req); n != 0 {
		t.Fatalf("after clear total = %d, want 0", n)
	}
	// Small maxBytes: these appends rotate the active file.
	for i := 0; i < 6; i++ {
		appendTestRecord(t, st, store.LogRecord{T: now, Model: "rotating-model-name"})
	}
	if n := summaryTotal(t, req); n != 6 {
		t.Fatalf("after rotation total = %d, want 6", n)
	}
	// Out-of-band edit: drop the rotated history behind the store's back.
	if err := st.Flush(); err != nil {
		t.Fatal(err)
	}
	rotated, _ := filepath.Glob(filepath.Join(dir, "usage.jsonl.*"))
	if len(rotated) == 0 {
		t.Fatal("expected rotated files")
	}
	for _, p := range rotated {
		if err := os.Remove(p); err != nil {
			t.Fatal(err)
		}
	}
	if n := summaryTotal(t, req); n >= 6 {
		t.Fatalf("external edit not noticed: total = %d", n)
	}
}

func TestSummaryCacheIsBoundedLRU(t *testing.T) {
	var c summaryCache
	for i := 0; i < summaryCacheSize+3; i++ {
		c.put(string(rune('a'+i)), &summaryAggregate{TotalCount: i})
	}
	if len(c.entries) != summaryCacheSize {
		t.Fatalf("entries = %d, want %d", len(c.entries), summaryCacheSize)
	}
	if _, ok := c.get("a"); ok {
		t.Fatal("oldest entry not evicted")
	}
	last := string(rune('a' + summaryCacheSize + 2))
	if agg, ok := c.get(last); !ok || agg.TotalCount != summaryCacheSize+2 {
		t.Fatal("newest entry missing")
	}
}
