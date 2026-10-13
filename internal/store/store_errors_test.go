package store

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFlushErrorsPropagate(t *testing.T) {
	for _, operation := range []string{"flush", "snapshot", "close", "periodic append", "rotation"} {
		t.Run(operation, func(t *testing.T) {
			s, err := Open(filepath.Join(t.TempDir(), "usage.jsonl"), DefaultMaxBytes, 2)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if err := s.Append(LogRecord{}); err != nil {
				t.Fatal(err)
			}
			// Fault injection: buffered writes cannot reach a closed descriptor.
			if err := s.f.Close(); err != nil {
				t.Fatal(err)
			}
			switch operation {
			case "flush":
				err = s.Flush()
			case "snapshot":
				_, err = s.ReadAll()
			case "close":
				err = s.Close()
			case "periodic append":
				s.lineCount = defaultFlushEvery - 1
				err = s.Append(LogRecord{})
			case "rotation":
				s.maxBytes = 1
				err = s.Append(LogRecord{})
			}
			if !errors.Is(err, os.ErrClosed) {
				t.Fatalf("error = %v, want closed descriptor error", err)
			}
		})
	}
}

func TestRotationFailureIsReturnedAndWriterRecovers(t *testing.T) {
	for _, operation := range []string{"remove oldest", "rename active"} {
		t.Run(operation, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "usage.jsonl")
			s, err := Open(path, 1, 2)
			if err != nil {
				t.Fatal(err)
			}
			defer s.Close()
			if err := s.Append(LogRecord{Model: "before"}); err != nil {
				t.Fatal(err)
			}
			if operation == "remove oldest" {
				if err := os.Mkdir(path+".2", 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(path+".2", "block"), nil, 0o600); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Rename(path, path+".backup"); err != nil {
					t.Fatal(err)
				}
			}
			if err := s.Append(LogRecord{Model: "must-not-write"}); err == nil {
				t.Fatal("rotation failure ignored")
			}
			if operation == "remove oldest" {
				if err := os.RemoveAll(path + ".2"); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := os.Rename(path+".backup", path); err != nil {
					t.Fatal(err)
				}
			}
			if err := s.Append(LogRecord{Model: "after"}); err != nil {
				t.Fatal(err)
			}
			recs, err := s.ReadAll()
			if err != nil || len(recs) != 2 || recs[0].Model != "before" || recs[1].Model != "after" {
				t.Fatalf("recovery: %+v, %v", recs, err)
			}
		})
	}
}

func TestClearDiscardsBufferedRecordsAndReportsRemovalFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "usage.jsonl")
	s, err := Open(path, DefaultMaxBytes, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Append(LogRecord{Model: "buffered-before-clear"}); err != nil {
		t.Fatal(err)
	}
	if s.bw.Buffered() == 0 {
		t.Fatal("test requires pending buffered data")
	}
	if err := s.Clear(); err != nil {
		t.Fatal(err)
	}
	if err := s.Append(LogRecord{Model: "after-clear"}); err != nil {
		t.Fatal(err)
	}
	recs, err := s.ReadAll()
	if err != nil || len(recs) != 1 || recs[0].Model != "after-clear" {
		t.Fatalf("buffer resurrected: %+v, %v", recs, err)
	}
	// A failed unlink must not be reported as a successful full deletion, and
	// must not prevent recreation of the active writer.
	if err := os.Mkdir(path+".1", 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path+".1", "block"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.Clear(); err == nil {
		t.Fatal("clear removal error ignored")
	}
	if err := os.RemoveAll(path + ".1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Append(LogRecord{Model: "recovered"}); err != nil {
		t.Fatal(err)
	}
	recs, err = s.ReadAll()
	if err != nil || len(recs) != 1 || recs[0].Model != "recovered" {
		t.Fatalf("clear recovery: %+v, %v", recs, err)
	}
}
