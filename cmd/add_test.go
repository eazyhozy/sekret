package cmd_test

import (
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/eazyhozy/sekret/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd_EnvVarDirect(t *testing.T) {
	setup(t)
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-test-key-12345678", nil
	})

	require.NoError(t, executeCmd(t, "add", "OPENAI_API_KEY"))

	val, err := testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "sk-test-key-12345678", val)

	cfg, _ := config.Load()
	entry := cfg.FindKeyByEnvVar("OPENAI_API_KEY")
	require.NotNil(t, entry)
	assert.Empty(t, entry.Name)
}

func TestAdd_Shorthand(t *testing.T) {
	setup(t)
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return true, nil
	})
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-test-key-12345678", nil
	})

	require.NoError(t, executeCmd(t, "add", "openai"))

	val, err := testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err)
	assert.Equal(t, "sk-test-key-12345678", val)

	cfg, _ := config.Load()
	assert.NotNil(t, cfg.FindKeyByEnvVar("OPENAI_API_KEY"))
}

func TestAdd_ShorthandCancelled(t *testing.T) {
	setup(t)
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return false, nil
	})

	require.NoError(t, executeCmd(t, "add", "openai"))

	cfg, _ := config.Load()
	assert.Nil(t, cfg.FindKeyByEnvVar("OPENAI_API_KEY"))
}

func TestAdd_CustomEnvVar(t *testing.T) {
	setup(t)
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "my-secret-value", nil
	})

	require.NoError(t, executeCmd(t, "add", "MY_SERVICE_KEY"))

	cfg, _ := config.Load()
	assert.NotNil(t, cfg.FindKeyByEnvVar("MY_SERVICE_KEY"))
}

func TestAdd_DuplicateEnvVar(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-existing")

	err := executeCmd(t, "add", "OPENAI_API_KEY")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
}

func TestAdd_InvalidEnvVarName(t *testing.T) {
	setup(t)

	err := executeCmd(t, "add", "not-a-valid-thing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid environment variable name")
}

func TestAdd_EmptyKey(t *testing.T) {
	setup(t)
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "   ", nil
	})

	err := executeCmd(t, "add", "OPENAI_API_KEY")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}
