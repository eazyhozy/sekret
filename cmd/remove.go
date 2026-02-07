package cmd

import (
	"bufio"
	"fmt"
	"os"
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
	fmt.Fprintf(os.Stderr, "  Remove '%s' (%s)? [y/N]: ", name, entry.EnvVar)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		fmt.Fprintln(os.Stderr, "  Cancelled")
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

	fmt.Fprintln(os.Stderr, "  Removed")
	return nil
}
