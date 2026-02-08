package cmd_test

import (
	"strings"
	"testing"

	"github.com/eazyhozy/sekret/cmd"
)

func TestSet_ExistingKey(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-old-value")
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-new-value-12345", nil
	})

	if err := executeCmd(t, "set", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	val, _ := testStore.Get("OPENAI_API_KEY")
	if val != "sk-new-value-12345" {
		t.Errorf("got %q, want %q", val, "sk-new-value-12345")
	}
}

func TestSet_ViaShorthand(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-old-value")
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-new-value-12345", nil
	})

	// "openai" should resolve to OPENAI_API_KEY via registry
	if err := executeCmd(t, "set", "openai"); err != nil {
		t.Fatalf("set via shorthand failed: %v", err)
	}

	val, _ := testStore.Get("OPENAI_API_KEY")
	if val != "sk-new-value-12345" {
		t.Errorf("got %q, want %q", val, "sk-new-value-12345")
	}
}

func TestSet_LegacyKey(t *testing.T) {
	setup(t)
	seedLegacyKey(t, "openai", "OPENAI_API_KEY", "sk-old-value")
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-new-value-12345", nil
	})

	// Should resolve via env var, and use legacy name as keychain key
	if err := executeCmd(t, "set", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	// Legacy key uses name as keychain key
	val, _ := testStore.Get("openai")
	if val != "sk-new-value-12345" {
		t.Errorf("got %q, want %q", val, "sk-new-value-12345")
	}
}

func TestSet_NonexistentKey(t *testing.T) {
	setup(t)

	err := executeCmd(t, "set", "NONEXISTENT_KEY")
	if err == nil {
		t.Fatal("expected error for nonexistent key, got nil")
	}
	if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSet_EmptyKey(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-old-value")
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "", nil
	})

	err := executeCmd(t, "set", "OPENAI_API_KEY")
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}
