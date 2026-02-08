package cmd_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/eazyhozy/sekret/cmd"
	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/keychain"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

var testStore *keychain.MockStore

func setup(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	config.SetPath(dir)
	testStore = keychain.NewMockStore()
	cmd.SetStore(testStore)
	cmd.SetReadPassword(func(_ string) (string, error) {
		return "", fmt.Errorf("readPassword not configured for this test")
	})
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return false, fmt.Errorf("readConfirm not configured for this test")
	})
	t.Cleanup(func() {
		config.SetPath("")
		cmd.SetStore(keychain.NewOSStore())
		cmd.SetReadPassword(nil)
		cmd.SetReadConfirm(nil)
		testStore = nil
	})
}

// seedKey creates a new-style key entry (no name, env var as keychain key).
func seedKey(t *testing.T, envVar, value string) {
	t.Helper()
	cfg, err := config.Load()
	require.NoError(t, err, "failed to load config")
	require.NoError(t, cfg.AddKey("", envVar), "failed to add key")
	require.NoError(t, config.Save(cfg), "failed to save config")
	require.NoError(t, testStore.Set(envVar, value), "failed to set key in store")
}

// seedLegacyKey creates a legacy-style key entry (with name as keychain key).
func seedLegacyKey(t *testing.T, name, envVar, value string) {
	t.Helper()
	cfg, err := config.Load()
	require.NoError(t, err, "failed to load config")
	require.NoError(t, cfg.AddKey(name, envVar), "failed to add key")
	require.NoError(t, config.Save(cfg), "failed to save config")
	require.NoError(t, testStore.Set(name, value), "failed to set key in store")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err, "failed to create pipe")
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	require.NoError(t, err, "failed to read pipe")
	return buf.String()
}

func executeCmd(t *testing.T, args ...string) error {
	t.Helper()
	rootCmd := cmd.RootCmd()
	rootCmd.SetArgs(args)
	rootCmd.Flags().Visit(func(f *pflag.Flag) { _ = f.Value.Set(f.DefValue) })
	for _, c := range rootCmd.Commands() {
		c.Flags().Visit(func(f *pflag.Flag) { _ = f.Value.Set(f.DefValue) })
	}
	return rootCmd.Execute()
}
