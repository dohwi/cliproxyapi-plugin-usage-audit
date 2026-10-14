package store

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Rotation is a same-directory staged transaction for ordinary filesystem
// errors, not a power-loss transaction. No retained file is unlinked until all
// renames and the replacement writer are ready. Reverse moves restore the exact
// previous namespace on failure. A failed rollback closes the store (fail closed)
// and leaves staged files for operator recovery instead of overwriting them.
func (s *Store) rotateLocked() (err error) {
	defer func() {
		if err != nil {
			s.health.RotationFailures++
			s.failureLocked("rotation")
		}
	}()
	if err = s.flushLocked(); err != nil {
		return err
	}
	s.newGenerationLocked()
	replacement, err := s.createTemp(filepath.Dir(s.path), ".usage-new-*")
	if err != nil {
		return fmt.Errorf("store: prepare writer: %w", err)
	}
	replacementName := replacement.Name()
	committed := false
	defer func() {
		if !committed {
			_ = replacement.Close()
		}
		_ = os.Remove(replacementName)
	}()
	if err = replacement.Chmod(0o640); err != nil {
		return err
	}
	// CreateTemp uses O_RDWR, not O_APPEND. Acquire the final append-only
	// descriptor before touching the retained namespace.
	appendWriter, err := openAppend(replacementName, 0, false)
	if err != nil {
		return err
	}
	if err = replacement.Close(); err != nil {
		_ = appendWriter.Close()
		return err
	}
	replacement = appendWriter
	stage, err := os.MkdirTemp(filepath.Dir(s.path), ".usage-rotation-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(stage) }()
	type move struct{ from, to string }
	var moves []move
	moveFile := func(from, to string) error {
		if e := s.rename(from, to); e != nil {
			return e
		}
		moves = append(moves, move{from, to})
		return nil
	}
	rollback := func(cause error) error {
		var rollbackErr error
		for i := len(moves) - 1; i >= 0; i-- {
			if e := s.rename(moves[i].to, moves[i].from); e != nil {
				rollbackErr = errors.Join(rollbackErr, e)
			}
		}
		if rollbackErr != nil {
			s.closed = true
			_ = s.f.Close()
			s.f, s.bw = nil, nil
		}
		return errors.Join(cause, rollbackErr)
	}
	// Preflight rejects directories/symlinks and requires an active file.
	present := make([]bool, s.maxRotate+1)
	for i := 0; i <= s.maxRotate; i++ {
		name := s.path
		if i > 0 {
			name = fmt.Sprintf("%s.%d", s.path, i)
		}
		info, e := os.Lstat(name)
		if os.IsNotExist(e) && i > 0 {
			continue
		}
		if e != nil {
			return e
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("store: rotation requires regular files")
		}
		present[i] = true
	}
	for i, exists := range present {
		if !exists {
			continue
		}
		name := s.path
		if i > 0 {
			name = fmt.Sprintf("%s.%d", s.path, i)
		}
		if e := moveFile(name, filepath.Join(stage, fmt.Sprint(i))); e != nil {
			return rollback(e)
		}
	}
	for i := 0; i < s.maxRotate; i++ {
		if !present[i] {
			continue
		}
		if e := moveFile(filepath.Join(stage, fmt.Sprint(i)), fmt.Sprintf("%s.%d", s.path, i+1)); e != nil {
			return rollback(e)
		}
	}
	if e := moveFile(replacementName, s.path); e != nil {
		return rollback(e)
	}
	if present[s.maxRotate] {
		if e := s.remove(filepath.Join(stage, fmt.Sprint(s.maxRotate))); e != nil {
			return rollback(e)
		}
	}
	// No fallible namespace operations after retention commit.
	closeErr := s.f.Close()
	s.f = replacement
	s.bw = bufio.NewWriterSize(replacement, 64*1024)
	s.written, s.lineCount, s.needsNewline = 0, 0, false
	committed = true
	return errors.Join(closeErr, s.refreshManifestLocked())
}
