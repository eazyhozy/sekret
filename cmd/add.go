package cmd

import (
	"fmt"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/registry"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <ENV_VAR>",
	Short: "Register a new API key",
	Long:  buildAddLong(),
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func buildAddLong() string {
	var b strings.Builder
	b.WriteString(`Register a new API key in the OS keychain.

Accepts an environment variable name directly:
  sekret add OPENAI_API_KEY

Built-in shorthands:`)
	for _, e := range registry.All() {
		b.WriteString(fmt.Sprintf("\n  %-12s -> %s", e.Name, e.EnvVar))
	}
	return b.String()
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(_ *cobra.Command, args []string) error {
	arg := args[0]

	// Resolve argument to env var
	envVar, regEntry, err := resolveEnvVar(arg)
	if err != nil {
		return err
	}

	// If shorthand was used, confirm with user
	if regEntry != nil && arg != envVar {
		confirmed, err := readConfirm(fmt.Sprintf("  %q -> %s. Continue? [Y/n]: ", arg, envVar))
		if err != nil {
			return err
		}
		if !confirmed {
			_, _ = fmt.Fprintln(rootCmd.ErrOrStderr(), "  Cancelled")
			return nil
		}
	}

	// Check if env var already registered (before any key input)
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.FindKeyByEnvVar(envVar) != nil {
		return fmt.Errorf("key %q is already registered (use 'sekret set %s' to update)", envVar, envVar)
	}

	// Read key interactively
	value, err := readPassword("  API Key: ")
	if err != nil {
		return err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Validate format for known keys
	if regEntry == nil {
		regEntry = registry.LookupByEnvVar(envVar)
	}
	if regEntry != nil && !registry.ValidateFormat(regEntry, value) {
		_, _ = fmt.Fprintf(rootCmd.ErrOrStderr(), "  Warning: key does not match expected format for %q (expected prefix: %s)\n",
			envVar, strings.Join(regEntry.Prefixes, " or "))
	}

	// Save to keychain (using env var as the keychain key)
	if err := store.Set(envVar, value); err != nil {
		return err
	}

	// Save metadata to config (name is empty for new entries)
	if err := cfg.AddKey("", envVar); err != nil {
		return err
	}
	if err := config.Save(cfg); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(rootCmd.ErrOrStderr(), "  Saved to OS keychain (%s)\n", envVar)
	return nil
}
