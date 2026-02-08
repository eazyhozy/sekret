package cmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeScanFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func setupScan(t *testing.T) {
	t.Helper()
	setup(t)
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
		require.NoError(t, executeCmd(t, "scan", "--path", path))
	})

	assert.Contains(t, output, "No plaintext keys found")
	assert.Contains(t, output, "Scanned 1 file")
	assert.Contains(t, output, "clean")
}

func TestScan_FindsKeys(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"
export PATH="/usr/bin"
export GITHUB_TOKEN="ghp_abcdef1234567890"
`)

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "scan", "--path", path))
	})

	assert.Contains(t, output, "Scanned 1 file")
	assert.Contains(t, output, "2 keys found")
	assert.Contains(t, output, "Found 2 potential plaintext keys")
	assert.Contains(t, output, "OPENAI_API_KEY")
	assert.Contains(t, output, "GITHUB_TOKEN")
	assert.NotContains(t, output, "sk-proj-abcdef1234")
}

func TestScan_AnnotateSafeToRemove(t *testing.T) {
	setupScan(t)
	seedKey(t, "OPENAI_API_KEY", "sk-proj-abcdef1234")

	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"
`)

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "scan", "--path", path))
	})

	assert.Contains(t, output, "safe to remove")
}

func TestScan_AnnotateValueDiffers(t *testing.T) {
	setupScan(t)
	seedKey(t, "OPENAI_API_KEY", "sk-proj-different-value")

	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"
`)

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "scan", "--path", path))
	})

	assert.Contains(t, output, "value differs!")
}

func TestScan_PathDirectory(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	writeScanFile(t, dir, "a.sh", `export MY_API_KEY="abc123secret"`)
	writeScanFile(t, dir, "b.sh", `export AUTH_TOKEN="token_value_here"`)

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "scan", "--path", dir))
	})

	assert.Contains(t, output, "MY_API_KEY")
	assert.Contains(t, output, "AUTH_TOKEN")
}

func TestScan_ExitCode(t *testing.T) {
	setupScan(t)

	var exitCode int
	cmd.SetExitFunc(func(code int) { exitCode = code })

	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"`)

	captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "scan", "--path", path))
	})

	assert.Equal(t, 1, exitCode)
}

func TestScan_ExitCodeZero(t *testing.T) {
	setupScan(t)

	exitCode := -1
	cmd.SetExitFunc(func(code int) { exitCode = code })

	dir := t.TempDir()
	path := writeScanFile(t, dir, "config.sh", `export PATH="/usr/bin"`)

	captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "scan", "--path", path))
	})

	assert.Equal(t, -1, exitCode, "exitFunc should not have been called")
}

func TestScan_SummaryMultipleFiles(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	writeScanFile(t, dir, "a.sh", `export MY_API_KEY="abc123secret"`)
	writeScanFile(t, dir, "b.sh", `export PATH="/usr/bin"`)

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "scan", "--path", dir))
	})

	assert.Contains(t, output, "Scanned 2 files")
	assert.Contains(t, output, "1 key found")
	assert.Contains(t, output, "clean")
}

func TestScan_PathNotFound(t *testing.T) {
	setupScan(t)

	err := executeCmd(t, "scan", "--path", "/nonexistent/path/file.sh")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no such file")
}

func TestScan_DuplicateKeyAcrossFiles(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	writeScanFile(t, dir, "a.sh", `export OPENAI_API_KEY="sk-proj-abcdef1234"`)
	writeScanFile(t, dir, "b.sh", `export OPENAI_API_KEY="sk-proj-different9999"`)

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "scan", "--path", dir))
	})

	assert.Contains(t, output, "Found 2 potential plaintext keys")
	assert.Contains(t, output, "a.sh")
	assert.Contains(t, output, "b.sh")
}

func TestScan_SingularKey(t *testing.T) {
	setupScan(t)
	dir := t.TempDir()
	path := writeScanFile(t, dir, ".zshrc", `export OPENAI_API_KEY="sk-proj-abcdef1234"`)

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "scan", "--path", path))
	})

	assert.Contains(t, output, "Found 1 potential plaintext key:")
}
