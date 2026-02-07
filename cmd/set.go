package cmd

import (
	"fmt"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/registry"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Update an existing API key",
	Args:  cobra.ExactArgs(1),
	RunE:  runSet,
}

func init() {
	rootCmd.AddCommand(setCmd)
}

func runSet(_ *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	// Check if key exists
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	entry := cfg.FindKey(name)
	if entry == nil {
		return fmt.Errorf("key %q is not registered (use 'sekret add %s' first)", name, name)
	}

	// Show current masked value
	if current, err := store.Get(name); err == nil {
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
	regEntry := registry.Lookup(name)
	if regEntry != nil && !registry.ValidateFormat(regEntry, value) {
		_, _ = fmt.Fprintf(rootCmd.ErrOrStderr(), "  Warning: key does not match expected format for %q (expected prefix: %s)\n",
			name, strings.Join(regEntry.Prefixes, " or "))
	}

	// Update keychain
	if err := store.Set(name, value); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(rootCmd.ErrOrStderr(), "  Updated")
	return nil
}
