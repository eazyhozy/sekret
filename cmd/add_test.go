package cmd_test

import (
	"strings"
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/eazyhozy/sekret/internal/config"
)

func TestAdd_EnvVarDirect(t *testing.T) {
	setup(t)
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-test-key-12345678", nil
	})

	if err := executeCmd(t, "add", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	// Stored under env var key
	val, err := testStore.Get("OPENAI_API_KEY")
	if err != nil {
		t.Fatalf("key not found in store: %v", err)
	}
	if val != "sk-test-key-12345678" {
		t.Errorf("got %q, want %q", val, "sk-test-key-12345678")
	}

	cfg, _ := config.Load()
	entry := cfg.FindKeyByEnvVar("OPENAI_API_KEY")
	if entry == nil {
		t.Fatal("key not found in config")
	}
	if entry.Name != "" {
		t.Errorf("expected empty name, got %q", entry.Name)
	}
}

func TestAdd_Shorthand(t *testing.T) {
	setup(t)
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return true, nil // confirm shorthand expansion
	})
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-test-key-12345678", nil
	})

	if err := executeCmd(t, "add", "openai"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	val, err := testStore.Get("OPENAI_API_KEY")
	if err != nil {
		t.Fatalf("key not found in store: %v", err)
	}
	if val != "sk-test-key-12345678" {
		t.Errorf("got %q, want %q", val, "sk-test-key-12345678")
	}

	cfg, _ := config.Load()
	entry := cfg.FindKeyByEnvVar("OPENAI_API_KEY")
	if entry == nil {
		t.Fatal("key not found in config")
	}
}

func TestAdd_ShorthandCancelled(t *testing.T) {
	setup(t)
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return false, nil // reject shorthand expansion
	})

	if err := executeCmd(t, "add", "openai"); err != nil {
		t.Fatalf("add should not return error on cancel: %v", err)
	}

	// Should NOT be registered
	cfg, _ := config.Load()
	if cfg.FindKeyByEnvVar("OPENAI_API_KEY") != nil {
		t.Error("key should not be registered after cancel")
	}
}

func TestAdd_CustomEnvVar(t *testing.T) {
	setup(t)
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "my-secret-value", nil
	})

	if err := executeCmd(t, "add", "MY_SERVICE_KEY"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	cfg, _ := config.Load()
	entry := cfg.FindKeyByEnvVar("MY_SERVICE_KEY")
	if entry == nil {
		t.Fatal("key not found in config")
	}
}

func TestAdd_DuplicateEnvVar(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-existing")

	err := executeCmd(t, "add", "OPENAI_API_KEY")
	if err == nil {
		t.Fatal("expected error for duplicate env var, got nil")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdd_InvalidEnvVarName(t *testing.T) {
	setup(t)

	err := executeCmd(t, "add", "not-a-valid-thing")
	if err == nil {
		t.Fatal("expected error for invalid env var name, got nil")
	}
	if !strings.Contains(err.Error(), "invalid environment variable name") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdd_EmptyKey(t *testing.T) {
	setup(t)
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "   ", nil
	})

	err := executeCmd(t, "add", "OPENAI_API_KEY")
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}
