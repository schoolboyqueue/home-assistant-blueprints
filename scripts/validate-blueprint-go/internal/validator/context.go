// Package validator provides validation for Home Assistant Blueprint YAML files.
package validator

// ValidationContext encapsulates all validator state including errors, warnings,
// used inputs, and blueprint data. This provides a cleaner API for validation rules
// by centralizing all state management in one place.
type ValidationContext struct {
	// Errors contains legacy flat error strings for backward compatibility
	Errors []string
	// Warnings contains legacy flat warning strings for backward compatibility
	Warnings []string
	// CategorizedErrors contains errors with category and path information
	CategorizedErrors []CategorizedError
	// CategorizedWarnings contains warnings with category and path information
	CategorizedWarnings []CategorizedWarning

	// Data contains the parsed YAML blueprint data as a raw map for backward compatibility.
	// Use TypedData for type-safe access to blueprint fields.
	Data RawData

	// TypedData provides strongly-typed access to blueprint fields.
	// This is populated alongside Data for gradual migration to type-safe code.
	TypedData *BlueprintData

	// Input tracking
	// DefinedInputs tracks input names that are defined in the blueprint
	DefinedInputs map[string]bool
	// UsedInputs tracks input names that are referenced via !input
	UsedInputs map[string]bool
	// InputDefaults stores default values for inputs
	InputDefaults map[string]interface{}
	// InputSelectors stores selector configuration for each input (raw map for backward compatibility)
	InputSelectors map[string]RawData
	// TypedInputSelectors stores typed selector configuration for each input
	TypedInputSelectors map[string]*Selector
	// EntityInputs tracks inputs with entity selectors
	EntityInputs map[string]bool
	// InputDatetimeInputs tracks inputs with input_datetime entity domain
	InputDatetimeInputs map[string]bool

	// Variable tracking
	// DefinedVariables tracks variable names defined in the variables section
	DefinedVariables map[string]bool
	// JoinVariables tracks variables that use join() or similar patterns
	JoinVariables map[string]bool
	// NonzeroDefaultVars tracks variables with non-zero default values
	NonzeroDefaultVars map[string]bool
}

// NewValidationContext creates a new ValidationContext with all maps initialized.
func NewValidationContext() *ValidationContext {
	return &ValidationContext{
		Errors:              []string{},
		Warnings:            []string{},
		CategorizedErrors:   []CategorizedError{},
		CategorizedWarnings: []CategorizedWarning{},
		Data:                make(RawData),
		TypedData:           nil,
		DefinedInputs:       make(map[string]bool),
		UsedInputs:          make(map[string]bool),
		InputDefaults:       make(map[string]interface{}),
		InputSelectors:      make(map[string]RawData),
		TypedInputSelectors: make(map[string]*Selector),
		EntityInputs:        make(map[string]bool),
		InputDatetimeInputs: make(map[string]bool),
		DefinedVariables:    make(map[string]bool),
		JoinVariables:       make(map[string]bool),
		NonzeroDefaultVars:  make(map[string]bool),
	}
}

// NewValidationContextWithData creates a new ValidationContext initialized with the provided data.
func NewValidationContextWithData(data RawData) *ValidationContext {
	ctx := NewValidationContext()
	ctx.Data = data
	// Also initialize TypedData with a reference to the raw data
	ctx.TypedData = &BlueprintData{Raw: data}
	return ctx
}

// Reset clears all errors, warnings, and tracking maps while preserving the data.
// This is useful for re-validating the same blueprint data.
func (ctx *ValidationContext) Reset() {
	ctx.Errors = []string{}
	ctx.Warnings = []string{}
	ctx.CategorizedErrors = []CategorizedError{}
	ctx.CategorizedWarnings = []CategorizedWarning{}
	ctx.DefinedInputs = make(map[string]bool)
	ctx.UsedInputs = make(map[string]bool)
	ctx.InputDefaults = make(map[string]interface{})
	ctx.InputSelectors = make(map[string]RawData)
	ctx.TypedInputSelectors = make(map[string]*Selector)
	ctx.EntityInputs = make(map[string]bool)
	ctx.InputDatetimeInputs = make(map[string]bool)
	ctx.DefinedVariables = make(map[string]bool)
	ctx.JoinVariables = make(map[string]bool)
	ctx.NonzeroDefaultVars = make(map[string]bool)
}

// HasErrors returns true if there are any errors (categorized or legacy).
func (ctx *ValidationContext) HasErrors() bool {
	return len(ctx.Errors) > 0 || len(ctx.CategorizedErrors) > 0
}

// HasWarnings returns true if there are any warnings (categorized or legacy).
func (ctx *ValidationContext) HasWarnings() bool {
	return len(ctx.Warnings) > 0 || len(ctx.CategorizedWarnings) > 0
}

// ErrorCount returns the total number of errors.
func (ctx *ValidationContext) ErrorCount() int {
	return len(ctx.CategorizedErrors)
}

// WarningCount returns the total number of warnings.
func (ctx *ValidationContext) WarningCount() int {
	return len(ctx.CategorizedWarnings)
}

// DefinedInputCount returns the number of defined inputs.
func (ctx *ValidationContext) DefinedInputCount() int {
	return len(ctx.DefinedInputs)
}

// UsedInputCount returns the number of used input references.
func (ctx *ValidationContext) UsedInputCount() int {
	return len(ctx.UsedInputs)
}

// IsInputDefined checks if an input name is defined in the blueprint.
func (ctx *ValidationContext) IsInputDefined(name string) bool {
	return ctx.DefinedInputs[name]
}

// IsInputUsed checks if an input name is referenced via !input.
func (ctx *ValidationContext) IsInputUsed(name string) bool {
	return ctx.UsedInputs[name]
}

// MarkInputDefined marks an input name as defined.
func (ctx *ValidationContext) MarkInputDefined(name string) {
	ctx.DefinedInputs[name] = true
}

// MarkInputUsed marks an input name as used (referenced).
func (ctx *ValidationContext) MarkInputUsed(name string) {
	ctx.UsedInputs[name] = true
}

// SetInputDefault sets the default value for an input.
func (ctx *ValidationContext) SetInputDefault(name string, value interface{}) {
	ctx.InputDefaults[name] = value
}

// GetInputDefault returns the default value for an input and whether it exists.
func (ctx *ValidationContext) GetInputDefault(name string) (interface{}, bool) {
	val, ok := ctx.InputDefaults[name]
	return val, ok
}

// SetInputSelector sets the selector configuration for an input.
func (ctx *ValidationContext) SetInputSelector(name string, selector RawData) {
	ctx.InputSelectors[name] = selector
}

// SetTypedInputSelector sets the typed selector configuration for an input.
func (ctx *ValidationContext) SetTypedInputSelector(name string, selector *Selector) {
	ctx.TypedInputSelectors[name] = selector
}

// GetInputSelector returns the selector configuration for an input and whether it exists.
func (ctx *ValidationContext) GetInputSelector(name string) (RawData, bool) {
	sel, ok := ctx.InputSelectors[name]
	return sel, ok
}

// GetTypedInputSelector returns the typed selector configuration for an input and whether it exists.
func (ctx *ValidationContext) GetTypedInputSelector(name string) (*Selector, bool) {
	sel, ok := ctx.TypedInputSelectors[name]
	return sel, ok
}

// MarkEntityInput marks an input as having an entity selector.
func (ctx *ValidationContext) MarkEntityInput(name string) {
	ctx.EntityInputs[name] = true
}

// IsEntityInput checks if an input has an entity selector.
func (ctx *ValidationContext) IsEntityInput(name string) bool {
	return ctx.EntityInputs[name]
}

// MarkInputDatetimeInput marks an input as having an input_datetime entity domain.
func (ctx *ValidationContext) MarkInputDatetimeInput(name string) {
	ctx.InputDatetimeInputs[name] = true
}

// IsInputDatetimeInput checks if an input has an input_datetime entity domain.
func (ctx *ValidationContext) IsInputDatetimeInput(name string) bool {
	return ctx.InputDatetimeInputs[name]
}

// MarkVariableDefined marks a variable name as defined.
func (ctx *ValidationContext) MarkVariableDefined(name string) {
	ctx.DefinedVariables[name] = true
}

// IsVariableDefined checks if a variable name is defined.
func (ctx *ValidationContext) IsVariableDefined(name string) bool {
	return ctx.DefinedVariables[name]
}

// MarkJoinVariable marks a variable as using join() pattern.
func (ctx *ValidationContext) MarkJoinVariable(name string) {
	ctx.JoinVariables[name] = true
}

// IsJoinVariable checks if a variable uses join() pattern.
func (ctx *ValidationContext) IsJoinVariable(name string) bool {
	return ctx.JoinVariables[name]
}

// MarkNonzeroDefaultVar marks a variable as having a non-zero default.
func (ctx *ValidationContext) MarkNonzeroDefaultVar(name string) {
	ctx.NonzeroDefaultVars[name] = true
}

// IsNonzeroDefaultVar checks if a variable has a non-zero default.
func (ctx *ValidationContext) IsNonzeroDefaultVar(name string) bool {
	return ctx.NonzeroDefaultVars[name]
}

// GetUndefinedInputRefs returns input names that are used but not defined.
func (ctx *ValidationContext) GetUndefinedInputRefs() []string {
	var undefined []string
	for name := range ctx.UsedInputs {
		if !ctx.DefinedInputs[name] {
			undefined = append(undefined, name)
		}
	}
	return undefined
}

// GetUnusedInputs returns input names that are defined but not used.
func (ctx *ValidationContext) GetUnusedInputs() []string {
	var unused []string
	for name := range ctx.DefinedInputs {
		if !ctx.UsedInputs[name] {
			unused = append(unused, name)
		}
	}
	return unused
}

// Clone creates a deep copy of the ValidationContext.
// Note: Data is shared (not deep copied) for efficiency.
func (ctx *ValidationContext) Clone() *ValidationContext {
	clone := &ValidationContext{
		Errors:              make([]string, len(ctx.Errors)),
		Warnings:            make([]string, len(ctx.Warnings)),
		CategorizedErrors:   make([]CategorizedError, len(ctx.CategorizedErrors)),
		CategorizedWarnings: make([]CategorizedWarning, len(ctx.CategorizedWarnings)),
		Data:                ctx.Data,      // Shared reference
		TypedData:           ctx.TypedData, // Shared reference
		DefinedInputs:       make(map[string]bool),
		UsedInputs:          make(map[string]bool),
		InputDefaults:       make(map[string]interface{}),
		InputSelectors:      make(map[string]RawData),
		TypedInputSelectors: make(map[string]*Selector),
		EntityInputs:        make(map[string]bool),
		InputDatetimeInputs: make(map[string]bool),
		DefinedVariables:    make(map[string]bool),
		JoinVariables:       make(map[string]bool),
		NonzeroDefaultVars:  make(map[string]bool),
	}

	copy(clone.Errors, ctx.Errors)
	copy(clone.Warnings, ctx.Warnings)
	copy(clone.CategorizedErrors, ctx.CategorizedErrors)
	copy(clone.CategorizedWarnings, ctx.CategorizedWarnings)

	for k, v := range ctx.DefinedInputs {
		clone.DefinedInputs[k] = v
	}
	for k, v := range ctx.UsedInputs {
		clone.UsedInputs[k] = v
	}
	for k, v := range ctx.InputDefaults {
		clone.InputDefaults[k] = v
	}
	for k, v := range ctx.InputSelectors {
		clone.InputSelectors[k] = v
	}
	for k, v := range ctx.TypedInputSelectors {
		clone.TypedInputSelectors[k] = v
	}
	for k, v := range ctx.EntityInputs {
		clone.EntityInputs[k] = v
	}
	for k, v := range ctx.InputDatetimeInputs {
		clone.InputDatetimeInputs[k] = v
	}
	for k, v := range ctx.DefinedVariables {
		clone.DefinedVariables[k] = v
	}
	for k, v := range ctx.JoinVariables {
		clone.JoinVariables[k] = v
	}
	for k, v := range ctx.NonzeroDefaultVars {
		clone.NonzeroDefaultVars[k] = v
	}

	return clone
}
