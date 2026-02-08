package cmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eazyhozy/sekret/cmd"
)

func writeScanFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func setupScan(t *testing.T) {
	t.Helper()
	setup(t)
	// Override exit function to prevent os.Exit in tests
	cmd.SetExitFunc(func(_ int) {})
	t.Cleanup(func() {
		cmd.SetExitFunc(os.Exit)
	})
}

func TestScan_NoKeysFound(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	path := writeScanFile(t, dir, "config.sh", `export PATH="/usr/bin"
export EDITOR="vim"
`)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", path); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	if !strings.Contains(output, "No plaintext keys found") {
		t.Errorf("expected 'No plaintext keys found' message, got %q", output)
	}
	if !strings.Contains(output, "Scanned 1 file") {
		t.Errorf("expected scan summary, got %q", output)
	}
	if !strings.Contains(output, "clean") {
		t.Errorf("expected 'clean' in summary, got %q", output)
	}
}

func TestScan_FindsKeys(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"
export PATH="/usr/bin"
export GITHUB_TOKEN="ghp_abcdef1234567890"
`)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", path); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	if !strings.Contains(output, "Scanned 1 file") {
		t.Errorf("expected scan summary, got %q", output)
	}
	if !strings.Contains(output, "2 keys found") {
		t.Errorf("expected '2 keys found' in summary, got %q", output)
	}
	if !strings.Contains(output, "Found 2 potential plaintext keys") {
		t.Errorf("expected 'Found 2' message, got %q", output)
	}
	if !strings.Contains(output, "OPENAI_API_KEY") {
		t.Errorf("expected OPENAI_API_KEY in output, got %q", output)
	}
	if !strings.Contains(output, "GITHUB_TOKEN") {
		t.Errorf("expected GITHUB_TOKEN in output, got %q", output)
	}
	// Values should be masked
	if strings.Contains(output, "sk-proj-abcdef1234") {
		t.Error("full key value should not appear in output")
	}
}

func TestScan_AnnotateSafeToRemove(t *testing.T) {
	setupScan(t)
	seedKey(t, "OPENAI_API_KEY", "sk-proj-abcdef1234")

	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"
`)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", path); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	if !strings.Contains(output, "safe to remove") {
		t.Errorf("expected 'safe to remove' annotation, got %q", output)
	}
}

func TestScan_AnnotateValueDiffers(t *testing.T) {
	setupScan(t)
	seedKey(t, "OPENAI_API_KEY", "sk-proj-different-value")

	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"
`)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", path); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	if !strings.Contains(output, "value differs!") {
		t.Errorf("expected 'value differs!' annotation, got %q", output)
	}
}

func TestScan_PathDirectory(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	writeScanFile(t, dir, "a.sh", `export MY_API_KEY="abc123secret"`)
	writeScanFile(t, dir, "b.sh", `export AUTH_TOKEN="token_value_here"`)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", dir); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	if !strings.Contains(output, "MY_API_KEY") {
		t.Errorf("expected MY_API_KEY in output, got %q", output)
	}
	if !strings.Contains(output, "AUTH_TOKEN") {
		t.Errorf("expected AUTH_TOKEN in output, got %q", output)
	}
}

func TestScan_ExitCode(t *testing.T) {
	setupScan(t)

	var exitCode int
	cmd.SetExitFunc(func(code int) { exitCode = code })

	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"`)

	captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", path); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestScan_ExitCodeZero(t *testing.T) {
	setupScan(t)

	exitCode := -1
	cmd.SetExitFunc(func(code int) { exitCode = code })

	dir := t.TempDir()
	path := writeScanFile(t, dir, "config.sh", `export PATH="/usr/bin"`)

	captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", path); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	// exitFunc should not have been called
	if exitCode != -1 {
		t.Errorf("expected exit func not to be called, got code %d", exitCode)
	}
}

func TestScan_SummaryMultipleFiles(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	writeScanFile(t, dir, "a.sh", `export MY_API_KEY="abc123secret"`)
	writeScanFile(t, dir, "b.sh", `export PATH="/usr/bin"`)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", dir); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	if !strings.Contains(output, "Scanned 2 files") {
		t.Errorf("expected 'Scanned 2 files', got %q", output)
	}
	if !strings.Contains(output, "1 key found") {
		t.Errorf("expected '1 key found' in summary, got %q", output)
	}
	if !strings.Contains(output, "clean") {
		t.Errorf("expected 'clean' for file with no keys, got %q", output)
	}
}

func TestScan_PathNotFound(t *testing.T) {
	setupScan(t)

	err := executeCmd(t, "scan", "--path", "/nonexistent/path/file.sh")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
	if !strings.Contains(err.Error(), "no such file") {
		t.Errorf("expected 'no such file' error, got %q", err.Error())
	}
}

func TestScan_DuplicateKeyAcrossFiles(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	writeScanFile(t, dir, "a.sh", `export OPENAI_API_KEY="sk-proj-abcdef1234"`)
	writeScanFile(t, dir, "b.sh", `export OPENAI_API_KEY="sk-proj-different9999"`)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", dir); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	if !strings.Contains(output, "Found 2 potential plaintext keys") {
		t.Errorf("expected both duplicates reported, got %q", output)
	}
	// Both file names should appear
	if !strings.Contains(output, "a.sh") {
		t.Errorf("expected a.sh in output, got %q", output)
	}
	if !strings.Contains(output, "b.sh") {
		t.Errorf("expected b.sh in output, got %q", output)
	}
}

func TestScan_SingularKey(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"`)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "scan", "--path", path); err != nil {
			t.Fatalf("scan command failed: %v", err)
		}
	})

	if !strings.Contains(output, "Found 1 potential plaintext key:") {
		t.Errorf("expected singular 'key' in output, got %q", output)
	}
}
