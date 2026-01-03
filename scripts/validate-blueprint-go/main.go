// validate-blueprint - A comprehensive Home Assistant Blueprint validator
//
// Usage:
//
//	validate-blueprint <blueprint.yaml>
//	validate-blueprint --all
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli/v3"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/shutdown"
	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/validator"
)

// Version information - set by ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Create shutdown coordinator with signal handling
	coord, ctx := shutdown.New(
		shutdown.WithGracePeriod(5*time.Second),
		shutdown.WithOnShutdown(func(reason string) {
			fmt.Fprintf(os.Stderr, "\n%s\n", reason)
		}),
	)
	coord.HandleSignals()

	cmd := &cli.Command{
		Name:      "validate-blueprint",
		Usage:     "A comprehensive Home Assistant Blueprint validator",
		Version:   Version,
		ArgsUsage: "[blueprint.yaml]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Validate all blueprints in the repository",
			},
		},
		Action: runValidation,
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		// Check if it was an interrupt
		if coord.IsShuttingDown() {
			os.Exit(130) // Standard exit code for SIGINT
		}
		os.Exit(1)
	}
}

// runValidation is the main action for the CLI command
func runValidation(ctx context.Context, cmd *cli.Command) error {
	validateAll := cmd.Bool("all")
	args := cmd.Args()

	// Handle no arguments (show help with error exit)
	if args.Len() == 0 && !validateAll {
		_ = cli.ShowAppHelp(cmd) //nolint:errcheck // error not relevant, we exit immediately
		return cli.Exit("", 1)
	}

	// Execute the appropriate command
	var success bool
	switch {
	case validateAll:
		success = runValidateAllWithContext(ctx)
	case args.Len() > 0:
		success = validateSingle(args.First())
	default:
		_ = cli.ShowAppHelp(cmd) //nolint:errcheck // error not relevant, we exit immediately
		return cli.Exit("", 1)
	}

	if !success {
		return cli.Exit("", 1)
	}
	return nil
}

// findAllBlueprints finds all blueprint YAML files in the repository
func findAllBlueprints(basePath string) ([]string, error) {
	var blueprints []string
	patterns := []string{"*_pro.yaml", "*_pro_blueprint.yaml", "blueprint.yaml"}
	excludeDirs := map[string]bool{
		".git": true, "node_modules": true, "venv": true, ".venv": true, "__pycache__": true,
	}

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr // Return the actual error instead of nil
		}

		// Skip excluded directories
		if info.IsDir() && excludeDirs[info.Name()] {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches any pattern
		for _, pattern := range patterns {
			matched, matchErr := filepath.Match(pattern, info.Name())
			if matchErr != nil {
				continue // Skip invalid patterns
			}
			if matched {
				blueprints = append(blueprints, path)
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(blueprints)
	return blueprints, nil
}

// validateSingle validates a single blueprint file
func validateSingle(blueprintPath string) bool {
	v := validator.New(blueprintPath)
	return v.Validate()
}

// runValidateAllWithContext validates all blueprints with context support for interruption.
//
//nolint:gocyclo // Complexity is acceptable for main orchestration function
func runValidateAllWithContext(ctx context.Context) bool {
	// Navigate up from scripts/validate-blueprint-go/ to the repo root
	execPath, err := os.Executable()
	if err != nil {
		// Fall back to current directory
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			execPath = "."
		} else {
			execPath = cwd
		}
	}

	// Try to find repo root by looking for common markers
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(execPath)))

	// If running from source, use relative path
	if _, statErr := os.Stat(filepath.Join(repoRoot, "blueprints")); os.IsNotExist(statErr) {
		// Try current working directory
		cwd, cwdErr := os.Getwd()
		if cwdErr == nil {
			if _, checkErr := os.Stat(filepath.Join(cwd, "blueprints")); checkErr == nil {
				repoRoot = cwd
			} else {
				// Go up directories looking for blueprints folder
				for range 5 {
					parent := filepath.Dir(repoRoot)
					if parent == repoRoot {
						break
					}
					repoRoot = parent
					if _, lookupErr := os.Stat(filepath.Join(repoRoot, "blueprints")); lookupErr == nil {
						break
					}
				}
			}
		}
	}

	blueprints, err := findAllBlueprints(repoRoot)
	if err != nil {
		fmt.Printf("Error finding blueprints: %v\n", err)
		return false
	}

	if len(blueprints) == 0 {
		fmt.Println("No blueprints found in repository")
		return false
	}

	fmt.Printf("Found %d blueprint(s) to validate (Ctrl+C to interrupt)\n\n", len(blueprints))

	// Track partial results for graceful shutdown reporting
	partialResult := shutdown.NewPartialResult(len(blueprints))

	type result struct {
		path    string
		success bool
	}
	var results []result
	interrupted := false

	for _, bp := range blueprints {
		// Check for context cancellation before each validation
		select {
		case <-ctx.Done():
			interrupted = true
			completed, total, _, _, _ := partialResult.Summary()
			fmt.Printf("\nValidation interrupted after %d/%d blueprints\n", completed, total)
			goto summary
		default:
		}

		v := validator.New(bp)
		success := v.Validate()
		results = append(results, result{path: bp, success: success})

		// Track partial results
		if success {
			partialResult.RecordPass(bp)
		} else {
			partialResult.RecordFail(bp, "validation failed")
		}

		fmt.Println(strings.Repeat("-", 80))
		fmt.Println()
	}

summary:
	// Summary
	fmt.Println(strings.Repeat("=", 80))
	if interrupted {
		fmt.Println("PARTIAL SUMMARY (interrupted)")
	} else {
		fmt.Println("SUMMARY")
	}
	fmt.Println(strings.Repeat("=", 80))

	passed := 0
	for _, r := range results {
		if r.success {
			passed++
		}
	}
	failed := len(results) - passed

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	for _, r := range results {
		relPath, relErr := filepath.Rel(repoRoot, r.path)
		if relErr != nil || relPath == "" {
			relPath = r.path
		}
		if r.success {
			fmt.Printf("%s %s\n", green("OK"), relPath)
		} else {
			fmt.Printf("%s %s\n", red("X"), relPath)
		}
	}

	fmt.Println()
	if interrupted {
		skipped := len(blueprints) - len(results)
		fmt.Printf("Completed: %d | Passed: %d | Failed: %d | %s: %d\n",
			len(results), passed, failed, yellow("Skipped"), skipped)
	} else {
		fmt.Printf("Total: %d | Passed: %d | Failed: %d\n", len(results), passed, failed)
	}

	// Return false if any failures or if interrupted
	return failed == 0 && !interrupted
}
