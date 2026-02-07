package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/registry"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var addEnvFlag string

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Register a new API key",
	Long: `Register a new API key in the OS keychain.

For built-in keys (openai, anthropic, gemini, github, groq),
the environment variable name is automatically mapped.

For custom keys, use the --env flag:
  sekret add my-service --env MY_SERVICE_KEY`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addEnvFlag, "env", "", "environment variable name (required for custom keys)")
	rootCmd.AddCommand(addCmd)
}

func isValidName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func isValidEnvVar(envVar string) bool {
	if envVar == "" {
		return false
	}
	for i, c := range envVar {
		if i == 0 && c >= '0' && c <= '9' {
			return false
		}
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

func runAdd(cmd *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	if !isValidName(name) {
		return fmt.Errorf("invalid key name %q: use only lowercase letters, numbers, hyphens, and underscores", name)
	}

	// Determine env var name
	envVar := addEnvFlag
	entry := registry.Lookup(name)
	if envVar == "" {
		if entry == nil {
			return fmt.Errorf("unknown key %q: use --env to specify the environment variable name", name)
		}
		envVar = entry.EnvVar
	}

	if !isValidEnvVar(envVar) {
		return fmt.Errorf("invalid environment variable name %q: use only letters, numbers, and underscores (cannot start with a number)", envVar)
	}

	// Check if already registered
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.FindKey(name) != nil {
		return fmt.Errorf("key %q is already registered (use 'sekret set %s' to update)", name, name)
	}

	// Read key interactively
	fmt.Fprint(os.Stderr, "  API Key: ")
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
	if entry != nil && !registry.ValidateFormat(entry, value) {
		fmt.Fprintf(os.Stderr, "  Warning: key does not match expected format for %q (expected prefix: %s)\n",
			name, strings.Join(entry.Prefixes, " or "))
	}

	// Save to keychain
	if err := store.Set(name, value); err != nil {
		return err
	}

	// Save metadata to config
	if err := cfg.AddKey(name, envVar); err != nil {
		return err
	}
	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "  Saved to OS keychain (%s)\n", envVar)
	return nil
}
