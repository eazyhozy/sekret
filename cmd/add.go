package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/registry"
	"github.com/spf13/cobra"
)

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
	addCmd.Flags().String("env", "", "environment variable name (required for custom keys)")
	rootCmd.AddCommand(addCmd)
}

var validNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
var validEnvVarPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func runAdd(c *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	if !validNamePattern.MatchString(name) {
		return fmt.Errorf("invalid key name %q: use only lowercase letters, numbers, hyphens, and underscores", name)
	}

	// Determine env var name
	envVar, _ := c.Flags().GetString("env")
	entry := registry.Lookup(name)
	if envVar == "" {
		if entry == nil {
			return fmt.Errorf("unknown key %q: use --env to specify the environment variable name", name)
		}
		envVar = entry.EnvVar
	}

	if !validEnvVarPattern.MatchString(envVar) {
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
	value, err := readPassword("  API Key: ")
	if err != nil {
		return err
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// Validate format for known keys
	if entry != nil && !registry.ValidateFormat(entry, value) {
		_, _ = fmt.Fprintf(rootCmd.ErrOrStderr(), "  Warning: key does not match expected format for %q (expected prefix: %s)\n",
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

	_, _ = fmt.Fprintf(rootCmd.ErrOrStderr(), "  Saved to OS keychain (%s)\n", envVar)
	return nil
}
