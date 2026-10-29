package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSnapshotCacheCopyInvalidationAndLegacySequence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	if err := os.WriteFile(path, []byte("{\"model\":\"legacy\",\"api_key_id\":\"secret\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := Open(path, DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	first, err := st.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if first.Records[0].Sequence != 1 || first.Records[0].APIKeyID != "" {
		t.Fatal("legacy normalization failed")
	}
	first.Records[0].Model = "mutated"
	second, err := st.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if second.Records[0].Model != "legacy" || st.Health().SnapshotDecodes != 1 || st.Health().CacheHits != 1 {
		t.Fatal("cache copied unsafely or decoded twice")
	}
	if err := st.Append(LogRecord{Model: "append"}); err != nil {
		t.Fatal(err)
	}
	third, err := st.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if third.Generation != second.Generation || third.Boundary != 2 || third.Records[0].Sequence != 1 || third.Records[1].Sequence != 2 {
		t.Fatal("append changed generation or sequence")
	}
	if err := st.Flush(); err != nil {
		t.Fatal(err)
	}
	fourth, err := st.Snapshot()
	if err != nil || fourth.Generation != third.Generation {
		t.Fatal("flush expires cursor", err)
	}
	// An out-of-band replacement with the same byte length must invalidate too.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	changed := strings.Replace(string(data), "legacy", "edited", 1)
	if err := os.WriteFile(path, []byte(changed), 0o600); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatal(err)
	}
	fifth, err := st.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if fifth.Generation == fourth.Generation || fifth.Records[0].Model != "edited" {
		t.Fatal("external edit not detected")
	}
	if err := st.Clear(); err != nil {
		t.Fatal(err)
	}
	sixth, err := st.Snapshot()
	if err != nil || sixth.Generation == fifth.Generation || sixth.Boundary != 0 {
		t.Fatal("clear cache not invalidated", err)
	}
}
func TestSnapshotCacheMemoryCap(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "usage.jsonl"), DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	// Exercise accounting directly with strings larger than the retained-cache cap.
	old := maxSnapshotCacheBytes
	maxSnapshotCacheBytes = 1 << 20
	t.Cleanup(func() { maxSnapshotCacheBytes = old })
	if err := st.Append(LogRecord{Model: strings.Repeat("x", int(maxSnapshotCacheBytes)+1)}); err != nil {
		t.Fatal(err)
	}
	snap, err := st.Snapshot()
	if err != nil || len(snap.Records) != 1 {
		t.Fatal(err)
	}
	if st.cacheValid || st.cache != nil {
		t.Fatal("oversize snapshot retained")
	}
}

func rotationFixture(t *testing.T) *Store {
	t.Helper()
	st, err := Open(filepath.Join(t.TempDir(), "usage.jsonl"), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	for i := 0; i < 3; i++ {
		if err := st.Append(LogRecord{Model: fmt.Sprint(i)}); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.Flush(); err != nil {
		t.Fatal(err)
	}
	return st
}
func assertRotationUnchanged(t *testing.T, st *Store) {
	t.Helper()
	recs, err := st.ReadAll()
	if err != nil || len(recs) != 3 {
		t.Fatalf("lost history: %+v %v", recs, err)
	}
	for i := range recs {
		if recs[i].Model != fmt.Sprint(i) {
			t.Fatalf("history shifted: %+v", recs)
		}
	}
	entries, err := os.ReadDir(filepath.Dir(st.path))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".usage-") {
			t.Fatalf("leaked transaction %s", entry.Name())
		}
	}
}
func TestStagedRotationRollbackEveryRename(t *testing.T) {
	// 3 moves into staging, 2 retained-file installs, 1 new-writer install.
	for failAt := 1; failAt <= 6; failAt++ {
		t.Run(fmt.Sprint(failAt), func(t *testing.T) {
			st := rotationFixture(t)
			calls := 0
			st.rename = func(a, b string) error {
				calls++
				if calls == failAt {
					return errors.New("injected rename")
				}
				return os.Rename(a, b)
			}
			if err := st.Append(LogRecord{Model: "not-written"}); err == nil {
				t.Fatal("failure ignored")
			}
			st.rename = os.Rename
			assertRotationUnchanged(t, st)
			if st.Health().RotationFailures != 1 || st.Health().WriteFailures != 1 {
				t.Fatalf("health %+v", st.Health())
			}
			if err := st.Append(LogRecord{Model: "recovered"}); err != nil {
				t.Fatal(err)
			}
			recs, err := st.ReadAll()
			if err != nil || len(recs) != 3 || recs[0].Model != "1" || recs[2].Model != "recovered" {
				t.Fatalf("bad resume: %+v %v", recs, err)
			}
		})
	}
}
func TestStagedRotationPrepareAndRetentionFailures(t *testing.T) {
	for _, failure := range []string{"prepare", "retention"} {
		t.Run(failure, func(t *testing.T) {
			st := rotationFixture(t)
			if failure == "prepare" {
				st.createTemp = func(string, string) (*os.File, error) { return nil, errors.New("injected create") }
			} else {
				st.remove = func(string) error { return errors.New("injected remove") }
			}
			if err := st.Append(LogRecord{}); err == nil {
				t.Fatal("failure ignored")
			}
			st.createTemp, st.remove = os.CreateTemp, os.Remove
			assertRotationUnchanged(t, st)
		})
	}
}
func TestRotationRollbackFailureFailsClosed(t *testing.T) {
	st := rotationFixture(t)
	calls := 0
	st.rename = func(a, b string) error {
		calls++
		if calls >= 2 {
			return errors.New("persistent rename fault")
		}
		return os.Rename(a, b)
	}
	if err := st.Append(LogRecord{}); err == nil || !st.closed {
		t.Fatal("rollback failure did not fail closed", err)
	}
	if err := st.Append(LogRecord{}); !errors.Is(err, os.ErrClosed) {
		t.Fatal("unsafe writer still active", err)
	}
	// Staged active history remains available for manual recovery.
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(st.path), ".usage-rotation-*", "0"))
	if err != nil || len(matches) != 1 {
		t.Fatal("staged recovery file lost", matches, err)
	}
}

// TestHealthRecoversWhenCorruptionIsRepaired verifies Degraded does not stay
// latched forever: once a full decode finds no corrupt line (rotated away or
// repaired), the store reports healthy again while keeping the history counters.
func TestHealthRecoversWhenCorruptionIsRepaired(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "usage.jsonl"), DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := os.WriteFile(st.path, []byte("garbage\n{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Snapshot(); err != nil {
		t.Fatal(err)
	}
	if h := st.Health(); !h.Degraded || h.CorruptLines != 1 {
		t.Fatalf("expected degraded after corruption: %+v", h)
	}
	if err := os.WriteFile(st.path, []byte("{}\n{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	snap, err := st.Snapshot()
	if err != nil || len(snap.Records) != 2 {
		t.Fatalf("repaired snapshot: %+v %v", snap, err)
	}
	h := st.Health()
	if h.Degraded || h.LastError != "" || !h.LastErrorAt.IsZero() {
		t.Fatalf("still degraded after repair: %+v", h)
	}
	if h.CorruptLines != 1 || h.DecodeFailures != 1 {
		t.Fatalf("history counters must be preserved: %+v", h)
	}
}

func TestHealthSanitizesDecodeAndFlushFailures(t *testing.T) {

	st, err := Open(filepath.Join(t.TempDir(), "usage.jsonl"), DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	secret := "secret-invalid-must-not-surface"
	if err := os.WriteFile(st.path, []byte(secret+"\n{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	snap, err := st.Snapshot()
	if err != nil || len(snap.Records) != 1 {
		t.Fatalf("resilient snapshot: %+v %v", snap, err)
	}
	if h := st.Health(); h.DecodeFailures != 1 || h.CorruptLines != 1 || h.ReadFailures != 0 || !h.Degraded || h.DecodeMode != "resilient" || h.LastError != "corrupt_jsonl" || strings.Contains(h.LastError, "secret") || h.LastErrorAt.IsZero() {
		t.Fatalf("health: %+v", h)
	}
	// Cache reuse must not recount the same physical corruption.
	if _, err := st.Snapshot(); err != nil {
		t.Fatal(err)
	}
	if h := st.Health(); h.DecodeFailures != 1 || h.CorruptLines != 1 || h.CacheHits != 1 {
		t.Fatalf("cache recounted corruption: %+v", h)
	}
	if err := st.Clear(); err != nil {
		t.Fatal(err)
	}
	if err := st.Append(LogRecord{}); err != nil {
		t.Fatal(err)
	}
	if err := st.f.Close(); err != nil {
		t.Fatal(err)
	}
	if err := st.Flush(); err == nil {
		t.Fatal("flush accepted")
	}
	if h := st.Health(); h.FlushFailures != 1 || h.LastError != "flush" {
		t.Fatalf("health: %+v", h)
	}
}

func TestStrictStoreRejectsCorruptionWithSafeCategory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	secret := "secret-corrupt-payload"
	if err := os.WriteFile(path, []byte(secret+"\n{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := OpenWithOptions(path, DefaultMaxBytes, 2, Options{StrictDecoding: true})
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if _, err := st.Snapshot(); !errors.Is(err, ErrCorruptJSONL) || strings.Contains(err.Error(), secret) || strings.Contains(err.Error(), path) {
		t.Fatalf("strict error unsafe or missing: %v", err)
	}
	if h := st.Health(); h.DecodeMode != "strict" || h.DecodeFailures != 1 || h.CorruptLines != 1 || h.ReadFailures != 1 || h.LastError != "corrupt_jsonl" {
		t.Fatalf("strict health: %+v", h)
	}
}

// Append must extend (not discard) the decoded cache, and read-only views taken
// before an append must stay intact while appends and reads run concurrently.
func TestSnapshotCacheSurvivesAppendAndViewsStayStable(t *testing.T) {
	st, err := Open(filepath.Join(t.TempDir(), "usage.jsonl"), DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	for i := 0; i < 5; i++ {
		if err := st.Append(LogRecord{Model: fmt.Sprint("m", i), TotalTokens: int64(i)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := st.SnapshotView(); err != nil {
		t.Fatal(err)
	}
	decodes := st.Health().SnapshotDecodes
	view, err := st.SnapshotView()
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			if err := st.Append(LogRecord{Model: "x", TotalTokens: 99}); err != nil {
				t.Error(err)
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 200; i++ {
			v, err := st.SnapshotView()
			if err != nil {
				t.Error(err)
				return
			}
			for j := range v.Records[:5] {
				if v.Records[j].Model != fmt.Sprint("m", j) || v.Records[j].Sequence != uint64(j+1) {
					t.Error("view mutated")
					return
				}
			}
		}
	}()
	wg.Wait()
	if len(view.Records) != 5 || view.Records[4].Model != "m4" {
		t.Fatal("earlier view changed")
	}
	final, err := st.Snapshot()
	if err != nil || len(final.Records) != 205 || final.Records[204].Sequence != 205 {
		t.Fatal("appended records missing", err, len(final.Records))
	}
	if got := st.Health().SnapshotDecodes; got != decodes {
		t.Fatalf("append forced %d re-decodes", got-decodes)
	}
	// The extended cache must equal a fresh disk decode.
	if err := st.Flush(); err != nil {
		t.Fatal(err)
	}
	disk, err := ReadAll(st.path)
	if err != nil || len(disk) != len(final.Records) {
		t.Fatal(err, len(disk))
	}
	for i := range disk {
		disk[i].Sequence, disk[i].APIKeyID = final.Records[i].Sequence, ""
		if disk[i] != final.Records[i] {
			t.Fatalf("cache diverges from disk at %d", i)
		}
	}
}
