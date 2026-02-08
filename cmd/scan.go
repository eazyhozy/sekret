package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/eazyhozy/sekret/internal/config"
	"github.com/eazyhozy/sekret/internal/scanner"
	"github.com/spf13/cobra"
)

var scanPath string

// exitFunc is the function called to exit the process. Override for testing.
var exitFunc = os.Exit

// SetExitFunc overrides the exit function (for testing).
func SetExitFunc(fn func(int)) {
	exitFunc = fn
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Detect plaintext API keys in shell config files",
	Long: `Scan shell config files for plaintext API keys in export statements.

By default, scans ~/.zshrc, ~/.zshenv, ~/.zprofile, ~/.bashrc,
~/.bash_profile, and ~/.profile.

Use --path to scan a specific file or directory instead.`,
	Args: cobra.NoArgs,
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringVar(&scanPath, "path", "", "scan a specific file or directory")
	rootCmd.AddCommand(scanCmd)
}

// fileScanResult holds scan results for a single file.
type fileScanResult struct {
	path     string
	findings []scanner.Finding
	skipped  bool // true if file does not exist
}

func runScan(_ *cobra.Command, _ []string) error {
	paths, err := resolveScanTargets(scanPath)
	if err != nil {
		return err
	}

	// Scan each file individually to track per-file results
	var results []fileScanResult
	var allFindings []scanner.Finding

	for _, path := range paths {
		findings, err := scanner.ScanFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				results = append(results, fileScanResult{path: path, skipped: true})
				continue
			}
			return err
		}
		results = append(results, fileScanResult{path: path, findings: findings})
		allFindings = append(allFindings, findings...)
	}

	// Print per-file scan summary
	printScanSummary(results)

	if len(allFindings) == 0 {
		fmt.Println("\nNo plaintext keys found.")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Printf("\nFound %d potential plaintext %s:\n\n",
		len(allFindings), pluralize(len(allFindings), "key", "keys"))

	for _, f := range allFindings {
		annotation := annotate(cfg, f)
		displayPath := shortenHome(f.FilePath)

		line := fmt.Sprintf("  %s:%-8d export %s=\"%s\"",
			displayPath, f.Line, f.EnvVar, scanner.MaskValue(f.Value))

		if annotation != "" {
			line += "  (" + annotation + ")"
		}
		fmt.Println(line)
	}

	// Exit code 1 when keys are found
	exitFunc(1)
	return nil
}

// printScanSummary prints which files were scanned and how many keys each had.
func printScanSummary(results []fileScanResult) {
	scanned := 0
	for _, r := range results {
		if !r.skipped {
			scanned++
		}
	}

	fmt.Printf("Scanned %d %s:\n", scanned, pluralize(scanned, "file", "files"))
	for _, r := range results {
		if r.skipped {
			continue
		}
		displayPath := shortenHome(r.path)
		count := len(r.findings)
		if count == 0 {
			fmt.Printf("  %-28s clean\n", displayPath)
		} else {
			fmt.Printf("  %-28s %d %s found\n", displayPath, count, pluralize(count, "key", "keys"))
		}
	}
}

// resolveScanTargets determines which files to scan.
func resolveScanTargets(path string) ([]string, error) {
	if path == "" {
		targets := scanner.DefaultTargets()
		if targets == nil {
			return nil, fmt.Errorf("could not determine home directory")
		}
		return targets, nil
	}
	return scanner.ResolvePath(path)
}

// annotate returns a status annotation for a finding based on sekret state.
func annotate(cfg *config.Config, f scanner.Finding) string {
	entry := cfg.FindKeyByEnvVar(f.EnvVar)
	if entry == nil {
		return ""
	}

	keychainKey := entry.KeychainKey()
	stored, err := store.Get(keychainKey)
	if err != nil {
		return "already in sekret"
	}

	if stored == f.Value {
		return "already in sekret, safe to remove"
	}
	return "already in sekret, value differs!"
}

// shortenHome replaces the home directory prefix with ~.
func shortenHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

// pluralize returns singular or plural form based on count.
func pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
