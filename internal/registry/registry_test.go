package registry_test

import (
	"testing"

	"github.com/eazyhozy/sekret/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.NotNil(t, entry, "expected entry for %q", tt.name)
				assert.Equal(t, tt.wantEnv, entry.EnvVar)
			} else {
				assert.Nil(t, entry, "expected nil for %q", tt.name)
			}
		})
	}
}

func TestLookup_CaseInsensitive(t *testing.T) {
	entry := registry.Lookup("OpenAI")
	require.NotNil(t, entry)
	assert.Equal(t, "OPENAI_API_KEY", entry.EnvVar)
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
			assert.Equal(t, tt.want, registry.ValidateFormat(openai, tt.value))
		})
	}
}

func TestValidateFormat_NilEntry(t *testing.T) {
	assert.True(t, registry.ValidateFormat(nil, "anything"))
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
				require.NotNil(t, entry, "expected entry for %q", tt.envVar)
				assert.Equal(t, tt.wantName, entry.Name)
			} else {
				assert.Nil(t, entry, "expected nil for %q", tt.envVar)
			}
		})
	}
}
