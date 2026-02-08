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
	cmd.SetReadInput(func(_ string) (string, error) {
		return "", fmt.Errorf("readInput not configured for this test")
	})
	cmd.SetReadConfirm(func(_ string) (bool, error) {
		return false, fmt.Errorf("readConfirm not configured for this test")
	})
	t.Cleanup(func() {
		config.SetPath("")
		cmd.SetStore(keychain.NewOSStore())
		cmd.SetReadPassword(nil)
		cmd.SetReadInput(nil)
		cmd.SetReadConfirm(nil)
		testStore = nil
	})
}

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
