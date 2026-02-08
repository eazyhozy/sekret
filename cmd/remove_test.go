package cmd_test

import (
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/eazyhozy/sekret/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemove_Confirmed(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-to-delete")
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return true, nil
	})

	require.NoError(t, executeCmd(t, "remove", "OPENAI_API_KEY"))

	_, err := testStore.Get("OPENAI_API_KEY")
	assert.Error(t, err, "key should be deleted from store")

	cfg, _ := config.Load()
	assert.Nil(t, cfg.FindKeyByEnvVar("OPENAI_API_KEY"), "key should be deleted from config")
}

func TestRemove_ViaShorthand(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-to-delete")
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return true, nil
	})

	require.NoError(t, executeCmd(t, "remove", "openai"))

	_, err := testStore.Get("OPENAI_API_KEY")
	assert.Error(t, err, "key should be deleted from store")
}

func TestRemove_LegacyKey(t *testing.T) {
	setup(t)
	seedLegacyKey(t, "openai", "OPENAI_API_KEY", "sk-to-delete")
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return true, nil
	})

	require.NoError(t, executeCmd(t, "remove", "OPENAI_API_KEY"))

	_, err := testStore.Get("openai")
	assert.Error(t, err, "key should be deleted from store")

	cfg, _ := config.Load()
	assert.Nil(t, cfg.FindKeyByEnvVar("OPENAI_API_KEY"), "key should be deleted from config")
}

func TestRemove_Cancelled(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-keep-me")
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return false, nil
	})

	require.NoError(t, executeCmd(t, "remove", "OPENAI_API_KEY"))

	val, err := testStore.Get("OPENAI_API_KEY")
	require.NoError(t, err, "key should still exist")
	assert.Equal(t, "sk-keep-me", val)
}

func TestRemove_NonexistentKey(t *testing.T) {
	setup(t)

	err := executeCmd(t, "remove", "NONEXISTENT_KEY")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}
