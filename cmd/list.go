package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dustin/go-humanize"
	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/scanner"
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
	_, _ = fmt.Fprintln(w, "Env Variable\tKey Preview\tAdded")
	_, _ = fmt.Fprintln(w, "------------\t-----------\t-----")

	for _, k := range cfg.Keys {
		preview := "(unavailable)"
		if val, err := store.Get(k.KeychainKey()); err == nil {
			preview = scanner.MaskValue(val)
		}
		added := humanize.Time(k.AddedAt)
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", k.EnvVar, preview, added)
	}

	return w.Flush()
}
