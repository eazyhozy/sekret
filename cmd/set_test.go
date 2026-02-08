package cmd_test

import (
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSet_ExistingKey(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-old-value")
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-new-value-12345", nil
	})

	require.NoError(t, executeCmd(t, "set", "OPENAI_API_KEY"))

	val, _ := testStore.Get("OPENAI_API_KEY")
	assert.Equal(t, "sk-new-value-12345", val)
}

func TestSet_ViaShorthand(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-old-value")
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-new-value-12345", nil
	})

	require.NoError(t, executeCmd(t, "set", "openai"))

	val, _ := testStore.Get("OPENAI_API_KEY")
	assert.Equal(t, "sk-new-value-12345", val)
}

func TestSet_LegacyKey(t *testing.T) {
	setup(t)
	seedLegacyKey(t, "openai", "OPENAI_API_KEY", "sk-old-value")
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "sk-new-value-12345", nil
	})

	require.NoError(t, executeCmd(t, "set", "OPENAI_API_KEY"))

	val, _ := testStore.Get("openai")
	assert.Equal(t, "sk-new-value-12345", val)
}

func TestSet_NonexistentKey(t *testing.T) {
	setup(t)

	err := executeCmd(t, "set", "NONEXISTENT_KEY")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestSet_EmptyKey(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-old-value")
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "", nil
	})

	err := executeCmd(t, "set", "OPENAI_API_KEY")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}
