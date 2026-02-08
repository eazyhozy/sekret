package cmd_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList_Empty(t *testing.T) {
	setup(t)
	require.NoError(t, executeCmd(t, "list"))
}

func TestList_WithKeys(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-abcdefghijklmnop")

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "list"))
	})

	assert.Contains(t, output, "OPENAI_API_KEY")
	assert.NotContains(t, output, "sk-abcdefghijklmnop")
}

func TestList_NoNameColumn(t *testing.T) {
	setup(t)
	seedKey(t, "OPENAI_API_KEY", "sk-abcdefghijklmnop")

	output := captureStdout(t, func() {
		require.NoError(t, executeCmd(t, "list"))
	})

	lines := strings.Split(output, "\n")
	require.NotEmpty(t, lines)
	assert.False(t, strings.HasPrefix(strings.TrimSpace(lines[0]), "Name"))
	assert.True(t, strings.HasPrefix(strings.TrimSpace(lines[0]), "Env Variable"))
}
