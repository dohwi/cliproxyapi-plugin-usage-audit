package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// FlushPending writes buffered records to the file only when there are any.
func TestFlushPendingWritesOnlyBufferedBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	st, err := Open(path, DefaultMaxBytes, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if flushed, err := st.FlushPending(); flushed || err != nil {
		t.Fatalf("empty buffer: flushed=%v err=%v", flushed, err)
	}
	if err := st.Append(LogRecord{T: time.Now().UTC(), Model: "m"}); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); len(b) != 0 {
		t.Fatalf("record reached the file before a flush: %q", b)
	}
	if flushed, err := st.FlushPending(); !flushed || err != nil {
		t.Fatalf("buffered record: flushed=%v err=%v", flushed, err)
	}
	if b, _ := os.ReadFile(path); !strings.Contains(string(b), `"model":"m"`) {
		t.Fatalf("file after flush = %q", b)
	}
	if flushed, _ := st.FlushPending(); flushed {
		t.Fatal("second flush with nothing buffered")
	}
	// Our own flush must not look like an external edit (no new generation).
	v1, _ := st.Version()
	v2, _ := st.Version()
	if v1 != v2 {
		t.Fatalf("version changed without a write: %s → %s", v1, v2)
	}
	_ = st.Close()
	if flushed, err := st.FlushPending(); flushed || err != nil {
		t.Fatalf("closed store: flushed=%v err=%v", flushed, err)
	}
}

// Version changes on every append, on clear and on out-of-band edits.
func TestVersionTracksAppendsClearAndExternalEdits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	st, err := Open(path, DefaultMaxBytes, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	seen := map[string]bool{}
	next := func(what string) {
		t.Helper()
		v, err := st.Version()
		if err != nil {
			t.Fatal(err)
		}
		if seen[v] {
			t.Fatalf("%s: version %s repeated", what, v)
		}
		seen[v] = true
	}
	next("open")
	_ = st.Append(LogRecord{T: time.Now().UTC()})
	next("append")
	if err := st.Flush(); err != nil {
		t.Fatal(err)
	}
	if v, _ := st.Version(); !seen[v] {
		t.Fatal("own flush changed the version")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.WriteString(`{"t":"2026-01-01T00:00:00Z","model":"ext"}` + "\n")
	_ = f.Close()
	next("external edit")
	if err := st.Clear(); err != nil {
		t.Fatal(err)
	}
	next("clear")
}
