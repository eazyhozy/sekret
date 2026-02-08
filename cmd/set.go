package cmd

import (
	"fmt"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/registry"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <ENV_VAR>",
	Short: "Update an existing API key",
	Args:  cobra.ExactArgs(1),
	RunE:  runSet,
}

func init() {
	rootCmd.AddCommand(setCmd)
}

func runSet(_ *cobra.Command, args []string) error {
	arg := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	entry, err := resolveKey(cfg, arg)
	if err != nil {
		return fmt.Errorf("key %q is not registered (use 'sekret add %s' first)", arg, arg)
	}

	keychainKey := entry.KeychainKey()

	// Show current masked value
	if current, err := store.Get(keychainKey); err == nil {
		_, _ = fmt.Fprintf(rootCmd.ErrOrStderr(), "  Current: %s\n", maskKey(current))
	}

	// Read new key interactively
	value, err := readPassword("  New API Key: ")
	if err != nil {
		return err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Validate format for known keys
	regEntry := registry.LookupByEnvVar(entry.EnvVar)
	if regEntry != nil && !registry.ValidateFormat(regEntry, value) {
		_, _ = fmt.Fprintf(rootCmd.ErrOrStderr(), "  Warning: key does not match expected format for %q (expected prefix: %s)\n",
			entry.EnvVar, strings.Join(regEntry.Prefixes, " or "))
	}

	// Update keychain
	if err := store.Set(keychainKey, value); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(rootCmd.ErrOrStderr(), "  Updated")
	return nil
}
