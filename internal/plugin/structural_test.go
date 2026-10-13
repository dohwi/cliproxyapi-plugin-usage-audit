package plugin

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/queue"
	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
)

func testStore(t *testing.T, maxBytes int64) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "usage.jsonl"), maxBytes, 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}
func appendTestRecord(t *testing.T, st *store.Store, r store.LogRecord) {
	t.Helper()
	if err := st.Append(r); err != nil {
		t.Fatal(err)
	}
}
func TestKeysetSnapshotWithLiveBackdatedAndTiedAppends(t *testing.T) {
	for _, scope := range []string{"management", "resource"} {
		t.Run(scope, func(t *testing.T) {
			st := testStore(t, store.DefaultMaxBytes)
			base := time.Now().UTC().Add(-time.Hour)
			for i := 0; i < 30; i++ {
				appendTestRecord(t, st, store.LogRecord{Model: fmt.Sprint(i), RequestedAt: base.Add(time.Duration(i/3) * time.Second)})
			}
			expected, _ := st.ReadAll()
			expected = store.FilterRecords(expected, store.FilterOptions{})
			var got []store.LogRecord
			cursor := ""
			for i := 0; i < 20; i++ {
				page, next, err := recordsPage(st, store.FilterOptions{}, scope, cursor, 4)
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, page...)
				if next == "" {
					break
				}
				// Late completions, ties, and new timestamps must all remain outside the frontier.
				for _, ts := range []time.Time{base.Add(-time.Hour), base.Add(3 * time.Second), base.Add(time.Hour)} {
					appendTestRecord(t, st, store.LogRecord{Model: "new", RequestedAt: ts})
				}
				cursor = next
			}
			if len(got) != len(expected) {
				t.Fatalf("got %d want %d", len(got), len(expected))
			}
			for i := range got {
				if got[i].Model != expected[i].Model {
					t.Fatalf("position %d: %s want %s", i, got[i].Model, expected[i].Model)
				}
			}
		})
	}
}
func TestCursorStrictValidationAndExpiry(t *testing.T) {
	st := testStore(t, store.DefaultMaxBytes)
	for i := 0; i < 3; i++ {
		appendTestRecord(t, st, store.LogRecord{Model: "a", T: time.Now().UTC()})
	}
	_, cursor, err := recordsPage(st, store.FilterOptions{}, "resource", "", 1)
	if err != nil || cursor == "" {
		t.Fatal(cursor, err)
	}
	fp := filterFingerprint(store.FilterOptions{}, "resource")
	original, err := decodeRecordsCursor(cursor, fp)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{base64.URLEncoding.EncodeToString([]byte("2")), strings.Repeat("a", 2049), "broken", cursor[:len(cursor)-5] + "aaaaa"} {
		if _, _, err := recordsPage(st, store.FilterOptions{}, "resource", value, 1); err == nil {
			t.Fatalf("accepted %q", value)
		}
	}
	if _, _, err := recordsPage(st, store.FilterOptions{Model: "b"}, "resource", cursor, 1); err == nil {
		t.Fatal("changed filters accepted")
	}
	if _, _, err := recordsPage(st, store.FilterOptions{}, "management", cursor, 1); err == nil {
		t.Fatal("cross scope accepted")
	}
	for _, mutate := range []func(*recordsCursor){
		func(c *recordsCursor) { c.Version = 2 }, func(c *recordsCursor) { c.Sequence = 0 }, func(c *recordsCursor) { c.Sequence = c.Boundary + 1 }, func(c *recordsCursor) { c.Timestamp = c.Timestamp.Add(time.Hour) },
	} {
		c := *original
		mutate(&c)
		if _, _, err := recordsPage(st, store.FilterOptions{}, "resource", encodeRecordsCursor(c), 1); err == nil {
			t.Fatal("invalid signed payload accepted")
		}
	}
	if err := st.Clear(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		appendTestRecord(t, st, store.LogRecord{})
	}
	if _, _, err := recordsPage(st, store.FilterOptions{}, "resource", cursor, 1); err == nil {
		t.Fatal("cleared cursor accepted")
	} else if status, _ := recordsPageFailure(err); status != 409 {
		t.Fatal(err)
	}
}

func TestCursorExpiresOnRotation(t *testing.T) {
	st := testStore(t, 1024)
	for i := 0; i < 2; i++ {
		appendTestRecord(t, st, store.LogRecord{T: time.Now().UTC()})
	}
	_, cursor, err := recordsPage(st, store.FilterOptions{}, "resource", "", 1)
	if err != nil || cursor == "" {
		t.Fatal(cursor, err)
	}
	for i := 0; i < 10; i++ {
		appendTestRecord(t, st, store.LogRecord{})
	}
	if _, _, err := recordsPage(st, store.FilterOptions{}, "resource", cursor, 1); err == nil {
		t.Fatal("rotated cursor accepted")
	}
}

func installTestState(t *testing.T, st *store.Store) *pluginState {
	t.Helper()
	ps := &pluginState{st: st, q: queue.New(2048)}
	stateMu.Lock()
	if state != nil {
		stateMu.Unlock()
		t.Fatal("unexpected registered state")
	}
	state = ps
	stateMu.Unlock()
	t.Cleanup(func() { stateMu.Lock(); state = nil; stateMu.Unlock(); ps.q.Close() })
	return ps
}
func responseForTest(t *testing.T, raw []byte) managementResponse {
	t.Helper()
	var env envelope
	var mr managementResponse
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(env.Result, &mr); err != nil {
		t.Fatal(err)
	}
	return mr
}
func TestDeleteEpochRejectsQueuedAndAlreadyDequeuedRecords(t *testing.T) {
	st := testStore(t, store.DefaultMaxBytes)
	ps := installTestState(t, st)
	b, _ := json.Marshal(queuedUsage{LogRecord: store.LogRecord{Model: "old"}, Epoch: 0})
	// Simulate an already-dequeued entry waiting to commit, plus a full backlog.
	for i := 0; i < 100; i++ {
		ps.q.Enqueue(b)
	}
	resp := responseForTest(t, handleDeleteRecords())
	if resp.StatusCode != 200 {
		t.Fatal(string(resp.Body))
	}
	ps.appendRecord(b)
	ps.drainRemaining()
	records, err := st.ReadAll()
	if err != nil || len(records) != 0 {
		t.Fatalf("resurrected %d: %v", len(records), err)
	}
	handleUsage([]byte(`{"Model":"new"}`))
	ps.drainRemaining()
	records, err = st.ReadAll()
	if err != nil || len(records) != 1 || records[0].Model != "new" {
		t.Fatalf("new epoch: %+v %v", records, err)
	}
}
func TestConcurrentObserverConsumerAndDeleteEpoch(t *testing.T) {
	st := testStore(t, store.DefaultMaxBytes)
	ps := installTestState(t, st)
	ps.workerWg.Add(1)
	go ps.consumeLoop()
	// Ensure cleanup closes worker before store on any test failure.
	defer ps.shutdown()
	var wg sync.WaitGroup
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 300; i++ {
				handleUsage([]byte(`{"Model":"before"}`))
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 40; i++ {
			handleDeleteRecords()
		}
	}()
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("epoch deadlock")
	}
	if resp := responseForTest(t, handleDeleteRecords()); resp.StatusCode != 200 {
		t.Fatal(string(resp.Body))
	}
	handleUsage([]byte(`{"Model":"after"}`))
	// Stop consumer without closing store to inspect all drained records.
	ps.q.Close()
	stopped := make(chan struct{})
	go func() { ps.workerWg.Wait(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("consumer did not stop")
	}
	records, err := st.ReadAll()
	if err != nil || len(records) != 1 || records[0].Model != "after" {
		t.Fatalf("resurrection: %+v %v", records, err)
	}
}

func TestResourceBoundsCacheReuseAndSafeHealth(t *testing.T) {
	st := testStore(t, store.DefaultMaxBytes)
	installTestState(t, st)
	now := time.Now().UTC()
	appendTestRecord(t, st, store.LogRecord{T: now})
	appendTestRecord(t, st, store.LogRecord{T: now.Add(-31 * 24 * time.Hour)})
	req := managementHandleRequest{Path: resourceBase + "/data/summary"}
	resp := responseForTest(t, handleGetResourceSummary(req))
	if resp.StatusCode != 200 || !bytes.Contains(resp.Body, []byte(`"total_count":1`)) {
		t.Fatal(string(resp.Body))
	}
	recordsReq := managementHandleRequest{Path: resourceBase + "/data/records"}
	if r := responseForTest(t, handleGetResourceRecords(recordsReq)); r.StatusCode != 200 {
		t.Fatal(string(r.Body))
	}
	h := st.Health()
	if h.SnapshotDecodes != 1 || h.CacheHits != 1 {
		t.Fatalf("duplicate decode: %+v", h)
	}
	for _, query := range []map[string][]string{
		{"from": {now.Add(-367 * 24 * time.Hour).Format(time.RFC3339Nano)}, "to": {now.Format(time.RFC3339Nano)}},
		{"from": {"2025-01-02"}, "to": {"2025-01-01"}},
	} {
		req.Query = query
		if r := responseForTest(t, handleGetResourceSummary(req)); r.StatusCode != 400 {
			t.Fatalf("unbounded query accepted %s", r.Body)
		}
	}
	for _, days := range []int{0, 30} {
		req.Query = map[string][]string{"from": {now.AddDate(0, 0, -days).Format("2006-01-02")}, "to": {now.Format("2006-01-02")}}
		if r := responseForTest(t, handleGetResourceSummary(req)); r.StatusCode != 200 {
			t.Fatalf("dashboard range broken %s", r.Body)
		}
	}
	recordsReq.Query = map[string][]string{"limit": {"101"}}
	if r := responseForTest(t, handleGetResourceRecords(recordsReq)); r.StatusCode != 400 {
		t.Fatal("limit accepted")
	}
	for i := 0; i < resourceSummaryMaxBuckets; i++ {
		appendTestRecord(t, st, store.LogRecord{T: now, Model: fmt.Sprint(i)})
	}
	req.Query = nil
	if r := responseForTest(t, handleGetResourceSummary(req)); r.StatusCode != 400 {
		t.Fatal("bucket overflow accepted")
	}
}
