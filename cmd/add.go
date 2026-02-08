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
the environment variable name is suggested as a default.

Use the --env flag to skip the env var prompt:
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

// promptEnvVar prompts the user for an env var name.
// If a registry entry exists, the entry's env var is suggested as the default.
func promptEnvVar(entry *registry.Entry) (string, error) {
	if entry != nil {
		// Known key: suggest default, allow customization
		input, err := readInput(fmt.Sprintf("  Env variable (press Enter for %q): ", entry.EnvVar))
		if err != nil {
			return "", err
		}
		input = strings.TrimSpace(input)
		if input == "" {
			return entry.EnvVar, nil
		}
		return input, nil
	}

	// Unknown key: require input
	input, err := readInput("  Env variable: ")
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("environment variable name is required")
	}
	return input, nil
}

func runAdd(c *cobra.Command, args []string) error {
	name := strings.ToLower(args[0])

	if !validNamePattern.MatchString(name) {
		return fmt.Errorf("invalid key name %q: use only lowercase letters, numbers, hyphens, and underscores", name)
	}

	// Check if name already registered (before any prompts)
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if cfg.FindKey(name) != nil {
		return fmt.Errorf("key %q is already registered (use 'sekret set %s' to update)", name, name)
	}

	// Determine env var name
	envVar, _ := c.Flags().GetString("env")
	entry := registry.Lookup(name)
	if envVar == "" {
		// No --env flag: prompt interactively
		envVar, err = promptEnvVar(entry)
		if err != nil {
			return err
		}
	}

	if !validEnvVarPattern.MatchString(envVar) {
		return fmt.Errorf("invalid environment variable name %q: use only letters, numbers, and underscores (cannot start with a number)", envVar)
	}

	// Check if env var already used
	if existing := cfg.FindKeyByEnvVar(envVar); existing != nil {
		return fmt.Errorf("environment variable %q is already used by key %q", envVar, existing.Name)
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
