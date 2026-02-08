package scanner

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eazyhozy/sekret/internal/registry"
)

// Finding represents a detected plaintext key in a file.
type Finding struct {
	FilePath string
	Line     int
	EnvVar   string
	Value    string
}

// secretSuffixes are env var name suffixes that indicate a secret value.
var secretSuffixes = []string{"_KEY", "_TOKEN", "_SECRET", "_CREDENTIALS"}

// exportPattern matches `export KEY=VALUE` statements.
var exportPattern = regexp.MustCompile(`^\s*export\s+([A-Za-z_][A-Za-z0-9_]*)=(.+)$`)

// knownPrefixes are API key prefixes used for mask display (longest first).
var knownPrefixes = []string{
	"sk-proj-", "sk-ant-", "github_pat_",
	"sk-", "ghp_", "gsk_", "AIza",
}

// defaultTargetFiles are shell config filenames scanned by default.
var defaultTargetFiles = []string{
	".zshrc", ".zshenv", ".zprofile",
	".bashrc", ".bash_profile",
	".profile",
}

// DefaultTargets returns the default scan target paths in the home directory.
func DefaultTargets() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	targets := make([]string, len(defaultTargetFiles))
	for i, f := range defaultTargetFiles {
		targets[i] = filepath.Join(home, f)
	}
	return targets
}

// ScanFile parses a single file for export statements containing potential secrets.
func ScanFile(path string) ([]Finding, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var findings []Finding
	lineNum := 0
	s := bufio.NewScanner(f)

	for s.Scan() {
		lineNum++
		line := s.Text()

		// Skip comment lines
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		matches := exportPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		envVar := matches[1]
		value := unquote(matches[2])

		if !IsSecretEnvVar(envVar) {
			continue
		}

		findings = append(findings, Finding{
			FilePath: path,
			Line:     lineNum,
			EnvVar:   envVar,
			Value:    value,
		})
	}

	if err := s.Err(); err != nil {
		return nil, err
	}
	return findings, nil
}

// ScanFiles scans multiple files and returns all findings.
// Files that do not exist are silently skipped.
func ScanFiles(paths []string) ([]Finding, error) {
	var all []Finding
	for _, path := range paths {
		findings, err := ScanFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		all = append(all, findings...)
	}
	return all, nil
}

// IsSecretEnvVar returns true if the env var name matches a known registry entry
// or ends with a secret-indicating suffix.
func IsSecretEnvVar(envVar string) bool {
	if registry.LookupByEnvVar(envVar) != nil {
		return true
	}
	upper := strings.ToUpper(envVar)
	for _, suffix := range secretSuffixes {
		if strings.HasSuffix(upper, suffix) {
			return true
		}
	}
	return false
}

// MaskValue masks a secret value for display, showing a recognized prefix
// (or a short prefix) and the last 4 characters.
func MaskValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}

	prefix := ""
	for _, p := range knownPrefixes {
		if strings.HasPrefix(value, p) {
			prefix = p
			break
		}
	}

	if prefix == "" {
		if len(value) > 8 {
			prefix = value[:4]
		} else {
			prefix = value[:2]
		}
	}

	suffix := value[len(value)-4:]
	return prefix + "..." + suffix
}

// ResolvePath resolves a --path argument to a list of scannable file paths.
// If path is a file, returns it as a single-element slice.
// If path is a directory, returns all regular files in it (non-recursive).
func ResolvePath(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return []string{path}, nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		paths = append(paths, filepath.Join(path, e.Name()))
	}
	return paths, nil
}

// unquote removes surrounding double or single quotes from a value.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
