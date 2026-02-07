package cmd_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/keychain"
)

var testStore *keychain.MockStore

func setup(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	config.SetPath(dir)
	testStore = keychain.NewMockStore()
	cmd.SetStore(testStore)
	t.Cleanup(func() {
		config.SetPath("")
		cmd.SetStore(keychain.NewOSStore())
		testStore = nil
	})
}

// seedKey adds a key directly to config and keychain store for testing.
func seedKey(t *testing.T, name, envVar, value string) {
	t.Helper()
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if err := cfg.AddKey(name, envVar); err != nil {
		t.Fatalf("failed to add key: %v", err)
	}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	if err := testStore.Set(name, value); err != nil {
		t.Fatalf("failed to set key in store: %v", err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read pipe: %v", err)
	}
	return buf.String()
}

func TestEnv_Empty(t *testing.T) {
	setup(t)

	output := captureStdout(t, func() {
		rootCmd := cmd.RootCmd()
		rootCmd.SetArgs([]string{"env"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("env command failed: %v", err)
		}
	})

	if output != "" {
		t.Errorf("expected empty output, got %q", output)
	}
}

func TestEnv_WithKeys(t *testing.T) {
	setup(t)
	seedKey(t, "openai", "OPENAI_API_KEY", "sk-test123")
	seedKey(t, "anthropic", "ANTHROPIC_API_KEY", "sk-ant-test456")

	output := captureStdout(t, func() {
		rootCmd := cmd.RootCmd()
		rootCmd.SetArgs([]string{"env"})
		if err := rootCmd.Execute(); err != nil {
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

func TestEnv_ShellEscape(t *testing.T) {
	setup(t)
	seedKey(t, "test", "TEST_KEY", `value"with$special`)

	output := captureStdout(t, func() {
		rootCmd := cmd.RootCmd()
		rootCmd.SetArgs([]string{"env"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("env command failed: %v", err)
		}
	})

	if !strings.Contains(output, `export TEST_KEY="value\"with\$special"`) {
		t.Errorf("expected escaped output, got %q", output)
	}
}

func TestList_Empty(t *testing.T) {
	setup(t)

	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs([]string{"list"})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}
}

func TestList_WithKeys(t *testing.T) {
	setup(t)
	seedKey(t, "openai", "OPENAI_API_KEY", "sk-abcdefghijklmnop")

	output := captureStdout(t, func() {
		rootCmd := cmd.RootCmd()
		rootCmd.SetArgs([]string{"list"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("list command failed: %v", err)
		}
	})

	if !strings.Contains(output, "openai") {
		t.Errorf("expected key name in output, got %q", output)
	}
	if !strings.Contains(output, "OPENAI_API_KEY") {
		t.Errorf("expected env var in output, got %q", output)
	}
	// Should show masked value, not full key
	if strings.Contains(output, "sk-abcdefghijklmnop") {
		t.Error("full key should not appear in list output")
	}
}
