package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*.sh")
	require.NoError(t, err)
	_, err = f.WriteString(content)
	require.NoError(t, err)
	_ = f.Close()
	return f.Name()
}

func TestScanFile_DoubleQuoted(t *testing.T) {
	path := writeTempFile(t, `export OPENAI_API_KEY="sk-proj-abc123"`)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	require.Len(t, findings, 1)

	f := findings[0]
	assert.Equal(t, "OPENAI_API_KEY", f.EnvVar)
	assert.Equal(t, "sk-proj-abc123", f.Value)
	assert.Equal(t, 1, f.Line)
}

func TestScanFile_SingleQuoted(t *testing.T) {
	path := writeTempFile(t, `export ANTHROPIC_API_KEY='sk-ant-secret456'`)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "sk-ant-secret456", findings[0].Value)
}

func TestScanFile_Unquoted(t *testing.T) {
	path := writeTempFile(t, `export GITHUB_TOKEN=ghp_abcdef1234567890`)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "ghp_abcdef1234567890", findings[0].Value)
}

func TestScanFile_SuffixPatterns(t *testing.T) {
	content := `export MY_SERVICE_KEY="abc123secret"
export AUTH_TOKEN="token_value_here"
export DB_SECRET="supersecret1234"
export AWS_CREDENTIALS="cred_value_here"
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	require.Len(t, findings, 4)

	expected := []string{"MY_SERVICE_KEY", "AUTH_TOKEN", "DB_SECRET", "AWS_CREDENTIALS"}
	for i, envVar := range expected {
		assert.Equal(t, envVar, findings[i].EnvVar, "finding %d", i)
	}
}

func TestScanFile_IgnoresComments(t *testing.T) {
	content := `# export OPENAI_API_KEY="sk-proj-abc123"
  # export GITHUB_TOKEN="ghp_abcdef"
export GROQ_API_KEY="gsk_real_key_here"
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "GROQ_API_KEY", findings[0].EnvVar)
}

func TestScanFile_IgnoresNonSecret(t *testing.T) {
	content := `export PATH="/usr/local/bin:$PATH"
export EDITOR="vim"
export LANG="en_US.UTF-8"
export HOME="/Users/test"
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestScanFile_IgnoresNoValue(t *testing.T) {
	content := `export OPENAI_API_KEY
export PATH
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	assert.Empty(t, findings)
}

func TestScanFile_EmptyQuotedValue(t *testing.T) {
	content := `export OPENAI_API_KEY=""
export GITHUB_TOKEN=''
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, "OPENAI_API_KEY", findings[0].EnvVar)
	assert.Equal(t, "", findings[0].Value)
	assert.Equal(t, "GITHUB_TOKEN", findings[1].EnvVar)
	assert.Equal(t, "", findings[1].Value)
}

func TestScanFile_EmptyUnquotedValue(t *testing.T) {
	content := `export OPENAI_API_KEY=
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	require.Len(t, findings, 1)
	assert.Equal(t, "OPENAI_API_KEY", findings[0].EnvVar)
	assert.Equal(t, "", findings[0].Value)
}

func TestScanFile_MultipleFindings(t *testing.T) {
	content := `export EDITOR="vim"
export OPENAI_API_KEY="sk-proj-abc123"
export PATH="/usr/bin"
export GITHUB_TOKEN="ghp_abcdef"
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	require.Len(t, findings, 2)
	assert.Equal(t, 2, findings[0].Line)
	assert.Equal(t, 4, findings[1].Line)
}

func TestScanFile_LeadingWhitespace(t *testing.T) {
	path := writeTempFile(t, `  export OPENAI_API_KEY="sk-proj-abc123"`)

	findings, err := ScanFile(path)
	require.NoError(t, err)
	require.Len(t, findings, 1)
}

func TestScanFile_FileNotFound(t *testing.T) {
	_, err := ScanFile("/nonexistent/path/file.sh")
	require.Error(t, err)
	assert.True(t, os.IsNotExist(err))
}

func TestScanFiles_MergesResults(t *testing.T) {
	path1 := writeTempFile(t, `export OPENAI_API_KEY="sk-proj-abc"`)
	path2 := writeTempFile(t, `export GITHUB_TOKEN="ghp_xyz"`)

	findings, err := ScanFiles([]string{path1, path2})
	require.NoError(t, err)
	assert.Len(t, findings, 2)
}

func TestScanFiles_SkipsMissing(t *testing.T) {
	path := writeTempFile(t, `export OPENAI_API_KEY="sk-proj-abc"`)

	findings, err := ScanFiles([]string{"/nonexistent", path})
	require.NoError(t, err)
	assert.Len(t, findings, 1)
}

func TestIsSecretEnvVar_RegistryMatch(t *testing.T) {
	cases := []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY", "GITHUB_TOKEN", "GROQ_API_KEY"}
	for _, envVar := range cases {
		assert.True(t, IsSecretEnvVar(envVar), "%s should be a secret env var", envVar)
	}
}

func TestIsSecretEnvVar_SuffixMatch(t *testing.T) {
	cases := []string{"MY_API_KEY", "AUTH_TOKEN", "DB_SECRET", "AWS_CREDENTIALS"}
	for _, envVar := range cases {
		assert.True(t, IsSecretEnvVar(envVar), "%s should be a secret env var", envVar)
	}
}

func TestIsSecretEnvVar_NonSecret(t *testing.T) {
	cases := []string{"PATH", "HOME", "EDITOR", "LANG", "GOPATH", "NODE_ENV"}
	for _, envVar := range cases {
		assert.False(t, IsSecretEnvVar(envVar), "%s should NOT be a secret env var", envVar)
	}
}

func TestMaskValue_KnownPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"sk-proj-abcdef1234", "sk-proj-...1234"},
		{"sk-ant-abcdef1234", "sk-ant-...1234"},
		{"ghp_abcdef1234", "ghp_...1234"},
		{"github_pat_abcdef1234", "github_pat_...1234"},
		{"gsk_abcdef1234", "gsk_...1234"},
		{"AIzaSyExample1234", "AIza...1234"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, MaskValue(tt.input), "MaskValue(%q)", tt.input)
	}
}

func TestMaskValue_UnknownPrefix(t *testing.T) {
	// Long value (>8): first 4 + ... + last 4
	assert.Equal(t, "abcd...mnop", MaskValue("abcdefghijklmnop"))
	// Short value (5-8): first 2 + ... + last 4
	assert.Equal(t, "ab...efgh", MaskValue("abcdefgh"))
}

func TestMaskValue_VeryShort(t *testing.T) {
	assert.Equal(t, "****", MaskValue("abcd"))
	assert.Equal(t, "****", MaskValue("ab"))
	assert.Equal(t, "****", MaskValue(""))
}

func TestDefaultTargets(t *testing.T) {
	targets := DefaultTargets()
	require.Len(t, targets, 6)

	home, _ := os.UserHomeDir()
	expectedFiles := []string{".zshrc", ".zshenv", ".zprofile", ".bashrc", ".bash_profile", ".profile"}
	for i, f := range expectedFiles {
		assert.Equal(t, filepath.Join(home, f), targets[i], "target %d", i)
	}
}
