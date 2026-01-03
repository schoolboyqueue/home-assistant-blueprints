package validator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"

	errs "github.com/home-assistant-blueprints/validate-blueprint-go/internal/errors"
)

// ReportResults prints validation results and returns success status
func (v *BlueprintValidator) ReportResults() bool {
	fmt.Println()

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	// Use categorized output if we have categorized errors/warnings and grouping is enabled
	if v.GroupByCategory && (len(v.CategorizedErrors) > 0 || len(v.CategorizedWarnings) > 0) {
		output := FormatCategorySummary(v.CategorizedErrors, v.CategorizedWarnings, true)
		fmt.Print(output)
	} else {
		// Fall back to legacy flat output
		v.reportLegacyResults()
	}

	if len(v.Errors) == 0 && len(v.Warnings) == 0 {
		fmt.Printf("%s Blueprint is valid!\n", green("OK"))
		return true
	}

	if len(v.Errors) == 0 {
		fmt.Printf("%s Blueprint is valid (with %d warnings)\n", green("OK"), len(v.Warnings))
		return true
	}

	// Print category summary if we have categorized errors
	if v.GroupByCategory && len(v.CategorizedErrors) > 0 {
		v.printCategorySummary()
	}

	fmt.Printf("%s Blueprint validation failed with %d errors\n", red("FAIL"), len(v.Errors))
	return false
}

// reportLegacyResults prints errors and warnings in the legacy flat format
func (v *BlueprintValidator) reportLegacyResults() {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

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
}

// printCategorySummary prints a brief summary of errors by category
func (v *BlueprintValidator) printCategorySummary() {
	counts := v.ErrorCounts()
	if len(counts) == 0 {
		return
	}

	cyan := color.New(color.FgCyan).SprintFunc()

	fmt.Print("Error summary: ")
	first := true
	for _, cat := range AllCategories() {
		if count, ok := counts[cat]; ok && count > 0 {
			if !first {
				fmt.Print(", ")
			}
			fmt.Printf("%s: %d", cyan(cat.String()), count)
			first = false
		}
	}
	fmt.Println()
}

// ReportResultsFlat prints validation results without category grouping
func (v *BlueprintValidator) ReportResultsFlat() bool {
	fmt.Println()

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	v.reportLegacyResults()

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

// ReportResultsByCategory prints validation results grouped by category
func (v *BlueprintValidator) ReportResultsByCategory() bool {
	// Temporarily enable category grouping
	oldGroupByCategory := v.GroupByCategory
	v.GroupByCategory = true
	defer func() { v.GroupByCategory = oldGroupByCategory }()

	return v.ReportResults()
}

// ReportFilteredResults prints validation results for specific categories only
func (v *BlueprintValidator) ReportFilteredResults(categories ...ErrorCategory) bool {
	fmt.Println()

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	filteredErrors := v.GetErrorsByCategory(categories...)
	filteredWarnings := v.GetWarningsByCategory(categories...)

	if len(filteredErrors) > 0 || len(filteredWarnings) > 0 {
		output := FormatCategorySummary(filteredErrors, filteredWarnings, true)
		fmt.Print(output)
	}

	if len(filteredErrors) == 0 && len(filteredWarnings) == 0 {
		fmt.Printf("%s No issues found in selected categories!\n", green("OK"))
		return true
	}

	if len(filteredErrors) == 0 {
		fmt.Printf("%s No errors in selected categories (with %d warnings)\n", green("OK"), len(filteredWarnings))
		return true
	}

	fmt.Printf("%s Found %d errors in selected categories\n", red("FAIL"), len(filteredErrors))
	return false
}

// CheckReadmeExists checks if README.md exists in the blueprint directory
func (v *BlueprintValidator) CheckReadmeExists() {
	dir := filepath.Dir(v.FilePath)
	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		v.AddTypedWarning(errs.ErrMissingReadme().WithMessagef("No README.md found in %s/ directory", filepath.Base(dir)))
	}
}

// CheckChangelogExists checks if CHANGELOG.md exists in the blueprint directory
func (v *BlueprintValidator) CheckChangelogExists() {
	dir := filepath.Dir(v.FilePath)
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		v.AddTypedWarning(errs.ErrMissingChangelog().WithMessagef("No CHANGELOG.md found in %s/ directory", filepath.Base(dir)))
	}
}
