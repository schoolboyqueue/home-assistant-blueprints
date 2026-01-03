package validator

import "fmt"

// BlueprintValidator validates Home Assistant Blueprint YAML files
type BlueprintValidator struct {
	FilePath            string
	Errors              []string
	Warnings            []string
	CategorizedErrors   []CategorizedError
	CategorizedWarnings []CategorizedWarning
	Data                map[string]interface{}
	DefinedInputs       map[string]bool
	UsedInputs          map[string]bool
	InputDefaults       map[string]interface{}
	InputSelectors      map[string]map[string]interface{}
	EntityInputs        map[string]bool
	InputDatetimeInputs map[string]bool
	DefinedVariables    map[string]bool
	JoinVariables       map[string]bool
	NonzeroDefaultVars  map[string]bool
	// GroupByCategory controls whether errors/warnings are grouped by category in output
	GroupByCategory bool
}

// New creates a new validator instance
func New(filePath string) *BlueprintValidator {
	return &BlueprintValidator{
		FilePath:            filePath,
		Errors:              []string{},
		Warnings:            []string{},
		CategorizedErrors:   []CategorizedError{},
		CategorizedWarnings: []CategorizedWarning{},
		Data:                make(map[string]interface{}),
		DefinedInputs:       make(map[string]bool),
		UsedInputs:          make(map[string]bool),
		InputDefaults:       make(map[string]interface{}),
		InputSelectors:      make(map[string]map[string]interface{}),
		EntityInputs:        make(map[string]bool),
		InputDatetimeInputs: make(map[string]bool),
		DefinedVariables:    make(map[string]bool),
		JoinVariables:       make(map[string]bool),
		NonzeroDefaultVars:  make(map[string]bool),
		GroupByCategory:     true, // Enable categorized output by default
	}
}

// Validate runs all validation checks
func (v *BlueprintValidator) Validate() bool {
	fmt.Printf("Validating: %s\n", v.FilePath)

	if !v.LoadYAML() {
		return false
	}

	v.ValidateStructure()
	v.ValidateBlueprintSection()
	v.ValidateMode()
	v.ValidateInputs()
	v.ValidateHysteresisBoundaries()
	v.ValidateVariables()
	v.ValidateVersionSync()
	v.ValidateTriggers()
	v.ValidateConditions()
	v.ValidateActions()
	v.ValidateTemplates()
	v.ValidateInputReferences()
	v.CheckReadmeExists()
	v.CheckChangelogExists()

	return v.ReportResults()
}

// AddError adds an error to the validator
func (v *BlueprintValidator) AddError(msg string) {
	v.Errors = append(v.Errors, msg)
}

// AddErrorf adds a formatted error to the validator
func (v *BlueprintValidator) AddErrorf(format string, args ...interface{}) {
	v.Errors = append(v.Errors, fmt.Sprintf(format, args...))
}

// AddWarning adds a warning to the validator
func (v *BlueprintValidator) AddWarning(msg string) {
	v.Warnings = append(v.Warnings, msg)
}

// AddWarningf adds a formatted warning to the validator
func (v *BlueprintValidator) AddWarningf(format string, args ...interface{}) {
	v.Warnings = append(v.Warnings, fmt.Sprintf(format, args...))
}

// AddCategorizedError adds an error with category and optional path
func (v *BlueprintValidator) AddCategorizedError(category ErrorCategory, path, msg string) {
	v.CategorizedErrors = append(v.CategorizedErrors, CategorizedError{
		Category: category,
		Path:     path,
		Message:  msg,
	})
	// Also add to legacy Errors slice for backward compatibility
	if path != "" {
		v.Errors = append(v.Errors, fmt.Sprintf("%s: %s", path, msg))
	} else {
		v.Errors = append(v.Errors, msg)
	}
}

// AddCategorizedErrorf adds a formatted error with category and optional path
func (v *BlueprintValidator) AddCategorizedErrorf(category ErrorCategory, path, format string, args ...interface{}) {
	v.AddCategorizedError(category, path, fmt.Sprintf(format, args...))
}

// AddCategorizedWarning adds a warning with category and optional path
func (v *BlueprintValidator) AddCategorizedWarning(category ErrorCategory, path, msg string) {
	v.CategorizedWarnings = append(v.CategorizedWarnings, CategorizedWarning{
		Category: category,
		Path:     path,
		Message:  msg,
	})
	// Also add to legacy Warnings slice for backward compatibility
	if path != "" {
		v.Warnings = append(v.Warnings, fmt.Sprintf("%s: %s", path, msg))
	} else {
		v.Warnings = append(v.Warnings, msg)
	}
}

// AddCategorizedWarningf adds a formatted warning with category and optional path
func (v *BlueprintValidator) AddCategorizedWarningf(category ErrorCategory, path, format string, args ...interface{}) {
	v.AddCategorizedWarning(category, path, fmt.Sprintf(format, args...))
}

// GetErrorsByCategory returns all errors matching the specified categories
func (v *BlueprintValidator) GetErrorsByCategory(categories ...ErrorCategory) []CategorizedError {
	return FilterErrorsByCategory(v.CategorizedErrors, categories...)
}

// GetWarningsByCategory returns all warnings matching the specified categories
func (v *BlueprintValidator) GetWarningsByCategory(categories ...ErrorCategory) []CategorizedWarning {
	return FilterWarningsByCategory(v.CategorizedWarnings, categories...)
}

// ErrorCounts returns a map of category to error count
func (v *BlueprintValidator) ErrorCounts() map[ErrorCategory]int {
	return CountByCategory(v.CategorizedErrors)
}

// WarningCounts returns a map of category to warning count
func (v *BlueprintValidator) WarningCounts() map[ErrorCategory]int {
	return CountWarningsByCategory(v.CategorizedWarnings)
}

// HasCategoryErrors returns true if there are any errors in the specified categories
func (v *BlueprintValidator) HasCategoryErrors(categories ...ErrorCategory) bool {
	return len(v.GetErrorsByCategory(categories...)) > 0
}

// HasCategoryWarnings returns true if there are any warnings in the specified categories
func (v *BlueprintValidator) HasCategoryWarnings(categories ...ErrorCategory) bool {
	return len(v.GetWarningsByCategory(categories...)) > 0
}
