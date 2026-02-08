package cmd

import (
	"fmt"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <ENV_VAR>",
	Short: "Remove a registered key",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(_ *cobra.Command, args []string) error {
	arg := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	entry, err := resolveKey(cfg, arg)
	if err != nil {
		return fmt.Errorf("key %q is not registered", arg)
	}

	// Confirmation prompt
	confirmed, err := readConfirm(fmt.Sprintf("  Remove '%s'? [y/N]: ", entry.EnvVar))
	if err != nil {
		return err
	}
	if !confirmed {
		_, _ = fmt.Fprintln(rootCmd.ErrOrStderr(), "  Cancelled")
		return nil
	}

	// Delete from keychain
	if err := store.Delete(entry.KeychainKey()); err != nil {
		return err
	}

	// Delete from config (by env var)
	if err := cfg.RemoveKey(entry.EnvVar); err != nil {
		return err
	}
	if err := config.Save(cfg); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(rootCmd.ErrOrStderr(), "  Removed")
	return nil
}
