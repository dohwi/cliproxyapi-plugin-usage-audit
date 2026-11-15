package store

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenScrubsLegacySecretFieldsFromHistory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "usage.jsonl")
	active := "{\"model\":\"m1\",\"source\":\"sk-upstream-secret\",\"total_tokens\":3}\r\n" +
		"not json at all\n" +
		"\n" +
		"{\"model\":\"m2\",\"executor_type\":\"codex\"}\n" +
		"{\"model\":\"m3\",\"api_key_id\":\"sk-client-secret\"}"
	rotated := "{\"model\":\"m0\",\"source\":\"user@example.com\"}\n"
	clean := "{\"model\":\"untouched\"}\n"
	if err := os.WriteFile(path, []byte(active), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".1", []byte(rotated), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path+".2", []byte(clean), 0o600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.Stat(path + ".2")

	st, err := Open(path, DefaultMaxBytes, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	got, _ := os.ReadFile(path)
	want := "{\"model\":\"m1\",\"total_tokens\":3}\r\n" +
		"not json at all\n" +
		"\n" +
		"{\"model\":\"m2\",\"executor_type\":\"codex\"}\n" +
		"{\"model\":\"m3\"}"
	if string(got) != want {
		t.Fatalf("active log after scrub:\n%q\nwant\n%q", got, want)
	}
	if r, _ := os.ReadFile(path + ".1"); bytes.Contains(r, []byte("source")) || !bytes.Contains(r, []byte("m0")) {
		t.Fatalf("rotation not scrubbed: %s", r)
	}
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o640 && os.PathSeparator == '/' {
		t.Fatalf("permissions changed: %v", fi.Mode().Perm())
	}
	after, _ := os.Stat(path + ".2")
	if !after.ModTime().Equal(before.ModTime()) {
		t.Fatal("clean file was rewritten")
	}

	snap, err := st.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range snap.Records {
		if r.Source != "" || r.APIKeyID != "" {
			t.Fatalf("secret field in snapshot: %+v", r)
		}
	}
	leftovers, _ := filepath.Glob(filepath.Join(dir, ".usage-scrub-*"))
	if len(leftovers) != 0 {
		t.Fatalf("temp files left: %v", leftovers)
	}
}

func TestAppendNeverPersistsSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	st, err := Open(path, DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.Append(LogRecord{Model: "m", Source: "sk-upstream-secret", ExecutorType: "codex"}); err != nil {
		t.Fatal(err)
	}
	snap, err := st.Snapshot()
	if err != nil || len(snap.Records) != 1 || snap.Records[0].Source != "" || snap.Records[0].ExecutorType != "codex" {
		t.Fatalf("snapshot: %+v %v", snap.Records, err)
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if bytes.Contains(data, []byte("sk-upstream-secret")) || bytes.Contains(data, []byte(`"source"`)) {
		t.Fatalf("source persisted: %s", data)
	}
}
