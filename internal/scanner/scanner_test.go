package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "test-*.sh")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return f.Name()
}

func TestScanFile_DoubleQuoted(t *testing.T) {
	path := writeTempFile(t, `export OPENAI_API_KEY="sk-proj-abc123"`)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	f := findings[0]
	if f.EnvVar != "OPENAI_API_KEY" {
		t.Errorf("expected env var OPENAI_API_KEY, got %s", f.EnvVar)
	}
	if f.Value != "sk-proj-abc123" {
		t.Errorf("expected value sk-proj-abc123, got %s", f.Value)
	}
	if f.Line != 1 {
		t.Errorf("expected line 1, got %d", f.Line)
	}
}

func TestScanFile_SingleQuoted(t *testing.T) {
	path := writeTempFile(t, `export ANTHROPIC_API_KEY='sk-ant-secret456'`)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Value != "sk-ant-secret456" {
		t.Errorf("expected value sk-ant-secret456, got %s", findings[0].Value)
	}
}

func TestScanFile_Unquoted(t *testing.T) {
	path := writeTempFile(t, `export GITHUB_TOKEN=ghp_abcdef1234567890`)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].Value != "ghp_abcdef1234567890" {
		t.Errorf("expected value ghp_abcdef1234567890, got %s", findings[0].Value)
	}
}

func TestScanFile_SuffixPatterns(t *testing.T) {
	content := `export MY_SERVICE_KEY="abc123secret"
export AUTH_TOKEN="token_value_here"
export DB_SECRET="supersecret1234"
export AWS_CREDENTIALS="cred_value_here"
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 4 {
		t.Fatalf("expected 4 findings, got %d", len(findings))
	}

	expected := []string{"MY_SERVICE_KEY", "AUTH_TOKEN", "DB_SECRET", "AWS_CREDENTIALS"}
	for i, envVar := range expected {
		if findings[i].EnvVar != envVar {
			t.Errorf("finding %d: expected %s, got %s", i, envVar, findings[i].EnvVar)
		}
	}
}

func TestScanFile_IgnoresComments(t *testing.T) {
	content := `# export OPENAI_API_KEY="sk-proj-abc123"
  # export GITHUB_TOKEN="ghp_abcdef"
export GROQ_API_KEY="gsk_real_key_here"
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
	if findings[0].EnvVar != "GROQ_API_KEY" {
		t.Errorf("expected GROQ_API_KEY, got %s", findings[0].EnvVar)
	}
}

func TestScanFile_IgnoresNonSecret(t *testing.T) {
	content := `export PATH="/usr/local/bin:$PATH"
export EDITOR="vim"
export LANG="en_US.UTF-8"
export HOME="/Users/test"
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestScanFile_IgnoresNoValue(t *testing.T) {
	content := `export OPENAI_API_KEY
export PATH
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 0 {
		t.Fatalf("expected 0 findings, got %d", len(findings))
	}
}

func TestScanFile_MultipleFindings(t *testing.T) {
	content := `export EDITOR="vim"
export OPENAI_API_KEY="sk-proj-abc123"
export PATH="/usr/bin"
export GITHUB_TOKEN="ghp_abcdef"
`
	path := writeTempFile(t, content)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
	if findings[0].Line != 2 {
		t.Errorf("expected line 2, got %d", findings[0].Line)
	}
	if findings[1].Line != 4 {
		t.Errorf("expected line 4, got %d", findings[1].Line)
	}
}

func TestScanFile_LeadingWhitespace(t *testing.T) {
	path := writeTempFile(t, `  export OPENAI_API_KEY="sk-proj-abc123"`)

	findings, err := ScanFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestScanFile_FileNotFound(t *testing.T) {
	_, err := ScanFile("/nonexistent/path/file.sh")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected not-exist error, got %v", err)
	}
}

func TestScanFiles_MergesResults(t *testing.T) {
	path1 := writeTempFile(t, `export OPENAI_API_KEY="sk-proj-abc"`)
	path2 := writeTempFile(t, `export GITHUB_TOKEN="ghp_xyz"`)

	findings, err := ScanFiles([]string{path1, path2})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}
}

func TestScanFiles_SkipsMissing(t *testing.T) {
	path := writeTempFile(t, `export OPENAI_API_KEY="sk-proj-abc"`)

	findings, err := ScanFiles([]string{"/nonexistent", path})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}

func TestIsSecretEnvVar_RegistryMatch(t *testing.T) {
	cases := []string{"OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY", "GITHUB_TOKEN", "GROQ_API_KEY"}
	for _, envVar := range cases {
		if !IsSecretEnvVar(envVar) {
			t.Errorf("expected %s to be a secret env var", envVar)
		}
	}
}

func TestIsSecretEnvVar_SuffixMatch(t *testing.T) {
	cases := []string{"MY_API_KEY", "AUTH_TOKEN", "DB_SECRET", "AWS_CREDENTIALS"}
	for _, envVar := range cases {
		if !IsSecretEnvVar(envVar) {
			t.Errorf("expected %s to be a secret env var", envVar)
		}
	}
}

func TestIsSecretEnvVar_NonSecret(t *testing.T) {
	cases := []string{"PATH", "HOME", "EDITOR", "LANG", "GOPATH", "NODE_ENV"}
	for _, envVar := range cases {
		if IsSecretEnvVar(envVar) {
			t.Errorf("expected %s to NOT be a secret env var", envVar)
		}
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
		got := MaskValue(tt.input)
		if got != tt.expected {
			t.Errorf("MaskValue(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMaskValue_UnknownPrefix(t *testing.T) {
	// Long value (>8): first 4 + ... + last 4
	got := MaskValue("abcdefghijklmnop")
	if got != "abcd...mnop" {
		t.Errorf("MaskValue long = %q, want %q", got, "abcd...mnop")
	}

	// Short value (5-8): first 2 + ... + last 4
	got = MaskValue("abcdefgh")
	if got != "ab...efgh" {
		t.Errorf("MaskValue short = %q, want %q", got, "ab...efgh")
	}
}

func TestMaskValue_VeryShort(t *testing.T) {
	if got := MaskValue("abcd"); got != "****" {
		t.Errorf("MaskValue(4 chars) = %q, want ****", got)
	}
	if got := MaskValue("ab"); got != "****" {
		t.Errorf("MaskValue(2 chars) = %q, want ****", got)
	}
	if got := MaskValue(""); got != "****" {
		t.Errorf("MaskValue(empty) = %q, want ****", got)
	}
}

func TestDefaultTargets(t *testing.T) {
	targets := DefaultTargets()
	if len(targets) != 6 {
		t.Fatalf("expected 6 targets, got %d", len(targets))
	}

	home, _ := os.UserHomeDir()
	expectedFiles := []string{".zshrc", ".zshenv", ".zprofile", ".bashrc", ".bash_profile", ".profile"}
	for i, f := range expectedFiles {
		expected := filepath.Join(home, f)
		if targets[i] != expected {
			t.Errorf("target %d: expected %s, got %s", i, expected, targets[i])
		}
	}
}
