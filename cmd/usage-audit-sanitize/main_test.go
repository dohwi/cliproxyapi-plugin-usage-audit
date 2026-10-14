package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	secretAlpha = "sk-ant-test-secret-alpha-998877"
	secretBeta  = "bearer-token-codex-beta-445566"
)

func fileSHA256(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// runRuntimeDecoderProbe separates cold-cache compilation from the runtime
// deadline. The race-enabled parent test does not warm Go's non-race build
// cache, so a first go run can spend the entire runtime budget compiling.
func runRuntimeDecoderProbe(t *testing.T, logPath string) []byte {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "runtime-decoder")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	// Bound cold standard-library compilation separately, leaving headroom
	// within the package's 120-second test timeout on slower CI runners.
	buildCtx, buildCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer buildCancel()
	build := exec.CommandContext(buildCtx, "go", "build", "-o", binary, "./testdata/runtime-decoder")
	buildStarted := time.Now()
	buildOutput, err := build.CombinedOutput()
	buildElapsed := time.Since(buildStarted)
	buildErr := buildCtx.Err()
	buildCancel()
	if err != nil {
		t.Fatalf("runtime decoder probe build: %v (context: %v): %s", err, buildErr, buildOutput)
	}

	// Only execution of the already-built decoder gets the original budget.
	runtimeCtx, runtimeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer runtimeCancel()
	cmd := exec.CommandContext(runtimeCtx, binary, logPath)
	runtimeStarted := time.Now()
	output, err := cmd.CombinedOutput()
	runtimeElapsed := time.Since(runtimeStarted)
	if err != nil {
		// Do not echo output on failure: it may contain corrupt source bytes.
		t.Fatalf("runtime decoder probe execution: %v (context: %v)", err, runtimeCtx.Err())
	}
	t.Logf("runtime decoder probe: build=%s execution=%s", buildElapsed, runtimeElapsed)
	return output
}

// TestDryRunNoModification verifies that dry-run mode never modifies files or creates backups.
func TestDryRunNoModification(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "usage.jsonl")

	content := fmt.Sprintf(`{"t":"2026-09-12T10:00:00Z","provider":"codex","api_key_id":"%s","model":"gpt-4o","total_tokens":100}`+"\n"+
		`{"t":"2026-09-12T10:01:00Z","provider":"antigravity","model":"claude-3-5-sonnet","total_tokens":250}`+"\n",
		secretAlpha)

	if err := os.WriteFile(logPath, []byte(content), 0o640); err != nil {
		t.Fatalf("writing test log: %v", err)
	}
	origHash := fileSHA256(t, logPath)

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		FilePath:       logPath,
		IncludeRotated: false,
		Rewrite:        false,
		Verbose:        true,
	}

	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if summary.FilesInspected != 1 {
		t.Errorf("expected 1 file inspected, got %d", summary.FilesInspected)
	}
	if summary.ExposedRecords != 1 {
		t.Errorf("expected 1 exposed record, got %d", summary.ExposedRecords)
	}
	if summary.CleanRecords != 1 {
		t.Errorf("expected 1 clean record, got %d", summary.CleanRecords)
	}
	if summary.FilesRewritten != 0 {
		t.Errorf("expected 0 files rewritten in dry run, got %d", summary.FilesRewritten)
	}

	// Verify file was NOT modified
	currHash := fileSHA256(t, logPath)
	if currHash != origHash {
		t.Errorf("file was modified during dry run: original %s, current %s", origHash, currHash)
	}

	// Confidentiality: verify secret value was never printed
	outStr := stdout.String() + stderr.String()
	if strings.Contains(outStr, secretAlpha) {
		t.Errorf("CRITICAL: secret value leaked into output:\n%s", outStr)
	}
}

// TestRewriteRequiresBackupDir verifies that --rewrite without --backup-dir is strictly rejected.
func TestRewriteRequiresBackupDir(t *testing.T) {
	_, err := parseFlags([]string{"--rewrite", "--file", "dummy.jsonl"})
	if err == nil {
		t.Fatal("expected error when --rewrite is used without --backup-dir, got nil")
	}
	if !strings.Contains(err.Error(), "--backup-dir") {
		t.Errorf("expected error message mentioning --backup-dir, got: %v", err)
	}
}

// TestRewriteWithBackup verifies atomic rewrite, clean removal of api_key_id, and backup creation.
func TestRewriteWithBackup(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backups")
	logPath := filepath.Join(tmpDir, "usage.jsonl")

	record1 := fmt.Sprintf(`{"t":"2026-09-12T10:00:00Z","provider":"codex","api_key_id":"%s","model":"gpt-4o","total_tokens":100,"latency_ms":123.45}`, secretAlpha)
	record2 := `{"t":"2026-09-12T10:01:00Z","provider":"antigravity","model":"claude-3-5-sonnet","total_tokens":250,"generate":true}`
	record3 := fmt.Sprintf(`{"t":"2026-09-12T10:02:00Z","provider":"codex","api_key_id":"%s","model":"gpt-4o-mini","total_tokens":50}`, secretBeta)
	fileContent := record1 + "\n" + record2 + "\n" + record3 + "\n"

	origPerm := os.FileMode(0o640)
	if err := os.WriteFile(logPath, []byte(fileContent), origPerm); err != nil {
		t.Fatalf("writing test log: %v", err)
	}
	origHash := fileSHA256(t, logPath)

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		FilePath:       logPath,
		IncludeRotated: false,
		Rewrite:        true,
		BackupDir:      backupDir,
	}

	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if summary.FilesRewritten != 1 {
		t.Errorf("expected 1 file rewritten, got %d", summary.FilesRewritten)
	}
	if summary.ExposedRecords != 2 {
		t.Errorf("expected 2 exposed records, got %d", summary.ExposedRecords)
	}

	// 1. Verify backup exists and is identical to original
	backupFile := filepath.Join(backupDir, "usage.jsonl")
	backupHash := fileSHA256(t, backupFile)
	if backupHash != origHash {
		t.Errorf("backup hash %s does not match original %s", backupHash, origHash)
	}

	// 2. Verify rewritten file has permissions preserved
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat rewritten file: %v", err)
	}
	// Windows has no Unix permission bits (files always report 0666/0444).
	if info.Mode().Perm() != origPerm && runtime.GOOS != "windows" {
		t.Errorf("expected permission %v, got %v", origPerm, info.Mode().Perm())
	}

	// 3. Verify api_key_id is completely removed and other fields preserved
	rewrittenData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading rewritten file: %v", err)
	}
	rewrittenLines := strings.Split(strings.TrimSpace(string(rewrittenData)), "\n")
	if len(rewrittenLines) != 3 {
		t.Fatalf("expected 3 lines in rewritten file, got %d", len(rewrittenLines))
	}

	for i, line := range rewrittenLines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("unmarshaling rewritten line %d: %v", i+1, err)
		}
		if _, exists := m["api_key_id"]; exists {
			t.Errorf("line %d still contains api_key_id: %s", i+1, line)
		}
	}

	// 4. Verify specific fields preserved
	var m1 map[string]any
	_ = json.Unmarshal([]byte(rewrittenLines[0]), &m1)
	if m1["provider"] != "codex" || m1["model"] != "gpt-4o" || m1["total_tokens"] != float64(100) {
		t.Errorf("line 1 fields not preserved properly: %v", m1)
	}

	// 5. Subsequent dry-run finds 0 exposures
	stdout.Reset()
	stderr.Reset()
	dryCfg := &Config{FilePath: logPath, IncludeRotated: false}
	drySummary, err := run(dryCfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dry run after rewrite failed: %v", err)
	}
	if drySummary.ExposedRecords != 0 {
		t.Errorf("expected 0 exposures after rewrite, got %d", drySummary.ExposedRecords)
	}

	// 6. Confidentiality check: neither secret leaked in output
	outStr := stdout.String() + stderr.String()
	if strings.Contains(outStr, secretAlpha) || strings.Contains(outStr, secretBeta) {
		t.Errorf("CRITICAL: secret value leaked in output:\n%s", outStr)
	}
}

// TestNeverOutputSecrets verifies that secret tokens are never output across all modes.
func TestNeverOutputSecrets(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "usage.jsonl")

	secretVal := "super-confidential-secret-key-12345"
	line := fmt.Sprintf(`{"t":"2026-09-12T10:00:00Z","provider":"codex","api_key_id":"%s"}`, secretVal) + "\n"
	if err := os.WriteFile(logPath, []byte(line), 0o640); err != nil {
		t.Fatalf("writing log: %v", err)
	}

	testCases := []struct {
		name string
		cfg  Config
	}{
		{
			name: "dry-run verbose",
			cfg:  Config{FilePath: logPath, Verbose: true},
		},
		{
			name: "dry-run json",
			cfg:  Config{FilePath: logPath, JSONOutput: true},
		},
		{
			name: "rewrite verbose",
			cfg:  Config{FilePath: logPath, Rewrite: true, BackupDir: filepath.Join(tmpDir, "bak"), Verbose: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cfgCopy := tc.cfg
			_, err := run(&cfgCopy, &stdout, &stderr)
			if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			combined := stdout.String() + stderr.String()
			if strings.Contains(combined, secretVal) {
				t.Fatalf("CRITICAL SECURITY LEAK: secret %q found in output of %s:\n%s", secretVal, tc.name, combined)
			}
		})
	}
}

// TestCorruptedJSONHandling verifies that corrupted lines do not crash the tool,
// are counted accurately, and are preserved verbatim during rewrite.
func TestCorruptedJSONHandling(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backup")
	logPath := filepath.Join(tmpDir, "usage.jsonl")

	corruptedLine1 := `{"t":"2026-09-12T10:00:00Z", truncated json line`
	corruptedLine2 := `THIS IS NOT JSON AT ALL`
	validExposed := fmt.Sprintf(`{"t":"2026-09-12T10:01:00Z","api_key_id":"%s","provider":"codex"}`, secretAlpha)
	emptyLine := `   `

	fileData := corruptedLine1 + "\n" + validExposed + "\n" + corruptedLine2 + "\n" + emptyLine + "\n"
	if err := os.WriteFile(logPath, []byte(fileData), 0o640); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	// 1. Dry run
	var stdout, stderr bytes.Buffer
	cfg := &Config{FilePath: logPath, IncludeRotated: false, Verbose: true}
	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}

	if summary.CorruptedLines != 2 {
		t.Errorf("expected 2 corrupted lines, got %d", summary.CorruptedLines)
	}
	if summary.ValidRecords != 1 {
		t.Errorf("expected 1 valid record, got %d", summary.ValidRecords)
	}
	if summary.ExposedRecords != 1 {
		t.Errorf("expected 1 exposed record, got %d", summary.ExposedRecords)
	}
	if summary.EmptyLines != 1 {
		t.Errorf("expected 1 empty line, got %d", summary.EmptyLines)
	}

	// 2. Rewrite mode: verify corrupted lines preserved verbatim
	cfg.Rewrite = true
	cfg.BackupDir = backupDir
	stdout.Reset()
	stderr.Reset()

	summary, err = run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("rewrite failed: %v", err)
	}

	rewrittenData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading rewritten file: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(rewrittenData), "\n"), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d: %q", len(lines), rewrittenData)
	}

	if lines[0] != corruptedLine1 {
		t.Errorf("corrupted line 1 not preserved verbatim:\nexpected: %s\ngot:      %s", corruptedLine1, lines[0])
	}
	if lines[2] != corruptedLine2 {
		t.Errorf("corrupted line 2 not preserved verbatim:\nexpected: %s\ngot:      %s", corruptedLine2, lines[2])
	}

	// Verify valid line was cleaned
	var m map[string]any
	if err := json.Unmarshal([]byte(lines[1]), &m); err != nil {
		t.Fatalf("parsing cleaned line: %v", err)
	}
	if _, exists := m["api_key_id"]; exists {
		t.Errorf("api_key_id was not removed from line 2: %s", lines[1])
	}
}

func TestSanitizerPreservesCorruptionAndRuntimeDecoderContinues(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backup")
	logPath := filepath.Join(tmpDir, "usage.jsonl")
	corruptSecret := "approved-corrupt-secret-bytes"
	corrupt := []byte(corruptSecret + "\r\n")
	validExposed := []byte(fmt.Sprintf(`{"model":"before","api_key_id":"%s"}`, secretAlpha) + "\n")
	validAfter := []byte(`{"model":"after"}`) // deliberately no final newline
	original := append(append(append([]byte(nil), validExposed...), corrupt...), validAfter...)
	if err := os.WriteFile(logPath, original, 0o640); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{FilePath: logPath, Rewrite: true, BackupDir: backupDir, IncludeRotated: false, Verbose: true}
	if _, err := run(cfg, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout.String()+stderr.String(), corruptSecret) {
		t.Fatal("corrupt source bytes leaked to sanitizer output")
	}
	rewritten, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(rewritten, corrupt) || !bytes.HasSuffix(rewritten, validAfter) {
		t.Fatalf("corrupt bytes or unterminated clean tail changed: %q", rewritten)
	}
	if bytes.Contains(rewritten, []byte(secretAlpha)) {
		t.Fatal("exposed valid record was not sanitized")
	}

	// Exercise the production store package through the runtime fixture.
	output := runRuntimeDecoderProbe(t, logPath)
	if string(output) != "before\nafter\n" || bytes.Contains(output, []byte(corruptSecret)) {
		t.Fatalf("runtime decoder output: %q", output)
	}
}

// TestActiveAndRotatedDiscovery verifies that rotated files (.1, .2, ...) are discovered and processed.
func TestActiveAndRotatedDiscovery(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backup")

	active := filepath.Join(tmpDir, "usage.jsonl")
	rot1 := filepath.Join(tmpDir, "usage.jsonl.1")
	rot2 := filepath.Join(tmpDir, "usage.jsonl.2")

	lineActive := fmt.Sprintf(`{"t":"2026-09-12T10:00:00Z","api_key_id":"%s","tokens":10}`, secretAlpha)
	lineRot1 := fmt.Sprintf(`{"t":"2026-09-12T09:00:00Z","api_key_id":"%s","tokens":20}`, secretBeta)
	lineRot2 := `{"t":"2026-09-12T08:00:00Z","tokens":30}` // clean

	_ = os.WriteFile(active, []byte(lineActive+"\n"), 0o640)
	_ = os.WriteFile(rot1, []byte(lineRot1+"\n"), 0o640)
	_ = os.WriteFile(rot2, []byte(lineRot2+"\n"), 0o640)

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		FilePath:       active,
		IncludeRotated: true,
		Rewrite:        true,
		BackupDir:      backupDir,
	}

	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if summary.FilesInspected != 3 {
		t.Errorf("expected 3 files inspected, got %d", summary.FilesInspected)
	}
	if summary.FilesWithExposures != 2 {
		t.Errorf("expected 2 files with exposures, got %d", summary.FilesWithExposures)
	}
	if summary.FilesRewritten != 2 {
		t.Errorf("expected 2 files rewritten (rot2 was clean), got %d", summary.FilesRewritten)
	}

	// Verify backups created for the 2 rewritten files
	if _, err := os.Stat(filepath.Join(backupDir, "usage.jsonl")); err != nil {
		t.Errorf("missing backup for active file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(backupDir, "usage.jsonl.1")); err != nil {
		t.Errorf("missing backup for rot1 file: %v", err)
	}
}

// TestCleanFileSkippedRewrite verifies that files with 0 exposures are left untouched.
func TestCleanFileSkippedRewrite(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backup")
	logPath := filepath.Join(tmpDir, "clean.jsonl")

	cleanLine := `{"t":"2026-09-12T10:00:00Z","provider":"antigravity","total_tokens":42}` + "\n"
	_ = os.WriteFile(logPath, []byte(cleanLine), 0o640)
	origHash := fileSHA256(t, logPath)

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		FilePath:       logPath,
		IncludeRotated: false,
		Rewrite:        true,
		BackupDir:      backupDir,
	}

	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	if summary.FilesRewritten != 0 {
		t.Errorf("expected 0 files rewritten for clean file, got %d", summary.FilesRewritten)
	}
	if fileSHA256(t, logPath) != origHash {
		t.Errorf("clean file was modified")
	}

	// Verify no backup was created
	entries, _ := os.ReadDir(backupDir)
	if len(entries) != 0 {
		t.Errorf("expected empty backup dir for clean file, found %d files", len(entries))
	}
}

// TestEmptyAndWhitespaceFiles verifies that empty files and files with only whitespace/empty lines are handled safely.
func TestEmptyAndWhitespaceFiles(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "empty.jsonl")
	if err := os.WriteFile(logPath, []byte("\n  \n\t\n"), 0o640); err != nil {
		t.Fatalf("writing empty file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{FilePath: logPath, IncludeRotated: false}
	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run on empty file failed: %v", err)
	}

	if summary.FilesInspected != 1 {
		t.Errorf("expected 1 file inspected, got %d", summary.FilesInspected)
	}
	if summary.ValidRecords != 0 {
		t.Errorf("expected 0 valid records, got %d", summary.ValidRecords)
	}
	if summary.ExposedRecords != 0 {
		t.Errorf("expected 0 exposed records, got %d", summary.ExposedRecords)
	}
	if summary.EmptyLines != 3 {
		t.Errorf("expected 3 empty lines, got %d", summary.EmptyLines)
	}
}

// TestEmptyStringAndNullApiKeyID verifies that empty string or null api_key_id values are sanitized.
func TestEmptyStringAndNullApiKeyID(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backup")
	logPath := filepath.Join(tmpDir, "null_empty.jsonl")

	line1 := `{"t":"2026-09-12T10:00:00Z","provider":"codex","api_key_id":"","tokens":10}`
	line2 := `{"t":"2026-09-12T10:01:00Z","provider":"antigravity","api_key_id":null,"tokens":20}`
	content := line1 + "\n" + line2 + "\n"
	if err := os.WriteFile(logPath, []byte(content), 0o640); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		FilePath:       logPath,
		IncludeRotated: false,
		Rewrite:        true,
		BackupDir:      backupDir,
	}

	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("rewrite run failed: %v", err)
	}

	if summary.ExposedRecords != 2 {
		t.Errorf("expected 2 exposed records, got %d", summary.ExposedRecords)
	}
	if summary.FilesRewritten != 1 {
		t.Errorf("expected 1 file rewritten, got %d", summary.FilesRewritten)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("reading rewritten file: %v", err)
	}
	for i, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("parsing line %d: %v", i+1, err)
		}
		if _, exists := m["api_key_id"]; exists {
			t.Errorf("line %d still has api_key_id: %s", i+1, line)
		}
	}
}

// TestDirectoryDiscovery verifies that scanning a directory discovers all jsonl files.
func TestDirectoryDiscovery(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "usage.jsonl"), []byte(`{"t":"2026-09-12T00:00:00Z"}`+"\n"), 0o640)
	_ = os.WriteFile(filepath.Join(tmpDir, "usage.jsonl.1"), []byte(`{"t":"2026-09-12T00:00:00Z"}`+"\n"), 0o640)
	_ = os.WriteFile(filepath.Join(tmpDir, "other.jsonl"), []byte(`{"t":"2026-09-12T00:00:00Z"}`+"\n"), 0o640)
	_ = os.WriteFile(filepath.Join(tmpDir, "ignore.txt"), []byte("ignore me"), 0o640)

	var stdout, stderr bytes.Buffer
	cfg := &Config{DirPath: tmpDir, IncludeRotated: true}
	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run on dir failed: %v", err)
	}

	if summary.FilesInspected != 3 {
		t.Errorf("expected 3 files inspected, got %d", summary.FilesInspected)
	}
}

// TestJSONSummaryOutput verifies that --json produces valid JSON statistics.
func TestJSONSummaryOutput(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "usage.jsonl")
	_ = os.WriteFile(logPath, []byte(`{"t":"2026-09-12T00:00:00Z","tokens":10}`+"\n"), 0o640)

	var stdout, stderr bytes.Buffer
	cfg := &Config{FilePath: logPath, JSONOutput: true}
	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}

	var parsed SummaryStats
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("decoding stdout JSON: %v\nOutput: %s", err, stdout.String())
	}

	if parsed.FilesInspected != summary.FilesInspected || parsed.TotalLines != summary.TotalLines {
		t.Errorf("parsed JSON stats do not match struct stats")
	}
}

// TestSanitizerPreservesTwoOperationalCorruptLinesVerbatimAndRuntimeDecodes verifies that
// the two operational corrupted lines are preserved byte-for-byte during rewrite,
// exposed valid records are sanitized, and the runtime resilient decoder successfully
// skips the two corrupted lines and reads all valid records before, between, and after.
func TestSanitizerPreservesTwoOperationalCorruptLinesVerbatimAndRuntimeDecodes(t *testing.T) {
	tmpDir := t.TempDir()
	backupDir := filepath.Join(tmpDir, "backup")
	logPath := filepath.Join(tmpDir, "usage.jsonl")

	corruptedLine1 := `{"t":"2026-09-12T10:00:00Z", truncated json line`
	corruptedLine2 := `THIS IS NOT JSON AT ALL`

	rec1 := fmt.Sprintf(`{"t":"2026-09-12T10:01:00Z","model":"m1","api_key_id":"%s"}`, secretAlpha)
	rec2 := fmt.Sprintf(`{"t":"2026-09-12T10:02:00Z","model":"m2","api_key_id":"%s"}`, secretBeta)
	rec3 := fmt.Sprintf(`{"t":"2026-09-12T10:03:00Z","model":"m3","api_key_id":"%s"}`, secretAlpha)

	fileData := rec1 + "\n" +
		corruptedLine1 + "\n" +
		rec2 + "\n" +
		corruptedLine2 + "\n" +
		rec3 + "\n"

	if err := os.WriteFile(logPath, []byte(fileData), 0o640); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	cfg := &Config{
		FilePath:       logPath,
		Rewrite:        true,
		BackupDir:      backupDir,
		IncludeRotated: false,
		Verbose:        true,
	}

	summary, err := run(cfg, &stdout, &stderr)
	if err != nil {
		t.Fatalf("sanitizer run failed: %v", err)
	}
	if summary.CorruptedLines != 2 {
		t.Errorf("expected 2 corrupted lines, got %d", summary.CorruptedLines)
	}
	if summary.ExposedRecords != 3 {
		t.Errorf("expected 3 exposed records, got %d", summary.ExposedRecords)
	}
	if summary.ValidRecords != 3 {
		t.Errorf("expected 3 valid records, got %d", summary.ValidRecords)
	}

	// Verify rewritten file content
	rewrittenBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	rewrittenStr := string(rewrittenBytes)
	lines := strings.Split(strings.TrimRight(rewrittenStr, "\n"), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 lines in rewritten file, got %d: %q", len(lines), rewrittenStr)
	}

	// Corrupted lines must be preserved verbatim
	if lines[1] != corruptedLine1 {
		t.Errorf("corrupted line 1 mismatch:\nwant: %s\ngot:  %s", corruptedLine1, lines[1])
	}
	if lines[3] != corruptedLine2 {
		t.Errorf("corrupted line 2 mismatch:\nwant: %s\ngot:  %s", corruptedLine2, lines[3])
	}

	// Secrets must be stripped
	if strings.Contains(rewrittenStr, secretAlpha) || strings.Contains(rewrittenStr, secretBeta) {
		t.Fatal("secret leaked in rewritten file")
	}

	// Verify runtime store package reads m1, m2, m3 without failing on the 2 preserved corrupt lines.
	output := runRuntimeDecoderProbe(t, logPath)

	wantOutput := "m1\nm2\nm3\n"
	if string(output) != wantOutput {
		t.Fatalf("runtime decoder output mismatch: want %q, got %q", wantOutput, string(output))
	}
}
