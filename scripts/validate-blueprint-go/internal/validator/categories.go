package validator

import (
	"fmt"
	"slices"
	"strings"

	"github.com/fatih/color"

	errs "github.com/home-assistant-blueprints/validate-blueprint-go/internal/errors"
)

// ErrorCategory represents the type/category of a validation error
type ErrorCategory int

const (
	// CategorySyntax covers YAML syntax and parsing errors
	CategorySyntax ErrorCategory = iota
	// CategorySchema covers structural issues (missing keys, wrong types)
	CategorySchema
	// CategoryReferences covers undefined or unused input references
	CategoryReferences
	// CategoryTemplates covers Jinja2 template syntax errors
	CategoryTemplates
	// CategoryInputs covers input definition and selector issues
	CategoryInputs
	// CategoryTriggers covers trigger validation errors
	CategoryTriggers
	// CategoryConditions covers condition validation errors
	CategoryConditions
	// CategoryActions covers action/service call validation errors
	CategoryActions
	// CategoryDocumentation covers README/CHANGELOG and documentation issues
	CategoryDocumentation
)

// categoryNames maps categories to their display names
var categoryNames = map[ErrorCategory]string{
	CategorySyntax:        "Syntax",
	CategorySchema:        "Schema",
	CategoryReferences:    "References",
	CategoryTemplates:     "Templates",
	CategoryInputs:        "Inputs",
	CategoryTriggers:      "Triggers",
	CategoryConditions:    "Conditions",
	CategoryActions:       "Actions",
	CategoryDocumentation: "Documentation",
}

// categoryDescriptions provides additional context for each category
var categoryDescriptions = map[ErrorCategory]string{
	CategorySyntax:        "YAML syntax and parsing issues",
	CategorySchema:        "Blueprint structure and required fields",
	CategoryReferences:    "Input references (!input) and variable usage",
	CategoryTemplates:     "Jinja2 template syntax and usage",
	CategoryInputs:        "Input definitions and selectors",
	CategoryTriggers:      "Trigger definitions and configuration",
	CategoryConditions:    "Condition definitions and configuration",
	CategoryActions:       "Action and service call configuration",
	CategoryDocumentation: "README, CHANGELOG, and documentation",
}

// String returns the display name for the error category
func (c ErrorCategory) String() string {
	if name, ok := categoryNames[c]; ok {
		return name
	}
	return "Unknown"
}

// Description returns a description of what the category covers
func (c ErrorCategory) Description() string {
	if desc, ok := categoryDescriptions[c]; ok {
		return desc
	}
	return "Unknown category"
}

// AllCategories returns all defined error categories in display order
func AllCategories() []ErrorCategory {
	return []ErrorCategory{
		CategorySyntax,
		CategorySchema,
		CategoryInputs,
		CategoryReferences,
		CategoryTemplates,
		CategoryTriggers,
		CategoryConditions,
		CategoryActions,
		CategoryDocumentation,
	}
}

// CategorizedError represents a validation error with its category
type CategorizedError struct {
	Category ErrorCategory
	Path     string
	Message  string
}

// String returns a formatted string representation of the error
func (e CategorizedError) String() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s", e.Path, e.Message)
	}
	return e.Message
}

// FullString returns a string with category prefix
func (e CategorizedError) FullString() string {
	if e.Path != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Category.String(), e.Path, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Category.String(), e.Message)
}

// CategorizedWarning represents a validation warning with its category
type CategorizedWarning struct {
	Category ErrorCategory
	Path     string
	Message  string
}

// String returns a formatted string representation of the warning
func (w CategorizedWarning) String() string {
	if w.Path != "" {
		return fmt.Sprintf("%s: %s", w.Path, w.Message)
	}
	return w.Message
}

// FullString returns a string with category prefix
func (w CategorizedWarning) FullString() string {
	if w.Path != "" {
		return fmt.Sprintf("[%s] %s: %s", w.Category.String(), w.Path, w.Message)
	}
	return fmt.Sprintf("[%s] %s", w.Category.String(), w.Message)
}

// ErrorsByCategory groups errors by their category
func ErrorsByCategory(errors []CategorizedError) map[ErrorCategory][]CategorizedError {
	grouped := make(map[ErrorCategory][]CategorizedError)
	for _, err := range errors {
		grouped[err.Category] = append(grouped[err.Category], err)
	}
	return grouped
}

// WarningsByCategory groups warnings by their category
func WarningsByCategory(warnings []CategorizedWarning) map[ErrorCategory][]CategorizedWarning {
	grouped := make(map[ErrorCategory][]CategorizedWarning)
	for _, warn := range warnings {
		grouped[warn.Category] = append(grouped[warn.Category], warn)
	}
	return grouped
}

// FilterErrorsByCategory returns errors matching the specified categories
func FilterErrorsByCategory(errors []CategorizedError, categories ...ErrorCategory) []CategorizedError {
	if len(categories) == 0 {
		return errors
	}
	categorySet := make(map[ErrorCategory]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}
	var filtered []CategorizedError
	for _, err := range errors {
		if categorySet[err.Category] {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// FilterWarningsByCategory returns warnings matching the specified categories
func FilterWarningsByCategory(warnings []CategorizedWarning, categories ...ErrorCategory) []CategorizedWarning {
	if len(categories) == 0 {
		return warnings
	}
	categorySet := make(map[ErrorCategory]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}
	var filtered []CategorizedWarning
	for _, warn := range warnings {
		if categorySet[warn.Category] {
			filtered = append(filtered, warn)
		}
	}
	return filtered
}

// ExcludeErrorsByCategory returns errors NOT matching the specified categories
func ExcludeErrorsByCategory(errors []CategorizedError, categories ...ErrorCategory) []CategorizedError {
	if len(categories) == 0 {
		return errors
	}
	categorySet := make(map[ErrorCategory]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}
	var filtered []CategorizedError
	for _, err := range errors {
		if !categorySet[err.Category] {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// ExcludeWarningsByCategory returns warnings NOT matching the specified categories
func ExcludeWarningsByCategory(warnings []CategorizedWarning, categories ...ErrorCategory) []CategorizedWarning {
	if len(categories) == 0 {
		return warnings
	}
	categorySet := make(map[ErrorCategory]bool)
	for _, cat := range categories {
		categorySet[cat] = true
	}
	var filtered []CategorizedWarning
	for _, warn := range warnings {
		if !categorySet[warn.Category] {
			filtered = append(filtered, warn)
		}
	}
	return filtered
}

// CountByCategory returns a map of category to error count
func CountByCategory(errors []CategorizedError) map[ErrorCategory]int {
	counts := make(map[ErrorCategory]int)
	for _, err := range errors {
		counts[err.Category]++
	}
	return counts
}

// CountWarningsByCategory returns a map of category to warning count
func CountWarningsByCategory(warnings []CategorizedWarning) map[ErrorCategory]int {
	counts := make(map[ErrorCategory]int)
	for _, warn := range warnings {
		counts[warn.Category]++
	}
	return counts
}

// FormatCategorySummary returns a formatted summary of errors/warnings by category
func FormatCategorySummary(errors []CategorizedError, warnings []CategorizedWarning, showCategories bool) string {
	var sb strings.Builder

	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	if showCategories && len(errors) > 0 {
		// Group errors by category
		grouped := ErrorsByCategory(errors)

		// Get categories in display order
		var presentCategories []ErrorCategory
		for _, cat := range AllCategories() {
			if catErrs, ok := grouped[cat]; ok && len(catErrs) > 0 {
				presentCategories = append(presentCategories, cat)
			}
		}

		for _, cat := range presentCategories {
			catErrs := grouped[cat]
			fmt.Fprintf(&sb, "%s %s ERRORS (%d):\n", red("X"), cyan(cat.String()), len(catErrs))
			for _, err := range catErrs {
				fmt.Fprintf(&sb, "  %s %s\n", red("*"), err.String())
			}
			sb.WriteString("\n")
		}
	} else if len(errors) > 0 {
		// Flat error display
		fmt.Fprintf(&sb, "%s ERRORS:\n", red("X"))
		for _, err := range errors {
			fmt.Fprintf(&sb, "  %s %s\n", red("*"), err.String())
		}
		sb.WriteString("\n")
	}

	if showCategories && len(warnings) > 0 {
		// Group warnings by category
		grouped := WarningsByCategory(warnings)

		// Get categories in display order
		var presentCategories []ErrorCategory
		for _, cat := range AllCategories() {
			if warns, ok := grouped[cat]; ok && len(warns) > 0 {
				presentCategories = append(presentCategories, cat)
			}
		}

		for _, cat := range presentCategories {
			warns := grouped[cat]
			fmt.Fprintf(&sb, "%s %s WARNINGS (%d):\n", yellow("!"), cyan(cat.String()), len(warns))
			for _, warn := range warns {
				fmt.Fprintf(&sb, "  %s %s\n", yellow("*"), warn.String())
			}
			sb.WriteString("\n")
		}
	} else if len(warnings) > 0 {
		// Flat warning display
		fmt.Fprintf(&sb, "%s WARNINGS:\n", yellow("!"))
		for _, warn := range warnings {
			fmt.Fprintf(&sb, "  %s %s\n", yellow("*"), warn.String())
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// SortedCategoryKeys returns category keys sorted by their enum order
func SortedCategoryKeys(m map[ErrorCategory][]CategorizedError) []ErrorCategory {
	var keys []ErrorCategory
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}

// errorTypeToCategoryMap maps errs.ErrorType to ErrorCategory
var errorTypeToCategoryMap = map[errs.ErrorType]ErrorCategory{
	errs.ErrorTypeSyntax:        CategorySyntax,
	errs.ErrorTypeSchema:        CategorySchema,
	errs.ErrorTypeReference:     CategoryReferences,
	errs.ErrorTypeTemplate:      CategoryTemplates,
	errs.ErrorTypeInput:         CategoryInputs,
	errs.ErrorTypeTrigger:       CategoryTriggers,
	errs.ErrorTypeCondition:     CategoryConditions,
	errs.ErrorTypeAction:        CategoryActions,
	errs.ErrorTypeDocumentation: CategoryDocumentation,
	// Parsing and Validation errors map to Syntax
	errs.ErrorTypeParsing:    CategorySyntax,
	errs.ErrorTypeValidation: CategorySchema,
}

// CategoryFromErrorType converts an errs.ErrorType to an ErrorCategory.
// Returns CategorySchema for unknown error types as a sensible default.
func CategoryFromErrorType(t errs.ErrorType) ErrorCategory {
	if cat, ok := errorTypeToCategoryMap[t]; ok {
		return cat
	}
	return CategorySchema
}

// CategorizedErrorFromError creates a CategorizedError from an errs.Error.
func CategorizedErrorFromError(e *errs.Error) CategorizedError {
	return CategorizedError{
		Category: CategoryFromErrorType(e.Type),
		Path:     e.Path,
		Message:  e.Message,
	}
}

// CategorizedWarningFromError creates a CategorizedWarning from an errs.Error.
func CategorizedWarningFromError(e *errs.Error) CategorizedWarning {
	return CategorizedWarning{
		Category: CategoryFromErrorType(e.Type),
		Path:     e.Path,
		Message:  e.Message,
	}
}
