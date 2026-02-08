package cmd

import (
	"fmt"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/scanner"
	"github.com/spf13/cobra"
)

var importFile string

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import plaintext API keys from shell config files into sekret",
	Long: `Parse export statements from shell config files and register them in sekret.

By default, scans ~/.zshrc, ~/.zshenv, ~/.zprofile, ~/.bashrc,
~/.bash_profile, and ~/.profile.

Use --file to import from a specific file instead.`,
	Args: cobra.NoArgs,
	RunE: runImport,
}

func init() {
	importCmd.Flags().StringVar(&importFile, "file", "", "import from a specific file")
	rootCmd.AddCommand(importCmd)
}

// importResult tracks the outcome of a single finding during import.
type importResult struct {
	finding scanner.Finding
	status  string // "imported", "skipped", "cancelled", "failed", "overwritten"
	err     error  // non-nil only for "failed"
}

func runImport(_ *cobra.Command, _ []string) error {
	stderr := rootCmd.ErrOrStderr()

	paths, err := resolveImportTargets(importFile)
	if err != nil {
		return err
	}

	findings, err := scanner.ScanFiles(paths)
	if err != nil {
		return err
	}

	if len(findings) == 0 {
		_, _ = fmt.Fprintln(stderr, "No exportable keys found.")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(stderr, "\nFound %d exportable %s:\n\n",
		len(findings), pluralize(len(findings), "key", "keys"))
	_, _ = fmt.Fprintln(stderr, "Import each key? (y: import / s: skip / q: quit)")

	results := make([]importResult, 0, len(findings))

	for i, f := range findings {
		result, err := processImportFinding(stderr, cfg, f, i, len(findings))
		if err != nil {
			return err
		}

		results = append(results, result)

		if result.status == "cancelled" {
			// Mark remaining findings as cancelled too
			for _, remaining := range findings[i+1:] {
				results = append(results, importResult{finding: remaining, status: "cancelled"})
			}
			break
		}
	}

	printImportSummary(results)
	return nil
}

// processImportFinding handles a single finding in the import loop.
// Returns a fatal error only for config save failures.
func processImportFinding(stderr interface{ Write([]byte) (int, error) }, cfg *config.Config, f scanner.Finding, index, total int) (importResult, error) {
	displayPath := shortenHome(f.FilePath)

	_, _ = fmt.Fprintf(stderr, "\n  [%d/%d] %s (%s)\n", index+1, total, f.EnvVar, displayMaskedValue(f.Value))
	_, _ = fmt.Fprintf(stderr, "         %s:%d\n", displayPath, f.Line)

	// Check if already registered in sekret
	existing := cfg.FindKeyByEnvVar(f.EnvVar)
	if existing != nil {
		_, _ = fmt.Fprintf(stderr, "         Already registered in sekret.\n")
		return handleOverwrite(stderr, cfg, f, existing)
	}

	return handleNewKey(stderr, cfg, f)
}

// handleOverwrite prompts for overwrite of an existing key.
func handleOverwrite(stderr interface{ Write([]byte) (int, error) }, cfg *config.Config, f scanner.Finding, existing *config.KeyEntry) (importResult, error) {
	choice, err := readChoice("         Overwrite? [y/N]: ")
	if err != nil {
		return importResult{}, fmt.Errorf("failed to read input: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(choice)) {
	case "y", "yes":
		keychainKey := existing.KeychainKey()
		if err := store.Set(keychainKey, f.Value); err != nil {
			_, _ = fmt.Fprintf(stderr, "         Failed — %s\n", err)
			return importResult{finding: f, status: "failed", err: err}, nil
		}
		_, _ = fmt.Fprintf(stderr, "         Overwritten %s\n", f.EnvVar)
		return importResult{finding: f, status: "overwritten"}, nil
	case "", "n", "no":
		_, _ = fmt.Fprintf(stderr, "         Skipped\n")
		return importResult{finding: f, status: "skipped"}, nil
	default:
		_, _ = fmt.Fprintf(stderr, "         Invalid choice. Use y or n.\n")
		return handleOverwrite(stderr, cfg, f, existing)
	}
}

// handleNewKey prompts for import of a new key.
func handleNewKey(stderr interface{ Write([]byte) (int, error) }, cfg *config.Config, f scanner.Finding) (importResult, error) {
	choice, err := readChoice("         Import? [Y/s/q]: ")
	if err != nil {
		return importResult{}, fmt.Errorf("failed to read input: %w", err)
	}

	switch strings.ToLower(strings.TrimSpace(choice)) {
	case "", "y", "yes":
		return doImport(stderr, cfg, f)
	case "s", "skip":
		_, _ = fmt.Fprintf(stderr, "         Skipped\n")
		return importResult{finding: f, status: "skipped"}, nil
	case "q", "quit":
		_, _ = fmt.Fprintf(stderr, "         Cancelled\n")
		return importResult{finding: f, status: "cancelled"}, nil
	default:
		// Invalid input — re-prompt
		_, _ = fmt.Fprintf(stderr, "         Invalid choice. Use y, s, or q.\n")
		return handleNewKey(stderr, cfg, f)
	}
}

// doImport saves a finding to keychain and config.
func doImport(stderr interface{ Write([]byte) (int, error) }, cfg *config.Config, f scanner.Finding) (importResult, error) {
	if err := store.Set(f.EnvVar, f.Value); err != nil {
		_, _ = fmt.Fprintf(stderr, "         Failed — %s\n", err)
		return importResult{finding: f, status: "failed", err: err}, nil
	}

	if err := cfg.AddKey("", f.EnvVar); err != nil {
		return importResult{}, fmt.Errorf("failed to register key: %w", err)
	}
	if err := config.Save(cfg); err != nil {
		return importResult{}, fmt.Errorf("failed to save config: %w", err)
	}

	_, _ = fmt.Fprintf(stderr, "         Imported %s\n", f.EnvVar)
	return importResult{finding: f, status: "imported"}, nil
}

// displayMaskedValue returns a display string for a value, handling empty values.
func displayMaskedValue(value string) string {
	if value == "" {
		return "empty value"
	}
	return scanner.MaskValue(value)
}

// printImportSummary outputs the final result summary to stdout.
func printImportSummary(results []importResult) {
	imported := filterResults(results, "imported")
	overwritten := filterResults(results, "overwritten")
	skipped := filterResults(results, "skipped")
	cancelled := filterResults(results, "cancelled")
	failed := filterResults(results, "failed")

	// Summary line
	parts := []string{}
	if n := len(imported) + len(overwritten); n > 0 {
		parts = append(parts, fmt.Sprintf("%d imported", n))
	}
	if len(skipped) > 0 {
		parts = append(parts, fmt.Sprintf("%d skipped", len(skipped)))
	}
	if len(cancelled) > 0 {
		parts = append(parts, fmt.Sprintf("%d cancelled", len(cancelled)))
	}
	if len(failed) > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", len(failed)))
	}

	fmt.Printf("\nDone. %s.\n", strings.Join(parts, ", "))

	// Detail sections
	if len(imported) > 0 {
		fmt.Println("\n  Imported:")
		for _, r := range imported {
			fmt.Printf("    %-12s %s\n", shortenHome(r.finding.FilePath)+":"+fmt.Sprint(r.finding.Line), r.finding.EnvVar)
		}
	}

	if len(overwritten) > 0 {
		fmt.Println("\n  Overwritten:")
		for _, r := range overwritten {
			fmt.Printf("    %-12s %s\n", shortenHome(r.finding.FilePath)+":"+fmt.Sprint(r.finding.Line), r.finding.EnvVar)
		}
	}

	if len(skipped) > 0 {
		fmt.Println("\n  Skipped:")
		for _, r := range skipped {
			fmt.Printf("    %-12s %s\n", shortenHome(r.finding.FilePath)+":"+fmt.Sprint(r.finding.Line), r.finding.EnvVar)
		}
	}

	if len(cancelled) > 0 {
		fmt.Println("\n  Cancelled:")
		for _, r := range cancelled {
			fmt.Printf("    %-12s %s\n", shortenHome(r.finding.FilePath)+":"+fmt.Sprint(r.finding.Line), r.finding.EnvVar)
		}
	}

	if len(failed) > 0 {
		fmt.Println("\n  Failed:")
		for _, r := range failed {
			fmt.Printf("    %-12s %s — %s\n",
				shortenHome(r.finding.FilePath)+":"+fmt.Sprint(r.finding.Line),
				r.finding.EnvVar, r.err)
		}
	}

	// Advice to remove imported keys
	if len(imported)+len(overwritten) > 0 {
		fmt.Println("\nRemove the imported keys from your shell config (lines listed above).")
	}
}

// filterResults returns results with the given status.
func filterResults(results []importResult, status string) []importResult {
	var filtered []importResult
	for _, r := range results {
		if r.status == status {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// resolveImportTargets determines which files to scan for import.
func resolveImportTargets(file string) ([]string, error) {
	if file != "" {
		return scanner.ResolvePath(file)
	}
	targets := scanner.DefaultTargets()
	if targets == nil {
		return nil, fmt.Errorf("could not determine home directory")
	}
	return targets, nil
}
