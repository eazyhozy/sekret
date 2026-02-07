package cmd

import (
	"github.com/eazyhozy/sekret/internal/keychain"
	"github.com/spf13/cobra"
)

// store is the keychain store used by all commands.
// Override with SetStore() for testing.
var store keychain.Store = keychain.NewOSStore()

// SetStore overrides the keychain store (for testing).
func SetStore(s keychain.Store) {
	store = s
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
