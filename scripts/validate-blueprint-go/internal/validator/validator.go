package validator

import "fmt"

// BlueprintValidator validates Home Assistant Blueprint YAML files
type BlueprintValidator struct {
	FilePath            string
	Errors              []string
	Warnings            []string
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
}

// New creates a new validator instance
func New(filePath string) *BlueprintValidator {
	return &BlueprintValidator{
		FilePath:            filePath,
		Errors:              []string{},
		Warnings:            []string{},
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
