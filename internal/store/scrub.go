package store

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// legacySecretFields are record keys older versions wrote that can hold a raw
// credential: api_key_id (the client API key) and source (the host's usage
// source, which is the upstream API key of an API-key credential, the account
// e-mail, or the client API key).
var legacySecretFields = []string{"api_key_id", "source"}

// lineHasLegacySecret reports whether a JSON object line carries any legacy
// secret field. It checks the cheap byte form first so clean history is not
// decoded twice.
func lineHasLegacySecret(content []byte) bool {
	found := false
	for _, k := range legacySecretFields {
		if bytes.Contains(content, []byte(`"`+k+`"`)) {
			found = true
			break
		}
	}
	if !found {
		return false
	}
	var m map[string]json.RawMessage
	if json.Unmarshal(content, &m) != nil {
		return false
	}
	for _, k := range legacySecretFields {
		if _, ok := m[k]; ok {
			return true
		}
	}
	return false
}

// ScrubLegacySecrets removes legacy secret fields from the active log and its
// rotations at path. Files without such fields are left untouched; a file that
// has them is rewritten through a temp file and renamed over the original with
// the same permissions. Malformed and empty lines are copied byte-for-byte.
// It returns how many lines were cleaned. Call it only while no Store has the
// files open (Open does so before creating its writer).
func ScrubLegacySecrets(path string) (int, error) {
	paths, err := logPaths(path)
	if err != nil {
		return 0, err
	}
	total := 0
	var errs []error
	for _, p := range paths {
		n, err := scrubFile(p)
		total += n
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
	}
	return total, errors.Join(errs...)
}

func scrubFile(path string) (int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if !info.Mode().IsRegular() {
		return 0, nil
	}
	if dirty, err := fileHasLegacySecret(path); err != nil || !dirty {
		return 0, err
	}

	in, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer in.Close()
	tmp, err := os.CreateTemp(filepath.Dir(path), ".usage-scrub-*.tmp")
	if err != nil {
		return 0, fmt.Errorf("store: scrub temp file: %w", err)
	}
	tmpPath := tmp.Name()
	keep := false
	defer func() {
		if !keep {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	cleaned := 0
	reader := bufio.NewReader(in)
	w := bufio.NewWriterSize(tmp, 64*1024)
	for {
		line, readErr := reader.ReadBytes('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return 0, readErr
		}
		if len(line) > 0 {
			content := bytes.TrimRight(line, "\r\n")
			ending := line[len(content):]
			out := line
			if lineHasLegacySecret(content) {
				var m map[string]json.RawMessage
				if json.Unmarshal(content, &m) == nil {
					for _, k := range legacySecretFields {
						delete(m, k)
					}
					clean, err := json.Marshal(m)
					if err != nil {
						return 0, fmt.Errorf("store: scrub marshal: %w", err)
					}
					out = append(clean, ending...)
					cleaned++
				}
			}
			if _, err := w.Write(out); err != nil {
				return 0, err
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
	}
	if err := w.Flush(); err != nil {
		return 0, err
	}
	if err := tmp.Sync(); err != nil {
		return 0, err
	}
	if err := tmp.Close(); err != nil {
		return 0, err
	}
	// Windows cannot replace a file that still has an open handle.
	if err := in.Close(); err != nil {
		return 0, err
	}
	if err := os.Chmod(tmpPath, info.Mode().Perm()); err != nil {
		return 0, err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return 0, fmt.Errorf("store: scrub rename: %w", err)
	}
	keep = true
	return cleaned, nil
}

func fileHasLegacySecret(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		line, readErr := reader.ReadBytes('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return false, readErr
		}
		if lineHasLegacySecret(bytes.TrimRight(line, "\r\n")) {
			return true, nil
		}
		if errors.Is(readErr, io.EOF) {
			return false, nil
		}
	}
}
