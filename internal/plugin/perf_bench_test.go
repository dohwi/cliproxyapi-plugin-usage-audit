package plugin

import (
	"fmt"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/queue"
	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
)

// benchState seeds n synthetic records spread over the last 30 days.
func benchState(b *testing.B, n int) {
	b.Helper()
	st, err := store.Open(b.TempDir()+"/usage.jsonl", store.DefaultMaxBytes, store.DefaultMaxRotation)
	if err != nil {
		b.Fatal(err)
	}
	now := time.Now().UTC()
	providers := []string{"codex", "claude", "antigravity"}
	for i := 0; i < n; i++ {
		ts := now.Add(-time.Duration(i) * (30 * 24 * time.Hour / time.Duration(n)))
		if err := st.Append(store.LogRecord{
			T: ts, RequestedAt: ts, Provider: providers[i%3], ExecutorType: "exec", Source: "src",
			Model: fmt.Sprintf("model-%d", i%7), AuthIndex: fmt.Sprintf("auth-%d", i%12),
			Generate: true, InputTokens: 1000, OutputTokens: 200, CachedTokens: 300, TotalTokens: 1200, LatencyMs: 900,
		}); err != nil {
			b.Fatal(err)
		}
	}
	ps := &pluginState{st: st, q: queue.New(16)}
	stateMu.Lock()
	state = ps
	stateMu.Unlock()
	b.Cleanup(func() { stateMu.Lock(); state = nil; stateMu.Unlock(); ps.q.Close(); _ = st.Close() })
}

func benchQuery(from, to time.Time) map[string][]string {
	return map[string][]string{
		"from":      {from.Format(time.RFC3339)},
		"to":        {to.Format(time.RFC3339)},
		"time_zone": {"Asia/Seoul"},
	}
}

func BenchmarkResourceSummary(b *testing.B) {
	for _, n := range []int{50_000, 200_000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			benchState(b, n)
			req := managementHandleRequest{Method: "GET", Path: resourceBase + "/data/summary", Query: benchQuery(time.Now().Add(-24*time.Hour), time.Now())}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				handleGetResourceSummary(req)
			}
		})
	}
}

func BenchmarkResourceRecords(b *testing.B) {
	for _, n := range []int{50_000, 200_000} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			benchState(b, n)
			req := managementHandleRequest{Method: "GET", Path: resourceBase + "/data/records", Query: benchQuery(time.Now().Add(-24*time.Hour), time.Now())}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				handleGetResourceRecords(req)
			}
		})
	}
}
