package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/keychain"
	"github.com/eazyhozy/sekret/internal/registry"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// version is set at build time via ldflags.
var version = "dev"

var validEnvVarPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// store is the keychain store used by all commands.
// Override with SetStore() for testing.
var store keychain.Store = keychain.NewOSStore()

// SetStore overrides the keychain store (for testing).
func SetStore(s keychain.Store) {
	store = s
}

// readPassword reads a secret from the terminal with the given prompt.
// Override with SetReadPassword() for testing.
var readPassword = func(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return string(password), nil
}

// SetReadPassword overrides the password reader (for testing).
func SetReadPassword(fn func(string) (string, error)) {
	readPassword = fn
}

// readConfirm reads a y/N confirmation from the user.
// Override with SetReadConfirm() for testing.
var readConfirm = func(prompt string) (bool, error) {
	fmt.Fprint(os.Stderr, prompt)
	var answer string
	if _, err := fmt.Fscanln(os.Stdin, &answer); err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}
	return answer == "y" || answer == "yes", nil
}

// SetReadConfirm overrides the confirm reader (for testing).
func SetReadConfirm(fn func(string) (bool, error)) {
	readConfirm = fn
}

// readChoice reads a single-line choice from the user.
// Override with SetReadChoice() for testing.
var readChoice = func(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	var answer string
	if _, err := fmt.Fscanln(os.Stdin, &answer); err != nil {
		// EOF or empty input (just Enter) — return empty string
		return "", nil
	}
	return answer, nil
}

// SetReadChoice overrides the choice reader (for testing).
func SetReadChoice(fn func(string) (string, error)) {
	readChoice = fn
}

var rootCmd = &cobra.Command{
	Use:     "sekret",
	Version: version,
	Short:   "Secure your API keys in OS keychain, load them as env vars",
	Long: `sekret stores your API keys in the OS keychain and injects them
as environment variables. No more plaintext secrets in .zshrc.

Add 'eval "$(sekret env)"' to your .zshrc to automatically load
all registered keys when opening a new terminal.`,
}

// RootCmd returns the root command for testing.
func RootCmd() *cobra.Command {
	return rootCmd
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// resolveEnvVar resolves a CLI argument to an env var name.
// Built-in shorthand names (openai, anthropic, etc.) are expanded to their env var.
// Otherwise the argument is validated as an env var name.
// Returns the env var and a registry entry (if matched), or an error.
func resolveEnvVar(arg string) (string, *registry.Entry, error) {
	// 1. Check built-in shorthand first
	if entry := registry.Lookup(arg); entry != nil {
		return entry.EnvVar, entry, nil
	}

	// 2. Validate as env var name
	if !validEnvVarPattern.MatchString(arg) {
		return "", nil, fmt.Errorf("invalid environment variable name %q: use only letters, numbers, and underscores (cannot start with a number)", arg)
	}

	entry := registry.LookupByEnvVar(arg)
	return arg, entry, nil
}

// resolveKey finds an existing key entry from a CLI argument.
// Tries: env var match → registry shorthand → legacy name match.
func resolveKey(cfg *config.Config, arg string) (*config.KeyEntry, error) {
	// 1. Direct env var match
	if entry := cfg.FindKeyByEnvVar(arg); entry != nil {
		return entry, nil
	}

	// 2. Registry shorthand → expand to env var
	if regEntry := registry.Lookup(arg); regEntry != nil {
		if entry := cfg.FindKeyByEnvVar(regEntry.EnvVar); entry != nil {
			return entry, nil
		}
	}

	// 3. Legacy name match (backward compat for v1 config entries)
	if entry := cfg.FindKey(arg); entry != nil {
		return entry, nil
	}

	return nil, fmt.Errorf("key %q is not registered", arg)
}
