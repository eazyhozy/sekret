package cmd_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/keychain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// choiceSequence returns a readChoice mock that returns answers in order.
func choiceSequence(answers ...string) func(string) (string, error) {
	i := 0
	return func(_ string) (string, error) {
		if i >= len(answers) {
			return "", fmt.Errorf("unexpected readChoice call #%d", i+1)
		}
		answer := answers[i]
		i++
		return answer, nil
	}
}

// executeImport runs an import command and captures both stdout and stderr.
func executeImport(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	var stderrBuf bytes.Buffer
	cmd.RootCmd().SetErr(&stderrBuf)
	t.Cleanup(func() { cmd.RootCmd().SetErr(nil) })

	fullArgs := append([]string{"import"}, args...)
	stdout = captureStdout(t, func() {
		err = executeCmd(t, fullArgs...)
	})
	stderr = stderrBuf.String()
	return
}

func writeImportFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.sh")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestImport_NoKeysFound(t *testing.T) {
	setup(t)
	path := writeImportFile(t, `export PATH="/usr/bin"
export EDITOR="vim"
`)

	_, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)
	assert.Contains(t, stderr, "No exportable keys found")
}

func TestImport_ImportSingleKey(t *testing.T) {
	setup(t)
	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-abcdef1234"`)
	cmd.SetReadChoice(choiceSequence("y"))

	stdout, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	// stderr: interactive messages
	assert.Contains(t, stderr, "Found 1 exportable key")
	assert.Contains(t, stderr, "[1/1] OPENAI_API_KEY")
	assert.Contains(t, stderr, "Imported OPENAI_API_KEY")

	// stdout: summary
	assert.Contains(t, stdout, "1 imported")
	assert.Contains(t, stdout, "Imported:")
	assert.Contains(t, stdout, "OPENAI_API_KEY")
	assert.Contains(t, stdout, "Remove the imported keys")

	// Verify keychain and config
	val, err := testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "sk-proj-abcdef1234", val)

	cfg, err := config.Load()
	require.NoError(t, err)
	assert.NotNil(t, cfg.FindKeyByEnvVar("OPENAI_API_KEY"))
}

func TestImport_DefaultYes(t *testing.T) {
	setup(t)
	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-abcdef1234"`)
	cmd.SetReadChoice(choiceSequence("")) // empty = Enter = default yes

	stdout, _, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stdout, "1 imported")

	val, err := testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "sk-proj-abcdef1234", val)
}

func TestImport_SkipKey(t *testing.T) {
	setup(t)
	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-abcdef1234"`)
	cmd.SetReadChoice(choiceSequence("s"))

	stdout, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stderr, "Skipped")
	assert.Contains(t, stdout, "1 skipped")
	assert.Contains(t, stdout, "Skipped:")
	assert.NotContains(t, stdout, "Remove the imported keys")
}

func TestImport_QuitCancelsCurrentAndRemaining(t *testing.T) {
	setup(t)
	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-abc"
export GITHUB_TOKEN="ghp_xyz1234567890"
export GROQ_API_KEY="gsk_test123456789"
`)
	cmd.SetReadChoice(choiceSequence("y", "q")) // import first, quit on second

	stdout, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stderr, "[1/3] OPENAI_API_KEY")
	assert.Contains(t, stderr, "Imported OPENAI_API_KEY")
	assert.Contains(t, stderr, "[2/3] GITHUB_TOKEN")
	assert.Contains(t, stderr, "Cancelled")

	// stdout summary
	assert.Contains(t, stdout, "1 imported")
	assert.Contains(t, stdout, "2 cancelled")
	assert.Contains(t, stdout, "Cancelled:")
	assert.Contains(t, stdout, "GITHUB_TOKEN")
	assert.Contains(t, stdout, "GROQ_API_KEY")

	// First key was imported
	_, err = testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err)

	// Second and third keys were not imported
	_, err = testStore.Get("GITHUB_TOKEN")
	require.Error(t, err)
	_, err = testStore.Get("GROQ_API_KEY")
	require.Error(t, err)
}

func TestImport_AlreadyRegistered_OverwriteYes(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-proj-old-value")

	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-new-value"`)
	cmd.SetReadChoice(choiceSequence("y"))

	stdout, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stderr, "Already registered in sekret")
	assert.Contains(t, stderr, "Overwritten OPENAI_API_KEY")
	assert.Contains(t, stdout, "1 imported")
	assert.Contains(t, stdout, "Overwritten:")

	// Verify value was updated
	val, err := testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "sk-proj-new-value", val)
}

func TestImport_AlreadyRegistered_OverwriteNo(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-proj-old-value")

	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-new-value"`)
	cmd.SetReadChoice(choiceSequence("n"))

	stdout, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stderr, "Already registered in sekret")
	assert.Contains(t, stderr, "Skipped")
	assert.Contains(t, stdout, "1 skipped")

	// Verify value was NOT updated
	val, err := testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "sk-proj-old-value", val)
}

func TestImport_AlreadyRegistered_DefaultNo(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-proj-old-value")

	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-new-value"`)
	cmd.SetReadChoice(choiceSequence("")) // Enter = default N for overwrite

	stdout, _, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stdout, "1 skipped")

	val, err := testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "sk-proj-old-value", val)
}

func TestImport_DuplicateKeyAcrossFiles(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	writeScanFile(t, dir, "a.sh", `export OPENAI_API_KEY="sk-proj-from-a"`)
	writeScanFile(t, dir, "b.sh", `export OPENAI_API_KEY="sk-proj-from-b"`)

	// Import first, then overwrite prompt for second (skip)
	cmd.SetReadChoice(choiceSequence("y", "n"))

	stdout, stderr, err := executeImport(t, "--file", dir)
	require.NoError(t, err)

	assert.Contains(t, stderr, "[1/2] OPENAI_API_KEY")
	assert.Contains(t, stderr, "Imported OPENAI_API_KEY")
	assert.Contains(t, stderr, "[2/2] OPENAI_API_KEY")
	assert.Contains(t, stderr, "Already registered in sekret")
	assert.Contains(t, stdout, "1 imported")
	assert.Contains(t, stdout, "1 skipped")
}

func TestImport_KeychainSaveFailure(t *testing.T) {
	setup(t)

	// Use a store that fails on Set
	failStore := &failingStore{MockStore: keychain.NewMockStore()}
	cmd.SetStore(failStore)

	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-abc"`)
	cmd.SetReadChoice(choiceSequence("y"))

	stdout, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stderr, "Failed")
	assert.Contains(t, stdout, "1 failed")
	assert.Contains(t, stdout, "Failed:")
	assert.NotContains(t, stdout, "Remove the imported keys")
}

func TestImport_EmptyValue(t *testing.T) {
	setup(t)
	path := writeImportFile(t, `export OPENAI_API_KEY=""`)
	cmd.SetReadChoice(choiceSequence("y"))

	_, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stderr, "[1/1] OPENAI_API_KEY (empty value)")
}

func TestImport_MixedActions(t *testing.T) {
	setup(t)
	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-abc"
export GITHUB_TOKEN="ghp_xyz1234567890"
export GROQ_API_KEY="gsk_test123456789"
`)
	// import first, skip second, import third
	cmd.SetReadChoice(choiceSequence("y", "s", "y"))

	stdout, _, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stdout, "2 imported")
	assert.Contains(t, stdout, "1 skipped")
	assert.Contains(t, stdout, "Imported:")
	assert.Contains(t, stdout, "OPENAI_API_KEY")
	assert.Contains(t, stdout, "GROQ_API_KEY")
	assert.Contains(t, stdout, "Skipped:")
	assert.Contains(t, stdout, "GITHUB_TOKEN")
	assert.Contains(t, stdout, "Remove the imported keys")
}

func TestImport_FileFlag(t *testing.T) {
	setup(t)
	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-abc"`)
	cmd.SetReadChoice(choiceSequence("y"))

	stdout, _, err := executeImport(t, "--file", path)
	require.NoError(t, err)
	assert.Contains(t, stdout, "1 imported")
}

func TestImport_InvalidChoice_Reprompts(t *testing.T) {
	setup(t)
	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-abc"`)
	// First: invalid input, second: valid "y"
	cmd.SetReadChoice(choiceSequence("x", "y"))

	stdout, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stderr, "Invalid choice. Use y, s, or q.")
	assert.Contains(t, stdout, "1 imported")
}

func TestImport_InvalidOverwriteChoice_Reprompts(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-proj-old-value")

	path := writeImportFile(t, `export OPENAI_API_KEY="sk-proj-new-value"`)
	// First: invalid input, second: valid "n"
	cmd.SetReadChoice(choiceSequence("asdf", "n"))

	stdout, stderr, err := executeImport(t, "--file", path)
	require.NoError(t, err)

	assert.Contains(t, stderr, "Invalid choice. Use y or n.")
	assert.Contains(t, stdout, "1 skipped")

	// Value should NOT have been overwritten
	val, err := testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "sk-proj-old-value", val)
}

func TestImport_FileNotFound(t *testing.T) {
	setup(t)
	_, _, err := executeImport(t, "--file", "/nonexistent/path")
	require.Error(t, err)
}

// failingStore wraps MockStore but always fails on Set.
type failingStore struct {
	*keychain.MockStore
}

func (s *failingStore) Set(_ string, _ string) error {
	return fmt.Errorf("keychain unavailable")
}
