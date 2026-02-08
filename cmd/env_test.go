package cmd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv_Empty(t *testing.T) {
	setup(t)

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "env"))
	})

	assert.Empty(t, output)
}

func TestEnv_WithKeys(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-test123")
	seedKey(t, "ANTHROPIC_API_KEY", "sk-ant-test456")

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "env"))
	})

	assert.Contains(t, output, `export OPENAI_API_KEY="sk-test123"`)
	assert.Contains(t, output, `export ANTHROPIC_API_KEY="sk-ant-test456"`)
}

func TestEnv_LegacyKeys(t *testing.T) {
	setup(t)
	seedLegacyKey(t, "openai", "OPENAI_API_KEY", "sk-test123")

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "env"))
	})

	assert.Contains(t, output, `export OPENAI_API_KEY="sk-test123"`)
}

func TestEnv_ShellEscape(t *testing.T) {
	setup(t)
	seedKey(t, "TEST_KEY", `value"with$special`)

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "env"))
	})

	assert.Contains(t, output, `export TEST_KEY="value\"with\$special"`)
}
