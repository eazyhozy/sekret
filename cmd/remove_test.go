package cmd_test

import (
	"strings"
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/eazyhozy/sekret/internal/config"
)

func TestRemove_Confirmed(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-to-delete")
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return true, nil
	})

	if err := executeCmd(t, "remove", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	// Verify removed from store
	_, err := testStore.Get("OPENAI_API_KEY")
	if err == nil {
		t.Error("key should be deleted from store")
	}

	// Verify removed from config
	cfg, _ := config.Load()
	if cfg.FindKeyByEnvVar("OPENAI_API_KEY") != nil {
		t.Error("key should be deleted from config")
	}
}

func TestRemove_ViaShorthand(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-to-delete")
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return true, nil
	})

	if err := executeCmd(t, "remove", "openai"); err != nil {
		t.Fatalf("remove via shorthand failed: %v", err)
	}

	_, err := testStore.Get("OPENAI_API_KEY")
	if err == nil {
		t.Error("key should be deleted from store")
	}
}

func TestRemove_LegacyKey(t *testing.T) {
	setup(t)
	seedLegacyKey(t, "openai", "OPENAI_API_KEY", "sk-to-delete")
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return true, nil
	})

	if err := executeCmd(t, "remove", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	// Legacy key stored under name "openai"
	_, err := testStore.Get("openai")
	if err == nil {
		t.Error("key should be deleted from store")
	}

	cfg, _ := config.Load()
	if cfg.FindKeyByEnvVar("OPENAI_API_KEY") != nil {
		t.Error("key should be deleted from config")
	}
}

func TestRemove_Cancelled(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-keep-me")
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return false, nil
	})

	if err := executeCmd(t, "remove", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	// Verify still exists
	val, err := testStore.Get("OPENAI_API_KEY")
	if err != nil {
		t.Fatalf("key should still exist: %v", err)
	}
	if val != "sk-keep-me" {
		t.Errorf("got %q, want %q", val, "sk-keep-me")
	}
}

func TestRemove_NonexistentKey(t *testing.T) {
	setup(t)

	err := executeCmd(t, "remove", "NONEXISTENT_KEY")
	if err == nil {
		t.Fatal("expected error for nonexistent key, got nil")
	}
	if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("unexpected error: %v", err)
	}
}
