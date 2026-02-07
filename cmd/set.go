package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/keychain"
	"github.com/eazyhozy/sekret/internal/registry"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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
	store := keychain.NewOSStore()
	if current, err := store.Get(name); err == nil {
		fmt.Fprintf(os.Stderr, "  Current: %s\n", maskKey(current))
	}

	// Read new key interactively
	fmt.Fprint(os.Stderr, "  New API Key: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to read key: %w", err)
	}

	value := strings.TrimSpace(string(password))
	if value == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Validate format for known keys
	regEntry := registry.Lookup(name)
	if regEntry != nil && !registry.ValidateFormat(regEntry, value) {
		fmt.Fprintf(os.Stderr, "  Warning: key does not match expected format for %q (expected prefix: %s)\n",
			name, strings.Join(regEntry.Prefixes, " or "))
	}

	// Update keychain
	if err := store.Set(name, value); err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr, "  Updated")
	return nil
}
