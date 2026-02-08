package cmd_test

import (
	"strings"
	"testing"
)

func TestEnv_Empty(t *testing.T) {
	setup(t)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "env"); err != nil {
			t.Fatalf("env command failed: %v", err)
		}
	})

	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}
}

func TestEnv_WithKeys(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-test123")
	seedKey(t, "ANTHROPIC_API_KEY", "sk-ant-test456")

	output := captureStdout(t, func() {
		if err := executeCmd(t, "env"); err != nil {
			t.Fatalf("env command failed: %v", err)
		}
	})

	if !strings.Contains(output, `export OPENAI_API_KEY="sk-test123"`) {
		t.Errorf("expected OPENAI_API_KEY export, got %q", output)
	}
	if !strings.Contains(output, `export ANTHROPIC_API_KEY="sk-ant-test456"`) {
		t.Errorf("expected ANTHROPIC_API_KEY export, got %q", output)
	}
}

func TestEnv_LegacyKeys(t *testing.T) {
	setup(t)
	seedLegacyKey(t, "openai", "OPENAI_API_KEY", "sk-test123")

	output := captureStdout(t, func() {
		if err := executeCmd(t, "env"); err != nil {
			t.Fatalf("env command failed: %v", err)
		}
	})

	// Legacy key should still produce correct export using name as keychain key
	if !strings.Contains(output, `export OPENAI_API_KEY="sk-test123"`) {
		t.Errorf("expected OPENAI_API_KEY export, got %q", output)
	}
}

func TestEnv_ShellEscape(t *testing.T) {
	setup(t)
	seedKey(t, "TEST_KEY", `value"with$special`)

	output := captureStdout(t, func() {
		if err := executeCmd(t, "env"); err != nil {
			t.Fatalf("env command failed: %v", err)
		}
	})

	if !strings.Contains(output, `export TEST_KEY="value\"with\$special"`) {
		t.Errorf("expected escaped output, got %q", output)
	}
}
