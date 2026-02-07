package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered keys",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func maskKey(value string) string {
	if len(value) <= 4 {
		return "****"
	}

	// Find the prefix portion (up to the first meaningful character after known prefix patterns)
	prefix := ""
	prefixes := []string{"sk-proj-", "sk-ant-", "github_pat_", "sk-", "ghp_", "gsk_", "AIza"}
	for _, p := range prefixes {
		if len(value) >= len(p) && value[:len(p)] == p {
			prefix = p
			break
		}
	}

	if prefix == "" && len(value) > 8 {
		prefix = value[:4]
	} else if prefix == "" {
		prefix = value[:2]
	}

	suffix := value[len(value)-4:]
	return prefix + "..." + suffix
}

func runList(_ *cobra.Command, _ []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Keys) == 0 {
		fmt.Fprintln(os.Stderr, "No keys registered. Use 'sekret add <name>' to get started.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	_, _ = fmt.Fprintln(w, "Name\tEnv Variable\tKey Preview\tAdded")
	_, _ = fmt.Fprintln(w, "----\t------------\t-----------\t-----")

	for _, k := range cfg.Keys {
		preview := "(unavailable)"
		if val, err := store.Get(k.Name); err == nil {
			preview = maskKey(val)
		}
		added := humanize.Time(k.AddedAt)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", k.Name, k.EnvVar, preview, added)
	}

	return w.Flush()
}
