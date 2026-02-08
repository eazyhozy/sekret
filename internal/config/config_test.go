package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	assert.Equal(t, 1, cfg.Version)
	assert.Empty(t, cfg.Keys)
}

func TestSaveAndLoad(t *testing.T) {
	setupTestDir(t)

	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	require.NoError(t, cfg.AddKey("openai", "OPENAI_API_KEY"))
	require.NoError(t, config.Save(cfg))

	loaded, err := config.Load()
	require.NoError(t, err)
	require.Len(t, loaded.Keys, 1)
	assert.Equal(t, "openai", loaded.Keys[0].Name)
	assert.Equal(t, "OPENAI_API_KEY", loaded.Keys[0].EnvVar)
}

func TestAddKey_Duplicate(t *testing.T) {
	setupTestDir(t)

	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	require.NoError(t, cfg.AddKey("openai", "OPENAI_API_KEY"))

	err := cfg.AddKey("openai", "OPENAI_API_KEY")
	assert.Error(t, err)
}

func TestAddKey_DuplicateEnvVar(t *testing.T) {
	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	require.NoError(t, cfg.AddKey("openai", "OPENAI_API_KEY"))

	err := cfg.AddKey("my-key", "OPENAI_API_KEY")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already used by")
}

func TestRemoveKey_ByEnvVar(t *testing.T) {
	setupTestDir(t)

	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	require.NoError(t, cfg.AddKey("openai", "OPENAI_API_KEY"))
	require.NoError(t, cfg.RemoveKey("OPENAI_API_KEY"))
	assert.Empty(t, cfg.Keys)
}

func TestRemoveKey_NotFound(t *testing.T) {
	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	err := cfg.RemoveKey("NONEXISTENT")
	assert.Error(t, err)
}

func TestKeychainKey_WithName(t *testing.T) {
	entry := &config.KeyEntry{Name: "openai", EnvVar: "OPENAI_API_KEY"}
	assert.Equal(t, "openai", entry.KeychainKey())
}

func TestKeychainKey_WithoutName(t *testing.T) {
	entry := &config.KeyEntry{Name: "", EnvVar: "MY_SERVICE_KEY"}
	assert.Equal(t, "MY_SERVICE_KEY", entry.KeychainKey())
}

func TestFindKey(t *testing.T) {
	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	require.NoError(t, cfg.AddKey("openai", "OPENAI_API_KEY"))

	entry := cfg.FindKey("openai")
	require.NotNil(t, entry)
	assert.Equal(t, "OPENAI_API_KEY", entry.EnvVar)
	assert.Nil(t, cfg.FindKey("nonexistent"))
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested", "path")
	config.SetPath(nested)
	t.Cleanup(func() { config.SetPath("") })

	cfg := &config.Config{Version: 1, Keys: []config.KeyEntry{}}
	require.NoError(t, config.Save(cfg))

	_, err := os.Stat(filepath.Join(nested, "config.json"))
	assert.NoError(t, err)
}
