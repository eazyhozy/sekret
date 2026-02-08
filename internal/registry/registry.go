package registry

import "strings"

// Entry defines a known key with its env var name and format prefixes.
type Entry struct {
	Name     string
	EnvVar   string
	Prefixes []string
}

var builtinEntries = []Entry{
	{Name: "openai", EnvVar: "OPENAI_API_KEY", Prefixes: []string{"sk-", "sk-proj-"}},
	{Name: "anthropic", EnvVar: "ANTHROPIC_API_KEY", Prefixes: []string{"sk-ant-"}},
	{Name: "gemini", EnvVar: "GEMINI_API_KEY", Prefixes: []string{"AIza"}},
	{Name: "github", EnvVar: "GITHUB_TOKEN", Prefixes: []string{"ghp_", "github_pat_"}},
	{Name: "groq", EnvVar: "GROQ_API_KEY", Prefixes: []string{"gsk_"}},
}

// All returns all built-in registry entries.
func All() []Entry {
	return builtinEntries
}

// Lookup returns the registry entry for the given name, or nil if not found.
func Lookup(name string) *Entry {
	lower := strings.ToLower(name)
	for i := range builtinEntries {
		if builtinEntries[i].Name == lower {
			return &builtinEntries[i]
		}
	}
	return nil
}

// LookupByEnvVar returns the registry entry for the given env var, or nil if not found.
func LookupByEnvVar(envVar string) *Entry {
	for i := range builtinEntries {
		if builtinEntries[i].EnvVar == envVar {
			return &builtinEntries[i]
		}
	}
	return nil
}

// ValidateFormat checks whether the value matches any known prefix for the entry.
// Returns true if the entry has no prefixes or if the value matches at least one.
func ValidateFormat(entry *Entry, value string) bool {
	if entry == nil || len(entry.Prefixes) == 0 {
		return true
	}
	for _, prefix := range entry.Prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
