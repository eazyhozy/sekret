package cmd

import (
	"fmt"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a registered key",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(_ *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	// Check if key exists
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	entry := cfg.FindKey(name)
	if entry == nil {
		return fmt.Errorf("key %q is not registered", name)
	}

	// Confirmation prompt
	confirmed, err := readConfirm(fmt.Sprintf("  Remove '%s' (%s)? [y/N]: ", name, entry.EnvVar))
	if err != nil {
		return err
	}
	if !confirmed {
		_, _ = fmt.Fprintln(rootCmd.ErrOrStderr(), "  Cancelled")
		return nil
	}

	// Delete from keychain
	if err := store.Delete(name); err != nil {
		return err
	}

	// Delete from config
	if err := cfg.RemoveKey(name); err != nil {
		return err
	}
	if err := config.Save(cfg); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(rootCmd.ErrOrStderr(), "  Removed")
	return nil
}
