package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eazyhozy/sekret/internal/config"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	config.SetPath(dir)
	t.Cleanup(func() { config.SetPath("") })
	return dir
}

func TestLoad_NoFile(t *testing.T) {
	setupTestDir(t)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Version != 1 {
		t.Errorf("got version %d, want 1", cfg.Version)
	}
	if len(cfg.Keys) != 0 {
		t.Errorf("got %d keys, want 0", len(cfg.Keys))
	}
}

func TestSaveAndLoad(t *testing.T) {
	setupTestDir(t)

	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	if err := cfg.AddKey("openai", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("AddKey failed: %v", err)
	}

	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := config.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(loaded.Keys) != 1 {
		t.Fatalf("got %d keys, want 1", len(loaded.Keys))
	}
	if loaded.Keys[0].Name != "openai" {
		t.Errorf("got name %q, want %q", loaded.Keys[0].Name, "openai")
	}
	if loaded.Keys[0].EnvVar != "OPENAI_API_KEY" {
		t.Errorf("got env var %q, want %q", loaded.Keys[0].EnvVar, "OPENAI_API_KEY")
	}
}

func TestAddKey_Duplicate(t *testing.T) {
	setupTestDir(t)

	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	if err := cfg.AddKey("openai", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("first AddKey failed: %v", err)
	}

	err := cfg.AddKey("openai", "OPENAI_API_KEY")
	if err == nil {
		t.Fatal("expected error for duplicate key, got nil")
	}
}

func TestRemoveKey(t *testing.T) {
	setupTestDir(t)

	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	if err := cfg.AddKey("openai", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("AddKey failed: %v", err)
	}

	if err := cfg.RemoveKey("openai"); err != nil {
		t.Fatalf("RemoveKey failed: %v", err)
	}
	if len(cfg.Keys) != 0 {
		t.Errorf("got %d keys, want 0", len(cfg.Keys))
	}
}

func TestRemoveKey_NotFound(t *testing.T) {
	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	err := cfg.RemoveKey("nonexistent")
	if err == nil {
		t.Fatal("expected error for removing nonexistent key, got nil")
	}
}

func TestFindKey(t *testing.T) {
	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	if err := cfg.AddKey("openai", "OPENAI_API_KEY"); err != nil {
		t.Fatalf("AddKey failed: %v", err)
	}

	entry := cfg.FindKey("openai")
	if entry == nil {
		t.Fatal("expected entry, got nil")
	}
	if entry.EnvVar != "OPENAI_API_KEY" {
		t.Errorf("got %q, want %q", entry.EnvVar, "OPENAI_API_KEY")
	}

	if cfg.FindKey("nonexistent") != nil {
		t.Error("expected nil for nonexistent key")
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested", "path")
	config.SetPath(nested)
	t.Cleanup(func() { config.SetPath("") })

	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(nested, "config.json")); err != nil {
		t.Fatalf("config file not created: %v", err)
	}
}
