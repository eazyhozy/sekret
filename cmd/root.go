package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sekret",
	Short: "Secure your API keys in OS keychain, load them as env vars",
	Long: `sekret stores your API keys in the OS keychain and injects them
as environment variables. No more plaintext secrets in .zshrc.

Add 'eval "$(sekret env)"' to your .zshrc to automatically load
all registered keys when opening a new terminal.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
