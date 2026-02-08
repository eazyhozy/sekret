package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Output all keys as export statements",
	Long: `Output all registered keys as shell export statements.

Add this to your .zshrc:
  eval "$(sekret env)"`,
	Args: cobra.NoArgs,
	RunE: runEnv,
}

func init() {
	rootCmd.AddCommand(envCmd)
}

// shellEscape escapes a value for safe use in a shell double-quoted string.
func shellEscape(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		`$`, `\$`,
		"`", "\\`",
	)
	return replacer.Replace(s)
}

func runEnv(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Keys) == 0 {
		return nil
	}

	for _, k := range cfg.Keys {
		val, err := store.Get(k.KeychainKey())
		if err != nil {
			fmt.Fprintf(os.Stderr, "sekret: warning: could not read key %q: %v\n", k.EnvVar, err)
			continue
		}
		fmt.Printf("export %s=\"%s\"\n", k.EnvVar, shellEscape(val))
	}

	return nil
}
