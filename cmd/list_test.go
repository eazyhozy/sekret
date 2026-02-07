package cmd_test

import (
	"strings"
	"testing"
)

func TestList_Empty(t *testing.T) {
	setup(t)

	if err := executeCmd(t, "list"); err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

func TestList_WithKeys(t *testing.T) {
	setup(t)
	seedKey(t, "openai", "OPENAI_API_KEY", "sk-abcdefghijklmnop")

	output := captureStdout(t, func() {
		if err := executeCmd(t, "list"); err != nil {
			t.Fatalf("list command failed: %v", err)
		}
	})

	if !strings.Contains(output, "openai") {
		t.Errorf("expected key name in output, got %q", output)
	}
	if !strings.Contains(output, "OPENAI_API_KEY") {
		t.Errorf("expected env var in output, got %q", output)
	}
	if strings.Contains(output, "sk-abcdefghijklmnop") {
		t.Error("full key should not appear in list output")
	}
}
