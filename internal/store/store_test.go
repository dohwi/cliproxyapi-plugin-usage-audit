package store_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-usage-audit/internal/store"
)

func TestFilterRecords(t *testing.T) {
	t1 := time.Date(2025, 1, 10, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 11, 10, 0, 0, 0, time.UTC)
	t3 := time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC)

	recs := []store.LogRecord{
		{T: t1, RequestedAt: t1, Provider: "p1", Model: "m1", FailureStatus: 0, Failed: false, Source: "s1", ExecutorType: "e1", AuthIndex: "a1"},
		{T: t2, RequestedAt: t2, Provider: "p2", Model: "m2", FailureStatus: 500, Failed: true, Source: "s2", ExecutorType: "e2", AuthIndex: "a2"},
		{T: t3, RequestedAt: t3, Provider: "p1", Model: "m2", FailureStatus: 0, Failed: false, Source: "s1", ExecutorType: "e2", AuthIndex: "a1"},
	}

	// Default newest-first
	all := store.FilterRecords(recs, store.FilterOptions{})
	if len(all) != 3 {
		t.Fatalf("expected 3 records, got %d", len(all))
	}
	if all[0].Provider != "p1" || all[0].Model != "m2" {
		t.Errorf("newest first failed: got %+v", all[0])
	}

	// Filter by Provider
	p1s := store.FilterRecords(recs, store.FilterOptions{Provider: "p1"})
	if len(p1s) != 2 {
		t.Fatalf("expected 2 records for p1, got %d", len(p1s))
	}

	// Filter by Status failed
	failed := store.FilterRecords(recs, store.FilterOptions{Status: "failed"})
	if len(failed) != 1 || failed[0].Provider != "p2" {
		t.Fatalf("expected 1 failed record (p2), got %+v", failed)
	}

	// Filter by Status success
	success := store.FilterRecords(recs, store.FilterOptions{Status: "success"})
	if len(success) != 2 {
		t.Fatalf("expected 2 success records, got %d", len(success))
	}

	// Filter by Date boundary
	fromT := time.Date(2025, 1, 11, 0, 0, 0, 0, time.UTC)
	toT := time.Date(2025, 1, 11, 23, 59, 59, 0, time.UTC)
	dateRange := store.FilterRecords(recs, store.FilterOptions{From: &fromT, To: &toT})
	if len(dateRange) != 1 || dateRange[0].Provider != "p2" {
		t.Fatalf("expected 1 record in range (p2), got %+v", dateRange)
	}
}

func TestFilterRecordsEndpointORMatch(t *testing.T) {
	recs := []store.LogRecord{
		{T: time.Date(2025, 1, 10, 10, 0, 0, 0, time.UTC), ExecutorType: "exec-a", Source: ""},
		{T: time.Date(2025, 1, 11, 10, 0, 0, 0, time.UTC), ExecutorType: "", Source: "src-b"},
		{T: time.Date(2025, 1, 12, 10, 0, 0, 0, time.UTC), ExecutorType: "exec-c", Source: "src-c"},
	}

	if got := len(store.FilterRecords(recs, store.FilterOptions{Endpoint: "exec-a"})); got != 1 {
		t.Errorf("endpoint=exec-a want 1, got %d", got)
	}
	if got := len(store.FilterRecords(recs, store.FilterOptions{Endpoint: "src-b"})); got != 1 {
		t.Errorf("endpoint=src-b want 1, got %d", got)
	}
	if got := len(store.FilterRecords(recs, store.FilterOptions{Endpoint: "src-c"})); got != 1 {
		t.Errorf("endpoint=src-c want 1, got %d", got)
	}
	if got := len(store.FilterRecords(recs, store.FilterOptions{Endpoint: "nope"})); got != 0 {
		t.Errorf("endpoint=nope want 0, got %d", got)
	}

	// Both records share the same ExecutorType and Source value: endpoint must
	// match once, and explicit executor_type+source must AND to 1 (not hide it).
	both := []store.LogRecord{{T: time.Now().UTC(), ExecutorType: "x", Source: "x"}}
	if got := len(store.FilterRecords(both, store.FilterOptions{Endpoint: "x"})); got != 1 {
		t.Errorf("endpoint=x want 1, got %d", got)
	}
	if got := len(store.FilterRecords(both, store.FilterOptions{ExecutorType: "x", Source: "x"})); got != 1 {
		t.Errorf("executor_type+source AND want 1, got %d", got)
	}
}

func TestRetainedRotationHistoryAndClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	st, err := store.Open(path, 1, 3) // one record per file, without an empty first rotation
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	for i := 0; i < 6; i++ {
		if err := st.Append(store.LogRecord{Model: strconv.Itoa(i)}); err != nil {
			t.Fatal(err)
		}
	}
	recs, err := st.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 4 {
		t.Fatalf("retained count = %d, want 4", len(recs))
	}
	for i, r := range recs {
		if r.Model != strconv.Itoa(i+2) {
			t.Errorf("record %d = %s", i, r.Model)
		}
	}
	// Discover older configuration rotations too; ignore unrelated suffixes.
	if err := os.Rename(path+".3", path+".10"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".bak", []byte("unrelated"), 0o600); err != nil {
		t.Fatal(err)
	}
	recs, err = st.ReadAll()
	if err != nil || len(recs) != 4 || recs[0].Model != "2" {
		t.Fatalf("numeric rotation order: %+v, %v", recs, err)
	}
	if err := st.Clear(); err != nil {
		t.Fatal(err)
	}
	recs, err = st.ReadAll()
	if err != nil || len(recs) != 0 {
		t.Fatalf("after clear: %+v, %v", recs, err)
	}
	for _, suffix := range []string{".1", ".2", ".3", ".10"} {
		if _, err := os.Stat(path + suffix); !os.IsNotExist(err) {
			t.Errorf("rotation %s remains: %v", suffix, err)
		}
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatal("unrelated file deleted", err)
	}
	if err := st.Append(store.LogRecord{Model: "after"}); err != nil {
		t.Fatal(err)
	}
	recs, err = st.ReadAll()
	if err != nil || len(recs) != 1 || recs[0].Model != "after" {
		t.Fatalf("writer did not resume: %+v, %v", recs, err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	if err := st.Close(); err != nil {
		t.Fatal("close not idempotent", err)
	}
	if err := st.Append(store.LogRecord{}); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("append after close = %v", err)
	}
}

func TestAppendSeparatesPreexistingUnterminatedTail(t *testing.T) {
	for _, tc := range []struct {
		name string
		tail string
		want []string
	}{
		{"valid tail", `{"model":"existing"}`, []string{"existing", "appended"}},
		{"truncated tail", `{"model":"approved-corrupt`, []string{"appended"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "usage.jsonl")
			if err := os.WriteFile(path, []byte(tc.tail), 0o600); err != nil {
				t.Fatal(err)
			}
			st, err := store.Open(path, store.DefaultMaxBytes, 2)
			if err != nil {
				t.Fatal(err)
			}
			defer st.Close()
			if err := st.Append(store.LogRecord{Model: "appended"}); err != nil {
				t.Fatal(err)
			}
			recs, err := st.ReadAll()
			if err != nil || len(recs) != len(tc.want) {
				t.Fatalf("records=%+v err=%v", recs, err)
			}
			for i := range recs {
				if recs[i].Model != tc.want[i] {
					t.Fatalf("record %d=%q want %q", i, recs[i].Model, tc.want[i])
				}
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(data, []byte(tc.tail+"\n")) {
				t.Fatalf("tail was not preserved/separated: %q", data)
			}
		})
	}
}

func TestStoreNeverPersistsAPIKeyID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	st, err := store.Open(path, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	for i := 0; i < 3; i++ {
		if err := st.Append(store.LogRecord{APIKeyID: "sk-raw-secret-must-not-persist", AuthIndex: "safe-index"}); err != nil {
			t.Fatal(err)
		}
	}
	recs, err := st.ReadAll()
	if err != nil || len(recs) != 3 {
		t.Fatalf("records: %d %v", len(recs), err)
	}
	for _, suffix := range []string{"", ".1", ".2"} {
		data, err := os.ReadFile(path + suffix)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(data, []byte("sk-raw-secret")) || bytes.Contains(data, []byte("api_key_id")) {
			t.Fatalf("secret persisted: %s", data)
		}
	}
}

func TestJSONLResilientAndStrictPolicies(t *testing.T) {
	valid := `{"model":"valid"}`
	for _, tc := range []struct {
		name, data string
		count      int
		strictBad  bool
	}{
		{"complete", valid + "\n", 1, false},
		{"no final newline", valid, 1, false},
		{"blank lines", "\n \n" + valid + "\n\t", 1, false},
		{"truncated final", valid + "\n" + `{"model":"part`, 1, true},
		{"truncated final unicode", valid + "\n" + `{"model":"\u12`, 1, true},
		{"truncated terminated", valid + "\n" + `{"model":"part` + "\n", 1, true},
		{"truncated middle", valid + "\n" + `{"model":` + "\n" + valid + "\n", 2, true},
		{"invalid middle", valid + "\nNOT JSON\n" + valid + "\n", 2, true},
		{"invalid final", valid + "\n" + `{"model":!}`, 1, true},
		{"wrong type final", valid + "\n" + `{"input_tokens":"bad"}`, 1, true},
		{"two objects one line", valid + valid + "\n", 0, true},
		{"null record", "null\n", 0, true},
		{"large line", `{"model":"` + strings.Repeat("x", 100000) + `"}` + "\n", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "usage.jsonl")
			if err := os.WriteFile(path, []byte(tc.data), 0o600); err != nil {
				t.Fatal(err)
			}
			recs, err := store.ReadAll(path)
			if err != nil || len(recs) != tc.count {
				t.Fatalf("resilient count=%d err=%v; want count=%d", len(recs), err, tc.count)
			}
			_, err = store.ReadAllStrict(path)
			if (err != nil) != tc.strictBad {
				t.Fatalf("strict err=%v; want bad=%v", err, tc.strictBad)
			}
			if err != nil && (!errors.Is(err, store.ErrCorruptJSONL) || strings.Contains(err.Error(), "usage.jsonl")) {
				t.Fatalf("strict error is not safe/category-only: %v", err)
			}
		})
	}
}

func TestRequestTimestampSortAndFallback(t *testing.T) {
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	recs := []store.LogRecord{
		{Model: "new-request", T: base, RequestedAt: base.Add(3 * time.Hour)},
		{Model: "fallback", T: base.Add(2 * time.Hour)},
		{Model: "late-completion", T: base.Add(4 * time.Hour), RequestedAt: base},
		{Model: "tie-later-ingest", T: base.Add(5 * time.Hour), RequestedAt: base.Add(3 * time.Hour)},
	}
	got := store.FilterRecords(recs, store.FilterOptions{})
	want := []string{"tie-later-ingest", "new-request", "fallback", "late-completion"}
	for i, r := range got {
		if r.Model != want[i] {
			t.Errorf("position %d = %s, want %s", i, r.Model, want[i])
		}
	}
	if recs[0].Model != "new-request" {
		t.Fatal("input mutated")
	}
	from := base.Add(time.Hour)
	if got := store.FilterRecords(recs, store.FilterOptions{From: &from}); len(got) != 3 {
		t.Fatalf("effective filter count = %d", len(got))
	}
}

func TestConcurrentAppendSnapshotAndClear(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	st, err := store.Open(path, 2048, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	var wg sync.WaitGroup
	for worker := 0; worker < 3; worker++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				var err error
				switch worker {
				case 0:
					err = st.Append(store.LogRecord{Model: strconv.Itoa(i)})
				case 1:
					_, err = st.ReadAll()
				case 2:
					err = st.Clear()
				}
				if err != nil {
					t.Errorf("worker %d: %v", worker, err)
					return
				}
			}
		}(worker)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("store operations deadlocked")
	}
	if err := st.Clear(); err != nil {
		t.Fatal(err)
	}
	if err := st.Append(store.LogRecord{Model: "final"}); err != nil {
		t.Fatal(err)
	}
	recs, err := st.ReadAll()
	if err != nil || len(recs) != 1 || recs[0].Model != "final" {
		t.Fatalf("final snapshot: %+v, %v", recs, err)
	}
}

func TestOperationalResilientDecodingWithCorruptedLinesAndRotations(t *testing.T) {
	tmpDir := t.TempDir()
	activePath := filepath.Join(tmpDir, "usage.jsonl")
	rot1Path := activePath + ".1"

	// Operational corrupted lines that must be preserved verbatim.
	corruptedLine1 := `{"t":"2026-09-12T10:00:00Z", truncated json line`
	corruptedLine2 := `THIS IS NOT JSON AT ALL`
	multiObjectLine := `{"model":"multi-1"}{"model":"multi-2"}`
	nonObjectLine := `[{"model":"array-item"}]`
	truncatedTail := `{"model":"approved-corrupt-tail`

	rotContent := `{"model":"rot-before"}` + "\n" +
		corruptedLine1 + "\n" +
		`{"model":"rot-after"}` + "\n"

	activeContent := `{"model":"active-1"}` + "\n" +
		corruptedLine2 + "\n" +
		multiObjectLine + "\n" +
		nonObjectLine + "\n" +
		`{"model":"active-2"}` + "\n" +
		truncatedTail

	if err := os.WriteFile(rot1Path, []byte(rotContent), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(activePath, []byte(activeContent), 0o640); err != nil {
		t.Fatal(err)
	}

	// 1. Path-based resilient ReadAll skips intermediate corrupt lines and truncated tail.
	recs, err := store.ReadAll(activePath)
	if err != nil {
		t.Fatalf("resilient ReadAll failed: %v", err)
	}
	wantModels := []string{"rot-before", "rot-after", "active-1", "active-2"}
	if len(recs) != len(wantModels) {
		t.Fatalf("expected %d records, got %d", len(wantModels), len(recs))
	}
	for i, want := range wantModels {
		if recs[i].Model != want {
			t.Errorf("record %d model = %q, want %q", i, recs[i].Model, want)
		}
	}

	// 2. Strict ReadAll rejects on the first corrupted line with a safe detail-free error.
	_, strictErr := store.ReadAllStrict(activePath)
	if strictErr == nil {
		t.Fatal("expected strict ReadAll to fail on corrupted line")
	}
	if !errors.Is(strictErr, store.ErrCorruptJSONL) {
		t.Fatalf("expected ErrCorruptJSONL, got %v", strictErr)
	}
	if strings.Contains(strictErr.Error(), activePath) || strings.Contains(strictErr.Error(), corruptedLine1) {
		t.Fatalf("strict error leaked path or content: %v", strictErr)
	}

	// 3. Store instance with resilient decoding: verifies Snapshot, sequence numbering, and health.
	st, err := store.Open(activePath, store.DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	snap, err := st.Snapshot()
	if err != nil {
		t.Fatalf("Snapshot failed: %v", err)
	}
	if len(snap.Records) != len(wantModels) {
		t.Fatalf("expected %d records in snapshot, got %d", len(wantModels), len(snap.Records))
	}
	for i, r := range snap.Records {
		if r.Sequence != uint64(i+1) {
			t.Errorf("record %d sequence = %d, want %d", i, r.Sequence, i+1)
		}
		if r.Model != wantModels[i] {
			t.Errorf("record %d model = %q, want %q", i, r.Model, wantModels[i])
		}
	}

	// Verify health counters and classifications:
	// - CorruptLines: corruptedLine1 (rot), corruptedLine2 (active), multiObjectLine (active), nonObjectLine (active), truncatedTail (active) = 5
	// - TruncatedLines: truncatedTail only (unterminated JSON object at EOF) = 1
	// - DecodeFailures: 1
	// - Degraded: true
	// - DecodeMode: "resilient"
	// - LastError: "corrupt_jsonl"
	h := st.Health()
	if h.DecodeMode != "resilient" {
		t.Errorf("DecodeMode = %q, want resilient", h.DecodeMode)
	}
	if !h.Degraded {
		t.Error("expected Degraded = true")
	}
	if h.DecodeFailures != 1 {
		t.Errorf("DecodeFailures = %d, want 1", h.DecodeFailures)
	}
	if h.CorruptLines != 5 {
		t.Errorf("CorruptLines = %d, want 5", h.CorruptLines)
	}
	if h.TruncatedLines != 1 {
		t.Errorf("TruncatedLines = %d, want 1", h.TruncatedLines)
	}
	if h.LastError != "corrupt_jsonl" {
		t.Errorf("LastError = %q, want corrupt_jsonl", h.LastError)
	}

	// 4. Cache hit must not recount corruption or increment decode failures.
	snap2, err := st.Snapshot()
	if err != nil {
		t.Fatalf("second snapshot failed: %v", err)
	}
	if len(snap2.Records) != len(wantModels) {
		t.Fatalf("second snapshot record count = %d, want %d", len(snap2.Records), len(wantModels))
	}
	h2 := st.Health()
	if h2.CacheHits != 1 {
		t.Errorf("CacheHits = %d, want 1", h2.CacheHits)
	}
	if h2.DecodeFailures != 1 || h2.CorruptLines != 5 || h2.TruncatedLines != 1 {
		t.Errorf("cache hit recounted corruption: %+v", h2)
	}

	// 5. Append new record separates from the preexisting unterminated tail.
	if err := st.Append(store.LogRecord{Model: "active-appended"}); err != nil {
		t.Fatalf("Append failed: %v", err)
	}
	snap3, err := st.Snapshot()
	if err != nil {
		t.Fatalf("third snapshot failed: %v", err)
	}
	wantAfterAppend := append(wantModels, "active-appended")
	if len(snap3.Records) != len(wantAfterAppend) {
		t.Fatalf("after append record count = %d, want %d", len(snap3.Records), len(wantAfterAppend))
	}
	for i, want := range wantAfterAppend {
		if snap3.Records[i].Model != want {
			t.Errorf("after append record %d model = %q, want %q", i, snap3.Records[i].Model, want)
		}
		if snap3.Records[i].Sequence != uint64(i+1) {
			t.Errorf("after append record %d sequence = %d, want %d", i, snap3.Records[i].Sequence, i+1)
		}
	}
}
