package plugin

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
)

// referenceRecordsPage preserves the old full-filter/copy/sort implementation.
func referenceRecordsPage(records []store.LogRecord, opts store.FilterOptions, cursor *recordsCursor, limit int) (page []store.LogRecord, more, anchorFound bool) {
	filtered := store.FilterRecords(records, opts)
	page = make([]store.LogRecord, 0, limit)
	anchorFound = cursor == nil
	for _, r := range filtered {
		ts := r.EffectiveTimestamp()
		if cursor != nil {
			if r.Sequence == cursor.Sequence && ts.Equal(cursor.Timestamp) {
				anchorFound = true
			}
			if ts.After(cursor.Timestamp) || (ts.Equal(cursor.Timestamp) && r.Sequence >= cursor.Sequence) {
				continue
			}
		}
		if len(page) == limit {
			more = true
			break
		}
		page = append(page, r)
	}
	return
}

func syntheticPageRecords(n int, seed int64) []store.LogRecord {
	rng := rand.New(rand.NewSource(seed))
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	records := make([]store.LogRecord, n)
	for i := range records {
		ts := base.Add(time.Duration(rng.Intn(31)-15) * time.Minute)
		r := store.LogRecord{
			Sequence: uint64(i + 1), T: ts.Add(time.Hour), RequestedAt: ts,
			Provider: fmt.Sprint("provider-", i%3), Model: fmt.Sprint("model-", i%7),
			Source: fmt.Sprint("endpoint-", i%4), ExecutorType: fmt.Sprint("endpoint-", (i+1)%4),
			AuthIndex: fmt.Sprint("auth-", i%5), Failed: i%2 == 0,
			InputTokens: int64(i), TotalTokens: int64(i + 100),
		}
		if i%5 == 0 {
			r.T, r.RequestedAt = ts, time.Time{} // legacy timestamp fallback
		}
		if i%17 == 0 {
			r.T, r.RequestedAt = time.Time{}, time.Time{} // zero timestamp also sorts safely
		}
		records[i] = r
	}
	return records
}

func TestRecordsPageTopKMatchesReference(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	from, to := base.Add(-5*time.Minute), base.Add(5*time.Minute)
	filters := []store.FilterOptions{
		{}, {From: &from}, {To: &to}, {From: &from, To: &to},
		{Provider: "provider-1"}, {Model: "model-2"}, {Status: "success"}, {Status: "failed"},
		{Source: "endpoint-1"}, {ExecutorType: "endpoint-1"}, {Endpoint: "endpoint-1"}, {AuthIndex: "auth-2"},
		{Provider: "provider-1", Status: "failed", Endpoint: "endpoint-1", From: &from, To: &to},
		{Model: "no-matches"}, {From: &to, To: &from},
	}
	for _, seed := range []int64{1, 42, 8675309} {
		for _, n := range []int{0, 1, 19, 20, 21, 257} {
			records := syntheticPageRecords(n, seed)
			before := append([]store.LogRecord(nil), records...)
			for f, opts := range filters {
				for _, limit := range []int{1, 2, 20, 100, 500} {
					t.Run(fmt.Sprintf("seed%d/n%d/filter%d/limit%d", seed, n, f, limit), func(t *testing.T) {
						var cursor *recordsCursor
						for p := 0; p <= n; p++ {
							want, wantMore, wantAnchor := referenceRecordsPage(records, opts, cursor, limit)
							got, more, anchor := selectRecordsPage(records, opts, cursor, limit)
							if !reflect.DeepEqual(got, want) || more != wantMore || anchor != wantAnchor {
								t.Fatalf("page %d: sequences %v / more %v / anchor %v; want %v / %v / %v", p, pageSequences(got), more, anchor, pageSequences(want), wantMore, wantAnchor)
							}
							if !more {
								return
							}
							r := got[len(got)-1]
							cursor = &recordsCursor{Timestamp: r.EffectiveTimestamp(), Sequence: r.Sequence}
						}
						t.Fatal("pagination did not terminate")
					})
				}
			}
			if n > 0 && !reflect.DeepEqual(records, before) {
				t.Fatal("selection mutated the read-only snapshot")
			}
		}
	}
}

func pageSequences(records []store.LogRecord) []uint64 {
	sequences := make([]uint64, len(records))
	for i := range records {
		sequences[i] = records[i].Sequence
	}
	return sequences
}

func TestRecordsPageTopKTiesAndLateCandidates(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	records := make([]store.LogRecord, 100)
	for i := range records {
		records[i] = store.LogRecord{Sequence: uint64(i + 1), RequestedAt: base}
	}
	page, more, anchor := selectRecordsPage(records, store.FilterOptions{}, nil, 20)
	if !more || !anchor || page[0].Sequence != 100 || page[19].Sequence != 81 {
		t.Fatalf("reverse-stable timestamp ties: %v, more=%v anchor=%v", pageSequences(page), more, anchor)
	}
	// The newest request can occur at either end of the append frontier.
	for _, index := range []int{0, 99} {
		records[index].RequestedAt = base.Add(time.Hour)
		page, _, _ = selectRecordsPage(records, store.FilterOptions{}, nil, 1)
		if page[0].Sequence != records[index].Sequence {
			t.Fatalf("missed newest timestamp at append index %d", index)
		}
		records[index].RequestedAt = base
	}
}

func TestRecordsPageTopKAnchorMustMatchFilters(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	st := testStore(t, store.DefaultMaxBytes)
	for i := 0; i < 6; i++ {
		model := "included"
		if i == 5 {
			model = "excluded"
		}
		appendTestRecord(t, st, store.LogRecord{Model: model, RequestedAt: base.Add(time.Duration(i) * time.Minute)})
	}
	snap, err := st.SnapshotView()
	if err != nil {
		t.Fatal(err)
	}
	opts := store.FilterOptions{Model: "included"}
	for _, scope := range []string{"management", "resource"} {
		for _, tc := range []struct {
			name     string
			sequence uint64
			ts       time.Time
		}{
			{"filtered-anchor", 6, base.Add(5 * time.Minute)},
			{"wrong-timestamp", 5, base.Add(10 * time.Minute)},
		} {
			t.Run(scope+"/"+tc.name, func(t *testing.T) {
				cursor := recordsCursor{1, filterFingerprint(opts, scope), snap.Generation, snap.Boundary, tc.ts, tc.sequence}
				page, next, err := recordsPage(st, opts, scope, encodeRecordsCursor(cursor), 1)
				if err == nil || page != nil || next != "" {
					t.Fatalf("invalid anchor accepted: page=%v next=%q err=%v", pageSequences(page), next, err)
				}
				status, message := recordsPageFailure(err)
				if status != 400 || message != "cursor anchor not found; restart pagination" {
					t.Fatalf("changed anchor validation: %d %s", status, message)
				}
			})
		}
	}
}

func TestRecordsPageTopKCursorReferenceWithBackdatedAppends(t *testing.T) {
	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	for _, scope := range []string{"management", "resource"} {
		for _, opts := range []store.FilterOptions{{}, {Provider: "provider-1"}} {
			t.Run(scope+"/"+opts.Provider, func(t *testing.T) {
				st := testStore(t, store.DefaultMaxBytes)
				for _, r := range syntheticPageRecords(120, 42) {
					appendTestRecord(t, st, r)
				}
				snap, err := st.SnapshotView()
				if err != nil {
					t.Fatal(err)
				}
				var value string
				var cursor *recordsCursor
				for p := 0; p <= 120; p++ {
					want, more, anchor := referenceRecordsPage(snap.Records, opts, cursor, 7)
					if !anchor {
						t.Fatal("reference lost its anchor")
					}
					wantNext := ""
					if more {
						r := want[len(want)-1]
						cursor = &recordsCursor{1, filterFingerprint(opts, scope), snap.Generation, snap.Boundary, r.EffectiveTimestamp().UTC(), r.Sequence}
						wantNext = encodeRecordsCursor(*cursor)
					}
					got, next, err := recordsPage(st, opts, scope, value, 7)
					if err != nil || !reflect.DeepEqual(got, want) || next != wantNext {
						t.Fatalf("page %d: got %v next=%q err=%v; want %v next=%q", p, pageSequences(got), next, err, pageSequences(want), wantNext)
					}
					if next == "" {
						return
					}
					// All three would be eligible by timestamp somewhere in the
					// traversal, but none may enter the original snapshot frontier.
					for _, ts := range []time.Time{base.Add(-24 * time.Hour), cursor.Timestamp, base.Add(24 * time.Hour)} {
						appendTestRecord(t, st, store.LogRecord{Provider: "provider-1", Model: "late", RequestedAt: ts})
					}
					value = next
				}
				t.Fatal("pagination did not terminate")
			})
		}
	}
}

var recordsPageBenchmarkSink []store.LogRecord
var recordsPageBenchmarkMore bool

func BenchmarkRecordsPageChronological(b *testing.B) {
	for _, n := range []int{50_000, 200_000} {
		records := syntheticPageRecords(n, 42)
		base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
		for i := range records {
			records[i].RequestedAt = base.Add(time.Duration(i) * time.Second)
		}
		opts := store.FilterOptions{}
		b.Run(fmt.Sprintf("%d/FullFilter", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				recordsPageBenchmarkSink, recordsPageBenchmarkMore, _ = referenceRecordsPage(records, opts, nil, 20)
			}
		})
		b.Run(fmt.Sprintf("%d/TopK", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				recordsPageBenchmarkSink, recordsPageBenchmarkMore, _ = selectRecordsPage(records, opts, nil, 20)
			}
		})
	}
}

// Benchmarks operate only on synthetic in-memory records (no store or IO).
// FullFilter is the previous full-copy/sort path followed by taking 20 records.
func BenchmarkRecordsPageTopK(b *testing.B) {
	for _, n := range []int{50_000, 200_000} {
		records := syntheticPageRecords(n, 42)
		for _, filtered := range []bool{false, true} {
			opts := store.FilterOptions{}
			if filtered {
				from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
				opts = store.FilterOptions{Provider: "provider-1", From: &from}
			}
			name := fmt.Sprintf("%d/filtered=%v", n, filtered)
			b.Run(name+"/FullFilter", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					page, more, _ := referenceRecordsPage(records, opts, nil, 20)
					recordsPageBenchmarkSink, recordsPageBenchmarkMore = page, more
				}
			})
			b.Run(name+"/TopK", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					page, more, _ := selectRecordsPage(records, opts, nil, 20)
					recordsPageBenchmarkSink, recordsPageBenchmarkMore = page, more
				}
			})
		}
	}
}
