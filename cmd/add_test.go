package cmd_test

import (
	"strings"
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/eazyhozy/sekret/internal/config"
)

func TestAdd_BuiltinKey(t *testing.T) {
	setup(t)
	cmd.SetReadInput(func(_ string) (string, error) {
		return "", nil // accept default env var
	})
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-test-key-12345678", nil
	})

	if err := executeCmd(t, "add", "openai"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	val, err := testStore.Get("openai")
	if err != nil {
		t.Fatalf("key not found in store: %v", err)
	}
	if val != "sk-test-key-12345678" {
		t.Errorf("got %q, want %q", val, "sk-test-key-12345678")
	}

	cfg, _ := config.Load()
	entry := cfg.FindKey("openai")
	if entry == nil {
		t.Fatal("key not found in config")
	}
	if entry.EnvVar != "OPENAI_API_KEY" {
		t.Errorf("got env var %q, want %q", entry.EnvVar, "OPENAI_API_KEY")
	}
}

func TestAdd_BuiltinKeyCustomEnv(t *testing.T) {
	setup(t)
	cmd.SetReadInput(func(_ string) (string, error) {
		return "MY_OPENAI_KEY", nil // override default
	})
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-test-key-12345678", nil
	})

	if err := executeCmd(t, "add", "openai"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	cfg, _ := config.Load()
	entry := cfg.FindKey("openai")
	if entry == nil {
		t.Fatal("key not found in config")
	}
	if entry.EnvVar != "MY_OPENAI_KEY" {
		t.Errorf("got env var %q, want %q", entry.EnvVar, "MY_OPENAI_KEY")
	}
}

func TestAdd_CustomKey(t *testing.T) {
	setup(t)
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "my-secret-value", nil
	})

	if err := executeCmd(t, "add", "my-service", "--env", "MY_SERVICE_KEY"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	cfg, _ := config.Load()
	entry := cfg.FindKey("my-service")
	if entry == nil {
		t.Fatal("key not found in config")
	}
	if entry.EnvVar != "MY_SERVICE_KEY" {
		t.Errorf("got env var %q, want %q", entry.EnvVar, "MY_SERVICE_KEY")
	}
}

func TestAdd_UnknownKeyInteractive(t *testing.T) {
	setup(t)
	cmd.SetReadInput(func(_ string) (string, error) {
		return "CUSTOM_SERVICE_KEY", nil
	})
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "my-secret-value", nil
	})

	if err := executeCmd(t, "add", "custom-service"); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	cfg, _ := config.Load()
	entry := cfg.FindKey("custom-service")
	if entry == nil {
		t.Fatal("key not found in config")
	}
	if entry.EnvVar != "CUSTOM_SERVICE_KEY" {
		t.Errorf("got env var %q, want %q", entry.EnvVar, "CUSTOM_SERVICE_KEY")
	}
}

func TestAdd_UnknownKeyEmptyEnvVar(t *testing.T) {
	setup(t)
	cmd.SetReadInput(func(_ string) (string, error) {
		return "", nil // empty input for unknown key
	})

	err := executeCmd(t, "add", "unknown-service")
	if err == nil {
		t.Fatal("expected error for empty env var, got nil")
	}
	if !strings.Contains(err.Error(), "environment variable name is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdd_DuplicateKey(t *testing.T) {
	setup(t)
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-test-key", nil
	})
	seedKey(t, "openai", "OPENAI_API_KEY", "sk-existing")

	err := executeCmd(t, "add", "openai")
	if err == nil {
		t.Fatal("expected error for duplicate key, got nil")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdd_InvalidName(t *testing.T) {
	setup(t)

	err := executeCmd(t, "add", "INVALID NAME!")
	if err == nil {
		t.Fatal("expected error for invalid name, got nil")
	}
	if !strings.Contains(err.Error(), "invalid key name") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdd_DuplicateEnvVar(t *testing.T) {
	setup(t)
	cmd.SetReadInput(func(_ string) (string, error) {
		return "", nil // accept default
	})
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "my-secret-value", nil
	})
	seedKey(t, "openai", "OPENAI_API_KEY", "sk-existing")

	err := executeCmd(t, "add", "my-key", "--env", "OPENAI_API_KEY")
	if err == nil {
		t.Fatal("expected error for duplicate env var, got nil")
	}
	if !strings.Contains(err.Error(), "already used by") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAdd_EmptyKey(t *testing.T) {
	setup(t)
	cmd.SetReadInput(func(_ string) (string, error) {
		return "", nil // accept default
	})
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "   ", nil
	})

	err := executeCmd(t, "add", "openai")
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("unexpected error: %v", err)
	}
}
