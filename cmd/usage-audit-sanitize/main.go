// Package main provides a non-destructive CLI inspection and sanitization tool
// for usage-audit JSONL logs. It inspects and removes legacy api_key_id fields
// without exposing secrets, crashing on corrupted records, or destroying data.
//
// Safety rules:
//  1. Default mode is read-only dry-run statistics.
//  2. Modifying files requires BOTH --rewrite AND --backup-dir explicitly.
//  3. Active and rotated log files (*.jsonl, *.jsonl.1, ...) are supported.
//  4. Modifications use atomic temp-file creation and rename in the same filesystem.
//  5. Original file permissions are preserved.
//  6. Malformed/corrupted JSON lines are handled gracefully and preserved verbatim.
//  7. Secret values are NEVER output to stdout, stderr, logs, or JSON summaries.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// FileStats holds the inspection/sanitization results for a single file.
type FileStats struct {
	Path           string `json:"path"`
	TotalLines     int64  `json:"total_lines"`
	ValidRecords   int64  `json:"valid_records"`
	ExposedRecords int64  `json:"exposed_records"`
	CleanRecords   int64  `json:"clean_records"`
	CorruptedLines int64  `json:"corrupted_lines"`
	EmptyLines     int64  `json:"empty_lines"`
	Rewritten      bool   `json:"rewritten"`
	BackupPath     string `json:"backup_path,omitempty"`
}

// SummaryStats holds the aggregate results across all processed files.
type SummaryStats struct {
	Mode               string      `json:"mode"`
	FilesInspected     int         `json:"files_inspected"`
	FilesWithExposures int         `json:"files_with_exposures"`
	TotalLines         int64       `json:"total_lines"`
	ValidRecords       int64       `json:"valid_records"`
	ExposedRecords     int64       `json:"exposed_records"`
	CleanRecords       int64       `json:"clean_records"`
	CorruptedLines     int64       `json:"corrupted_lines"`
	EmptyLines         int64       `json:"empty_lines"`
	FilesRewritten     int         `json:"files_rewritten"`
	BackupDir          string      `json:"backup_dir,omitempty"`
	Files              []FileStats `json:"files,omitempty"`
}

// Config holds runtime options parsed from flags and args.
type Config struct {
	FilePath       string
	DirPath        string
	PositionalArgs []string
	IncludeRotated bool
	Rewrite        bool
	BackupDir      string
	FailOnExposure bool
	JSONOutput     bool
	Verbose        bool
}

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	summary, err := run(cfg, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}

	if cfg.FailOnExposure && summary.ExposedRecords > 0 {
		os.Exit(2)
	}
}

func parseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("usage-audit-sanitize", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	cfg := &Config{}
	fs.StringVar(&cfg.FilePath, "file", "", "Path to a specific JSONL file or base log path")
	fs.StringVar(&cfg.DirPath, "dir", "", "Path to a directory containing JSONL log files")
	fs.BoolVar(&cfg.IncludeRotated, "include-rotated", true, "Automatically discover rotated files (.1, .2, ...)")
	fs.BoolVar(&cfg.Rewrite, "rewrite", false, "Sanitize files in place (requires --backup-dir)")
	fs.StringVar(&cfg.BackupDir, "backup-dir", "", "Directory where original files are backed up before rewrite")
	fs.BoolVar(&cfg.FailOnExposure, "fail-on-exposure", false, "Exit with code 2 if any api_key_id exposure is found")
	fs.BoolVar(&cfg.JSONOutput, "json", false, "Output summary statistics in JSON format")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "Show detailed per-line inspection progress")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: usage-audit-sanitize [options] [path ...]\n\n")
		fmt.Fprintf(os.Stderr, "Non-destructive inspection and sanitization tool for usage-audit JSONL logs.\n")
		fmt.Fprintf(os.Stderr, "Default operation is read-only dry-run statistics.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nSafety requirements:\n")
		fmt.Fprintf(os.Stderr, "  --rewrite requires --backup-dir to prevent irreversible data modification.\n")
		fmt.Fprintf(os.Stderr, "  Secret values are never displayed in output or logs.\n")
	}

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	cfg.PositionalArgs = fs.Args()

	// Strict Safety Guard: rewrite requires backup-dir
	if cfg.Rewrite && strings.TrimSpace(cfg.BackupDir) == "" {
		return nil, errors.New("--rewrite requires --backup-dir to be specified; refusing to modify files without a backup directory")
	}

	return cfg, nil
}

func run(cfg *Config, stdout, stderr io.Writer) (*SummaryStats, error) {
	targets, err := discoverFiles(cfg)
	if err != nil {
		return nil, fmt.Errorf("discovering log files: %w", err)
	}

	if len(targets) == 0 {
		return nil, errors.New("no log files found to inspect; provide --file, --dir, or paths as arguments")
	}

	mode := "DRY-RUN (read-only)"
	if cfg.Rewrite {
		mode = "REWRITE (sanitized with backup)"
		if err := os.MkdirAll(cfg.BackupDir, 0o750); err != nil {
			return nil, fmt.Errorf("creating backup directory %s: %w", cfg.BackupDir, err)
		}
	} else if cfg.BackupDir != "" {
		fmt.Fprintln(stderr, "Notice: --backup-dir specified without --rewrite; running in safe DRY-RUN mode.")
	}

	summary := &SummaryStats{
		Mode:      mode,
		BackupDir: cfg.BackupDir,
		Files:     make([]FileStats, 0, len(targets)),
	}

	for _, file := range targets {
		stats, err := processFile(file, cfg, stdout, stderr)
		if err != nil {
			return nil, fmt.Errorf("processing %s: %w", file, err)
		}

		summary.Files = append(summary.Files, stats)
		summary.FilesInspected++
		summary.TotalLines += stats.TotalLines
		summary.ValidRecords += stats.ValidRecords
		summary.ExposedRecords += stats.ExposedRecords
		summary.CleanRecords += stats.CleanRecords
		summary.CorruptedLines += stats.CorruptedLines
		summary.EmptyLines += stats.EmptyLines
		if stats.ExposedRecords > 0 {
			summary.FilesWithExposures++
		}
		if stats.Rewritten {
			summary.FilesRewritten++
		}
	}

	if cfg.JSONOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(summary); err != nil {
			return nil, fmt.Errorf("encoding JSON summary: %w", err)
		}
	} else {
		printHumanSummary(stdout, summary)
	}

	return summary, nil
}

// discoverFiles finds all target files based on config flags and positional args.
func discoverFiles(cfg *Config) ([]string, error) {
	fileSet := make(map[string]struct{})

	addFile := func(path string) error {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		info, err := os.Stat(abs)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return scanDirectory(abs, cfg.IncludeRotated, fileSet)
		}
		fileSet[abs] = struct{}{}
		if cfg.IncludeRotated {
			if err := discoverRotations(abs, fileSet); err != nil {
				return err
			}
		}
		return nil
	}

	if cfg.FilePath != "" {
		if err := addFile(cfg.FilePath); err != nil {
			return nil, err
		}
	}
	if cfg.DirPath != "" {
		if err := scanDirectory(cfg.DirPath, cfg.IncludeRotated, fileSet); err != nil {
			return nil, err
		}
	}
	for _, arg := range cfg.PositionalArgs {
		if err := addFile(arg); err != nil {
			return nil, err
		}
	}

	var files []string
	for f := range fileSet {
		files = append(files, f)
	}
	sort.Strings(files)
	return files, nil
}

func scanDirectory(dir string, includeRotated bool, fileSet map[string]struct{}) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(absDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match files like usage.jsonl or *.jsonl or rotated logs
		if strings.HasSuffix(name, ".jsonl") {
			full := filepath.Join(absDir, name)
			fileSet[full] = struct{}{}
			if includeRotated {
				_ = discoverRotations(full, fileSet)
			}
		} else if strings.Contains(name, ".jsonl.") {
			full := filepath.Join(absDir, name)
			fileSet[full] = struct{}{}
		}
	}
	return nil
}

// discoverRotations finds numeric sibling rotations (e.g. usage.jsonl.1, usage.jsonl.2).
func discoverRotations(basePath string, fileSet map[string]struct{}) error {
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)
	prefix := base + "."

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, prefix) {
			suffix := strings.TrimPrefix(name, prefix)
			if n, err := strconv.Atoi(suffix); err == nil && n > 0 && strconv.Itoa(n) == suffix {
				fileSet[filepath.Join(dir, name)] = struct{}{}
			}
		}
	}
	return nil
}

// processFile inspects a file and optionally rewrites it if --rewrite is true.
func processFile(path string, cfg *Config, stdout, stderr io.Writer) (FileStats, error) {
	stats := FileStats{Path: path}

	info, err := os.Stat(path)
	if err != nil {
		return stats, err
	}
	origPerm := info.Mode().Perm()

	// Pass 1: Read-only inspection
	f, err := os.Open(path)
	if err != nil {
		return stats, err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	var lineNum int64

	for {
		line, readErr := reader.ReadBytes('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return stats, readErr
		}

		if len(line) > 0 {
			lineNum++
			trimmed := bytes.TrimRight(line, "\r\n")
			if len(bytes.TrimSpace(trimmed)) == 0 {
				stats.EmptyLines++
			} else {
				var rawMap map[string]json.RawMessage
				if err := json.Unmarshal(trimmed, &rawMap); err != nil || len(trimmed) == 0 || trimmed[0] != '{' {
					stats.CorruptedLines++
					if cfg.Verbose {
						fmt.Fprintf(stderr, "  [corrupted] %s: line %d is not a valid JSON object\n", filepath.Base(path), lineNum)
					}
				} else {
					stats.ValidRecords++
					if _, hasExposure := rawMap["api_key_id"]; hasExposure {
						stats.ExposedRecords++
						if cfg.Verbose {
							// Strict: NEVER print secret or key value, only line and redacted marker
							fmt.Fprintf(stdout, "  [exposure] %s: line %d has api_key_id (redacted)\n", filepath.Base(path), lineNum)
						}
					} else {
						stats.CleanRecords++
					}
				}
			}
		}

		if errors.Is(readErr, io.EOF) {
			break
		}
	}
	stats.TotalLines = lineNum

	// If not rewriting or no exposures found, we are done
	if !cfg.Rewrite || stats.ExposedRecords == 0 {
		return stats, nil
	}

	// Pass 2: Rewrite with verified backup
	_ = f.Close() // close read handle before backup & rewrite

	backupPath, err := createVerifiedBackup(path, cfg.BackupDir)
	if err != nil {
		return stats, fmt.Errorf("creating backup for %s: %w", path, err)
	}
	stats.BackupPath = backupPath

	if err := atomicSanitizeRewrite(path, origPerm); err != nil {
		return stats, fmt.Errorf("atomic rewrite of %s: %w", path, err)
	}
	stats.Rewritten = true

	return stats, nil
}

// createVerifiedBackup copies src to backupDir and verifies SHA-256 integrity.
func createVerifiedBackup(src, backupDir string) (string, error) {
	baseName := filepath.Base(src)
	dest := filepath.Join(backupDir, baseName)

	// If dest already exists with same name, make a unique timestamped path to prevent clobbering
	if _, err := os.Stat(dest); err == nil {
		dest = filepath.Join(backupDir, fmt.Sprintf("%s.orig-%d", baseName, time.Now().UnixNano()))
	}

	in, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return "", err
	}

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return "", err
	}
	defer out.Close()

	srcHash := sha256.New()
	dstHash := sha256.New()

	reader := io.TeeReader(in, srcHash)
	writer := io.MultiWriter(out, dstHash)

	if _, err := io.Copy(writer, reader); err != nil {
		_ = os.Remove(dest)
		return "", err
	}

	if err := out.Sync(); err != nil {
		_ = os.Remove(dest)
		return "", err
	}

	srcSum := hex.EncodeToString(srcHash.Sum(nil))
	dstSum := hex.EncodeToString(dstHash.Sum(nil))
	if srcSum != dstSum {
		_ = os.Remove(dest)
		return "", fmt.Errorf("backup checksum mismatch: src %s != dst %s", srcSum, dstSum)
	}

	return dest, nil
}

// atomicSanitizeRewrite streams lines from path to a temp file in the same directory,
// deletes "api_key_id" from JSON objects, preserves corrupted lines verbatim,
// fsyncs, preserves permissions, and renames atomically.
func atomicSanitizeRewrite(path string, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, ".usage-sanitize-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	cleanUp := true
	defer func() {
		if cleanUp {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	in, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening source file for rewrite: %w", err)
	}
	defer in.Close()

	reader := bufio.NewReader(in)
	bw := bufio.NewWriterSize(tmpFile, 64*1024)

	for {
		line, readErr := reader.ReadBytes('\n')
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return readErr
		}

		if len(line) > 0 {
			content, ending := splitLineEnding(line)
			if len(bytes.TrimSpace(content)) == 0 {
				// Empty/whitespace line: preserve every byte, including CRLF or
				// the absence of a final terminator.
				if _, err := bw.Write(line); err != nil {
					return err
				}
			} else {
				var rawMap map[string]json.RawMessage
				if err := json.Unmarshal(content, &rawMap); err != nil || content[0] != '{' {
					// Approved corrupted/non-object lines are immutable evidence.
					if _, err := bw.Write(line); err != nil {
						return err
					}
				} else if _, hasExposure := rawMap["api_key_id"]; hasExposure {
					delete(rawMap, "api_key_id")
					cleanJSON, err := json.Marshal(rawMap)
					if err != nil {
						return fmt.Errorf("remarshaling sanitized JSON: %w", err)
					}
					if _, err := bw.Write(cleanJSON); err != nil {
						return err
					}
					if _, err := bw.Write(ending); err != nil {
						return err
					}
				} else {
					// Clean line: preserve the original representation exactly.
					if _, err := bw.Write(line); err != nil {
						return err
					}
				}
			}
		}

		if errors.Is(readErr, io.EOF) {
			break
		}
	}

	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flushing temp file: %w", err)
	}
	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("fsync temp file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("setting permissions on temp file: %w", err)
	}

	// Atomic rename replaces the original file
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("atomic rename %s -> %s: %w", tmpPath, path, err)
	}

	cleanUp = false
	return nil
}

func splitLineEnding(line []byte) (content, ending []byte) {
	if len(line) > 0 && line[len(line)-1] == '\n' {
		start := len(line) - 1
		if start > 0 && line[start-1] == '\r' {
			start--
		}
		return line[:start], line[start:]
	}
	return line, nil
}

func printHumanSummary(w io.Writer, s *SummaryStats) {
	fmt.Fprintf(w, "=== Usage Log Sanitization Report ===\n")
	fmt.Fprintf(w, "Mode:                 %s\n", s.Mode)
	fmt.Fprintf(w, "Files inspected:      %d\n", s.FilesInspected)
	fmt.Fprintf(w, "Files with exposures: %d\n", s.FilesWithExposures)
	fmt.Fprintf(w, "Total lines:          %d\n", s.TotalLines)
	fmt.Fprintf(w, "Valid JSON records:   %d\n", s.ValidRecords)
	fmt.Fprintf(w, "Exposed records:      %d\n", s.ExposedRecords)
	fmt.Fprintf(w, "Clean records:        %d\n", s.CleanRecords)
	fmt.Fprintf(w, "Corrupted JSON lines: %d\n", s.CorruptedLines)
	fmt.Fprintf(w, "Empty lines:          %d\n", s.EmptyLines)
	if s.BackupDir != "" {
		fmt.Fprintf(w, "Backup directory:     %s\n", s.BackupDir)
	}
	fmt.Fprintf(w, "Files rewritten:      %d\n", s.FilesRewritten)

	if len(s.Files) > 0 {
		fmt.Fprintf(w, "\n--- Per-File Breakdown ---\n")
		for _, f := range s.Files {
			status := "CLEAN"
			if f.ExposedRecords > 0 {
				if f.Rewritten {
					status = "SANITIZED"
				} else {
					status = "EXPOSURES FOUND"
				}
			}
			fmt.Fprintf(w, "- %s [%s]\n", f.Path, status)
			fmt.Fprintf(w, "    lines: %d, valid: %d, exposed: %d, corrupted: %d, empty: %d\n",
				f.TotalLines, f.ValidRecords, f.ExposedRecords, f.CorruptedLines, f.EmptyLines)
			if f.BackupPath != "" {
				fmt.Fprintf(w, "    backup: %s\n", f.BackupPath)
			}
		}
	}
	fmt.Fprintf(w, "=====================================\n")
}
