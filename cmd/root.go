package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/eazyhozy/sekret/internal/keychain"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

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

// readInput reads a line of visible text from the user.
// Returns empty string if user presses Enter without typing.
// Override with SetReadInput() for testing.
var readInput = func(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		return "", fmt.Errorf("failed to read input: EOF")
	}
	return scanner.Text(), nil
}

// SetReadInput overrides the input reader (for testing).
func SetReadInput(fn func(string) (string, error)) {
	readInput = fn
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

var rootCmd = &cobra.Command{
	Use:   "sekret",
	Short: "Secure your API keys in OS keychain, load them as env vars",
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
