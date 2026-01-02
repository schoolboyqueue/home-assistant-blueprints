// validate-blueprint - A comprehensive Home Assistant Blueprint validator
//
// This tool performs comprehensive validation of Home Assistant blueprint files:
// 1. YAML syntax validation
// 2. Blueprint schema validation (required keys, structure)
// 3. Input/selector validation
// 4. Template syntax checking
// 5. Service call structure validation
// 6. Version sync validation
// 7. Trigger validation
// 8. Condition validation
// 9. Mode validation
// 10. Input reference validation
// And more...
//
// Usage:
//
//	validate-blueprint <blueprint.yaml>
//	validate-blueprint --all
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/validator"
)

// Version information - set by ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// showHelp displays comprehensive usage information
func showHelp() {
	fmt.Print(`Usage: validate-blueprint <blueprint.yaml>
       validate-blueprint --all
       validate-blueprint --version
       validate-blueprint --help

validate-blueprint - A comprehensive Home Assistant Blueprint validator

Description:
  This tool performs comprehensive validation of Home Assistant blueprint files,
  checking for common errors and best practices.

Commands:
  <blueprint.yaml>     Validate a single blueprint file
  --all                Validate all blueprints in the repository
  --version, -v        Show version information
  --help, -h, help     Show this help message

Validation Checks:
  1.  YAML syntax validation
  2.  Blueprint schema validation (required keys, structure)
  3.  Input/selector validation
  4.  Template syntax checking
  5.  Service call structure validation
  6.  Version sync validation
  7.  Trigger validation
  8.  Condition validation
  9.  Mode validation
  10. Input reference validation
  11. Hysteresis boundary validation
  12. Variable definition validation
  13. README.md and CHANGELOG.md existence check

Valid Selectors:
  action, addon, area, attribute, boolean, color_rgb, color_temp, condition,
  conversation_agent, country, date, datetime, device, duration, entity, file,
  floor, icon, label, language, location, media, navigation, number, object,
  select, state, target, template, text, theme, time, trigger, ui_action, ui_color

Valid Modes:
  single, restart, queued, parallel

Valid Condition Types:
  and, or, not, state, numeric_state, template, time, zone, trigger, sun, device

Blueprint File Patterns (for --all):
  The tool searches for files matching these patterns:
  - *_pro.yaml
  - *_pro_blueprint.yaml
  - blueprint.yaml

Exit Codes:
  0    Validation passed (may have warnings)
  1    Validation failed (has errors) or invalid usage

Examples:
  validate-blueprint my_automation.yaml
  validate-blueprint blueprints/automation/motion_light.yaml
  validate-blueprint --all

`)
}

// isHelpRequested checks if any argument is a help flag
func isHelpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

// isVersionRequested checks if any argument is a version flag
func isVersionRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--version" || arg == "-v" || arg == "version" {
			return true
		}
	}
	return false
}

// showVersion displays version information
func showVersion() {
	fmt.Printf("validate-blueprint %s\n", Version)
	fmt.Printf("  Build time: %s\n", BuildTime)
	fmt.Printf("  Git commit: %s\n", GitCommit)
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

// validateAll validates all blueprints in the repository
func validateAll() bool {
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

	fmt.Printf("Found %d blueprint(s) to validate\n\n", len(blueprints))

	type result struct {
		path    string
		success bool
	}
	var results []result

	for _, bp := range blueprints {
		v := validator.New(bp)
		success := v.Validate()
		results = append(results, result{path: bp, success: success})
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("SUMMARY")
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
	fmt.Printf("Total: %d | Passed: %d | Failed: %d\n", len(results), passed, failed)

	return failed == 0
}

func main() {
	// Check for version first
	if isVersionRequested(os.Args[1:]) {
		showVersion()
		os.Exit(0)
	}

	// Check for help
	if len(os.Args) < 2 || isHelpRequested(os.Args[1:]) {
		showHelp()
		if len(os.Args) < 2 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	var success bool
	if os.Args[1] == "--all" {
		success = validateAll()
	} else {
		success = validateSingle(os.Args[1])
	}

	if success {
		os.Exit(0)
	}
	os.Exit(1)
}
