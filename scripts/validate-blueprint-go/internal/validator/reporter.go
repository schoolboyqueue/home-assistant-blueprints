package validator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

// ReportResults prints validation results and returns success status
func (v *BlueprintValidator) ReportResults() bool {
	fmt.Println()

	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	if len(v.Errors) > 0 {
		fmt.Printf("%s ERRORS:\n", red("X"))
		for _, err := range v.Errors {
			fmt.Printf("  %s %s\n", red("*"), err)
		}
		fmt.Println()
	}

	if len(v.Warnings) > 0 {
		fmt.Printf("%s WARNINGS:\n", yellow("!"))
		for _, warning := range v.Warnings {
			fmt.Printf("  %s %s\n", yellow("*"), warning)
		}
		fmt.Println()
	}

	if len(v.Errors) == 0 && len(v.Warnings) == 0 {
		fmt.Printf("%s Blueprint is valid!\n", green("OK"))
		return true
	}

	if len(v.Errors) == 0 {
		fmt.Printf("%s Blueprint is valid (with %d warnings)\n", green("OK"), len(v.Warnings))
		return true
	}

	fmt.Printf("%s Blueprint validation failed with %d errors\n", red("FAIL"), len(v.Errors))
	return false
}

// CheckReadmeExists checks if README.md exists in the blueprint directory
func (v *BlueprintValidator) CheckReadmeExists() {
	dir := filepath.Dir(v.FilePath)
	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		v.AddWarningf("No README.md found in %s/ directory", filepath.Base(dir))
	}
}

// CheckChangelogExists checks if CHANGELOG.md exists in the blueprint directory
func (v *BlueprintValidator) CheckChangelogExists() {
	dir := filepath.Dir(v.FilePath)
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		v.AddWarningf("No CHANGELOG.md found in %s/ directory", filepath.Base(dir))
	}
}
