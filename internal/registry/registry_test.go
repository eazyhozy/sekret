package registry_test

import (
	"testing"

	"github.com/eazyhozy/sekret/internal/registry"
)

func TestLookup_BuiltinKeys(t *testing.T) {
	tests := []struct {
		name      string
		wantEnv   string
		wantFound bool
	}{
		{"openai", "OPENAI_API_KEY", true},
		{"anthropic", "ANTHROPIC_API_KEY", true},
		{"gemini", "GEMINI_API_KEY", true},
		{"github", "GITHUB_TOKEN", true},
		{"groq", "GROQ_API_KEY", true},
		{"unknown-service", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := registry.Lookup(tt.name)
			if tt.wantFound {
				if entry == nil {
					t.Fatalf("expected entry for %q, got nil", tt.name)
				}
				if entry.EnvVar != tt.wantEnv {
					t.Errorf("got env var %q, want %q", entry.EnvVar, tt.wantEnv)
				}
			} else {
				if entry != nil {
					t.Errorf("expected nil for %q, got %+v", tt.name, entry)
				}
			}
		})
	}
}

func TestLookup_CaseInsensitive(t *testing.T) {
	entry := registry.Lookup("OpenAI")
	if entry == nil {
		t.Fatal("expected entry for 'OpenAI', got nil")
	}
	if entry.EnvVar != "OPENAI_API_KEY" {
		t.Errorf("got %q, want %q", entry.EnvVar, "OPENAI_API_KEY")
	}
}

func TestValidateFormat(t *testing.T) {
	openai := registry.Lookup("openai")

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"valid sk- prefix", "sk-abc123", true},
		{"valid sk-proj- prefix", "sk-proj-abc123", true},
		{"invalid prefix", "invalid-key", false},
		{"empty value", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.ValidateFormat(openai, tt.value)
			if got != tt.want {
				t.Errorf("ValidateFormat(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValidateFormat_NilEntry(t *testing.T) {
	if !registry.ValidateFormat(nil, "anything") {
		t.Error("expected true for nil entry")
	}
}

func TestLookupByEnvVar(t *testing.T) {
	tests := []struct {
		envVar    string
		wantName  string
		wantFound bool
	}{
		{"OPENAI_API_KEY", "openai", true},
		{"ANTHROPIC_API_KEY", "anthropic", true},
		{"GEMINI_API_KEY", "gemini", true},
		{"GITHUB_TOKEN", "github", true},
		{"GROQ_API_KEY", "groq", true},
		{"UNKNOWN_KEY", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.envVar, func(t *testing.T) {
			entry := registry.LookupByEnvVar(tt.envVar)
			if tt.wantFound {
				if entry == nil {
					t.Fatalf("expected entry for %q, got nil", tt.envVar)
				}
				if entry.Name != tt.wantName {
					t.Errorf("got name %q, want %q", entry.Name, tt.wantName)
				}
			} else {
				if entry != nil {
					t.Errorf("expected nil for %q, got %+v", tt.envVar, entry)
				}
			}
		})
	}
}
