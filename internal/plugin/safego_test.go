package plugin

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
)

// captureCrashLog replaces logPanic; the returned func waits for n entries and returns them.
func captureCrashLog(t *testing.T) func(n int) []string {
	t.Helper()
	var mu sync.Mutex
	var got []string
	prev := logPanic
	logPanic = func(name string, v any, stack []byte) {
		mu.Lock()
		defer mu.Unlock()
		got = append(got, name+": "+toString(v))
	}
	t.Cleanup(func() { logPanic = prev })
	return func(n int) []string {
		deadline := time.Now().Add(5 * time.Second)
		for {
			mu.Lock()
			out := append([]string(nil), got...)
			mu.Unlock()
			if len(out) >= n || time.Now().After(deadline) {
				return out
			}
			time.Sleep(2 * time.Millisecond)
		}
	}
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return "?"
}

// A panicking goroutine started through safeGo must not crash the test binary,
// its deferred cleanup must run, and the panic must be logged with its name.
func TestSafeGoRecoversAndRunsCleanup(t *testing.T) {
	logs := captureCrashLog(t)
	var mu sync.Mutex
	mu.Lock() // held by the goroutine below and released by its defer
	done := make(chan struct{})
	safeGo("boom", func() {
		defer close(done)
		defer mu.Unlock()
		panic("kaboom")
	})
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("goroutine did not finish")
	}
	if !mu.TryLock() {
		t.Fatal("deferred unlock did not run")
	}
	if got := logs(1); len(got) != 1 || got[0] != "boom: kaboom" {
		t.Fatalf("panic log = %q", got)
	}
	if !runRecovered("ok", func() {}) || runRecovered("bad", func() { panic("x") }) {
		t.Fatal("runRecovered must report whether fn returned normally")
	}
}

// A panic while storing one record is logged; the consumer keeps running and
// stores the next records, and shutdown still balances workerWg.
func TestConsumerSurvivesPanicAndKeepsRecording(t *testing.T) {
	logs := captureCrashLog(t)
	appendTestHook = func(r *store.LogRecord) {
		if r.Model == "poison" {
			panic("bad record")
		}
	}
	t.Cleanup(func() { appendTestHook = nil })
	st := testStore(t, store.DefaultMaxBytes)
	ps := installTestState(t, st)
	ps.workerWg.Add(1)
	safeGo("usage-audit consumer", ps.consumeLoop)

	handleUsage([]byte(`{"Model":"before"}`))
	handleUsage([]byte(`{"Model":"poison"}`))
	handleUsage([]byte(`{"Model":"after"}`))
	deadline := time.Now().Add(5 * time.Second)
	for {
		recs, err := st.ReadAll()
		if err == nil && len(recs) == 2 {
			if recs[0].Model != "before" || recs[1].Model != "after" {
				t.Fatalf("records = %+v", recs)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("consumer stopped after a panic: %+v %v", recs, err)
		}
		time.Sleep(5 * time.Millisecond)
	}
	if got := logs(1); len(got) != 1 || got[0] != "usage-audit consumer: bad record" {
		t.Fatalf("panic log = %q", got)
	}
	ps.q.Close()
	stopped := make(chan struct{})
	go func() { ps.workerWg.Wait(); close(stopped) }()
	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("workerWg did not balance after a recovered panic")
	}
}

// Accepted records reach the file within the flush interval even when fewer than
// the 256-line batch arrive (a quiet server).
func TestConsumerFlushesBufferedRecordsPeriodically(t *testing.T) {
	prev := storeFlushInterval
	storeFlushInterval = 20 * time.Millisecond
	t.Cleanup(func() { storeFlushInterval = prev })
	dir := t.TempDir()
	st, err := store.Open(dir+"/usage.jsonl", store.DefaultMaxBytes, 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	ps := installTestState(t, st)
	ps.workerWg.Add(1)
	safeGo("usage-audit consumer", ps.consumeLoop)
	defer func() { ps.q.Close(); ps.workerWg.Wait() }()

	handleUsage([]byte(`{"Model":"quiet"}`))
	deadline := time.Now().Add(5 * time.Second)
	for {
		// Read the file directly: Snapshot would flush by itself.
		b, _ := os.ReadFile(dir + "/usage.jsonl")
		if strings.Contains(string(b), `"model":"quiet"`) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("buffered record was not flushed by the periodic flush")
		}
		time.Sleep(5 * time.Millisecond)
	}
}
