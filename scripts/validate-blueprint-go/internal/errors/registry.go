// Package errors provides a centralized error handling system.
// This file contains the error registry and factory functions for blueprint validation.
package errors

import (
	"sync"
)

// ErrorDefinition holds the definition of a registered error.
type ErrorDefinition struct {
	// Code is the unique identifier for this error.
	Code string
	// Type is the category of the error.
	Type ErrorType
	// Message is the default message template for this error.
	Message string
}

// Registry holds registered error definitions for consistent error creation.
type Registry struct {
	mu          sync.RWMutex
	definitions map[string]ErrorDefinition
}

// NewRegistry creates a new error registry.
func NewRegistry() *Registry {
	return &Registry{
		definitions: make(map[string]ErrorDefinition),
	}
}

// Register adds an error definition to the registry.
func (r *Registry) Register(def ErrorDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.definitions[def.Code] = def
}

// Get retrieves an error definition by code.
func (r *Registry) Get(code string) *ErrorDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if def, ok := r.definitions[code]; ok {
		return &def
	}
	return nil
}

// Create creates a new Error from a registered definition.
func (r *Registry) Create(code string) *Error {
	def := r.Get(code)
	if def == nil {
		return nil
	}
	return &Error{
		Type:    def.Type,
		Code:    def.Code,
		Message: def.Message,
		Details: make(map[string]any),
	}
}

// CreateWithPath creates a new Error from a registered definition with a path.
func (r *Registry) CreateWithPath(code, path string) *Error {
	def := r.Get(code)
	if def == nil {
		return nil
	}
	return &Error{
		Type:    def.Type,
		Code:    def.Code,
		Path:    path,
		Message: def.Message,
		Details: make(map[string]any),
	}
}

// List returns all registered error definitions.
func (r *Registry) List() []ErrorDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	defs := make([]ErrorDefinition, 0, len(r.definitions))
	for _, def := range r.definitions {
		defs = append(defs, def)
	}
	return defs
}

// DefaultRegistry is the global error registry.
var DefaultRegistry = NewRegistry()

// Register adds an error definition to the default registry.
func Register(def ErrorDefinition) {
	DefaultRegistry.Register(def)
}

// Create creates a new Error from the default registry.
func Create(code string) *Error {
	return DefaultRegistry.Create(code)
}

// CreateWithPath creates a new Error from the default registry with a path.
func CreateWithPath(code, path string) *Error {
	return DefaultRegistry.CreateWithPath(code, path)
}

// Common error codes for blueprint validation.
const (
	// Syntax errors
	CodeYAMLSyntax     = "yaml_syntax"
	CodeInvalidYAML    = "invalid_yaml"
	CodeFileReadError  = "file_read_error"
	CodeFileNotFound   = "file_not_found"
	CodeFileParseError = "file_parse_error"

	// Schema errors
	CodeMissingBlueprint      = "missing_blueprint"
	CodeMissingName           = "missing_name"
	CodeMissingDomain         = "missing_domain"
	CodeInvalidDomain         = "invalid_domain"
	CodeMissingDescription    = "missing_description"
	CodeInvalidStructure      = "invalid_structure"
	CodeUnexpectedKey         = "unexpected_key"
	CodeMissingRequiredField  = "missing_required_field"
	CodeInvalidFieldType      = "invalid_field_type"
	CodeInvalidMode           = "invalid_mode"
	CodeInvalidMaxConcurrency = "invalid_max_concurrency"

	// Input errors
	CodeInvalidInput         = "invalid_input"
	CodeInvalidSelector      = "invalid_selector"
	CodeMissingInputName     = "missing_input_name"
	CodeInvalidInputDefault  = "invalid_input_default"
	CodeUnknownSelectorType  = "unknown_selector_type"
	CodeInvalidNumberRange   = "invalid_number_range"
	CodeInvalidSelectOptions = "invalid_select_options"

	// Reference errors
	CodeUndefinedInput     = "undefined_input"
	CodeUnusedInput        = "unused_input"
	CodeInvalidReference   = "invalid_reference"
	CodeCircularReference  = "circular_reference"
	CodeUndefinedVariable  = "undefined_variable"
	CodeInvalidVariableDef = "invalid_variable_def"

	// Template errors
	CodeInvalidTemplate         = "invalid_template"
	CodeUnclosedTemplateTag     = "unclosed_template_tag"
	CodeInvalidTemplateVariable = "invalid_template_variable"
	CodeTemplateParseError      = "template_parse_error"

	// Trigger errors
	CodeInvalidTrigger        = "invalid_trigger"
	CodeMissingTriggerType    = "missing_trigger_type"
	CodeUnknownTriggerType    = "unknown_trigger_type"
	CodeMissingTriggerField   = "missing_trigger_field"
	CodeInvalidTriggerConfig  = "invalid_trigger_config"
	CodeDeprecatedTriggerType = "deprecated_trigger_type"

	// Condition errors
	CodeInvalidCondition        = "invalid_condition"
	CodeMissingConditionType    = "missing_condition_type"
	CodeUnknownConditionType    = "unknown_condition_type"
	CodeMissingConditionField   = "missing_condition_field"
	CodeInvalidConditionConfig  = "invalid_condition_config"
	CodeDeprecatedConditionType = "deprecated_condition_type"

	// Action errors
	CodeInvalidAction        = "invalid_action"
	CodeMissingActionService = "missing_action_service"
	CodeInvalidServiceFormat = "invalid_service_format"
	CodeInvalidActionConfig  = "invalid_action_config"
	CodeDeprecatedActionType = "deprecated_action_type"

	// Documentation errors
	CodeMissingReadme        = "missing_readme"
	CodeMissingChangelog     = "missing_changelog"
	CodeVersionMismatch      = "version_mismatch"
	CodeMissingSourceURL     = "missing_source_url"
	CodeInvalidSourceURL     = "invalid_source_url"
	CodeMissingDocumentation = "missing_documentation"
)

// init registers the default error definitions.
func init() {
	// Syntax errors
	Register(ErrorDefinition{Code: CodeYAMLSyntax, Type: ErrorTypeSyntax, Message: "YAML syntax error"})
	Register(ErrorDefinition{Code: CodeInvalidYAML, Type: ErrorTypeSyntax, Message: "invalid YAML"})
	Register(ErrorDefinition{Code: CodeFileReadError, Type: ErrorTypeParsing, Message: "failed to read file"})
	Register(ErrorDefinition{Code: CodeFileNotFound, Type: ErrorTypeParsing, Message: "file not found"})
	Register(ErrorDefinition{Code: CodeFileParseError, Type: ErrorTypeParsing, Message: "failed to parse file"})

	// Schema errors
	Register(ErrorDefinition{Code: CodeMissingBlueprint, Type: ErrorTypeSchema, Message: "missing 'blueprint' section"})
	Register(ErrorDefinition{Code: CodeMissingName, Type: ErrorTypeSchema, Message: "missing 'name' field"})
	Register(ErrorDefinition{Code: CodeMissingDomain, Type: ErrorTypeSchema, Message: "missing 'domain' field"})
	Register(ErrorDefinition{Code: CodeInvalidDomain, Type: ErrorTypeSchema, Message: "invalid domain value"})
	Register(ErrorDefinition{Code: CodeMissingDescription, Type: ErrorTypeSchema, Message: "missing 'description' field"})
	Register(ErrorDefinition{Code: CodeInvalidStructure, Type: ErrorTypeSchema, Message: "invalid blueprint structure"})
	Register(ErrorDefinition{Code: CodeUnexpectedKey, Type: ErrorTypeSchema, Message: "unexpected key"})
	Register(ErrorDefinition{Code: CodeMissingRequiredField, Type: ErrorTypeSchema, Message: "missing required field"})
	Register(ErrorDefinition{Code: CodeInvalidFieldType, Type: ErrorTypeSchema, Message: "invalid field type"})
	Register(ErrorDefinition{Code: CodeInvalidMode, Type: ErrorTypeSchema, Message: "invalid mode value"})
	Register(ErrorDefinition{Code: CodeInvalidMaxConcurrency, Type: ErrorTypeSchema, Message: "invalid max concurrency value"})

	// Input errors
	Register(ErrorDefinition{Code: CodeInvalidInput, Type: ErrorTypeInput, Message: "invalid input definition"})
	Register(ErrorDefinition{Code: CodeInvalidSelector, Type: ErrorTypeInput, Message: "invalid selector"})
	Register(ErrorDefinition{Code: CodeMissingInputName, Type: ErrorTypeInput, Message: "missing input name"})
	Register(ErrorDefinition{Code: CodeInvalidInputDefault, Type: ErrorTypeInput, Message: "invalid input default value"})
	Register(ErrorDefinition{Code: CodeUnknownSelectorType, Type: ErrorTypeInput, Message: "unknown selector type"})
	Register(ErrorDefinition{Code: CodeInvalidNumberRange, Type: ErrorTypeInput, Message: "invalid number range"})
	Register(ErrorDefinition{Code: CodeInvalidSelectOptions, Type: ErrorTypeInput, Message: "invalid select options"})

	// Reference errors
	Register(ErrorDefinition{Code: CodeUndefinedInput, Type: ErrorTypeReference, Message: "undefined input reference"})
	Register(ErrorDefinition{Code: CodeUnusedInput, Type: ErrorTypeReference, Message: "unused input"})
	Register(ErrorDefinition{Code: CodeInvalidReference, Type: ErrorTypeReference, Message: "invalid reference"})
	Register(ErrorDefinition{Code: CodeCircularReference, Type: ErrorTypeReference, Message: "circular reference detected"})
	Register(ErrorDefinition{Code: CodeUndefinedVariable, Type: ErrorTypeReference, Message: "undefined variable"})
	Register(ErrorDefinition{Code: CodeInvalidVariableDef, Type: ErrorTypeReference, Message: "invalid variable definition"})

	// Template errors
	Register(ErrorDefinition{Code: CodeInvalidTemplate, Type: ErrorTypeTemplate, Message: "invalid template"})
	Register(ErrorDefinition{Code: CodeUnclosedTemplateTag, Type: ErrorTypeTemplate, Message: "unclosed template tag"})
	Register(ErrorDefinition{Code: CodeInvalidTemplateVariable, Type: ErrorTypeTemplate, Message: "invalid template variable"})
	Register(ErrorDefinition{Code: CodeTemplateParseError, Type: ErrorTypeTemplate, Message: "template parse error"})

	// Trigger errors
	Register(ErrorDefinition{Code: CodeInvalidTrigger, Type: ErrorTypeTrigger, Message: "invalid trigger"})
	Register(ErrorDefinition{Code: CodeMissingTriggerType, Type: ErrorTypeTrigger, Message: "missing trigger type"})
	Register(ErrorDefinition{Code: CodeUnknownTriggerType, Type: ErrorTypeTrigger, Message: "unknown trigger type"})
	Register(ErrorDefinition{Code: CodeMissingTriggerField, Type: ErrorTypeTrigger, Message: "missing required trigger field"})
	Register(ErrorDefinition{Code: CodeInvalidTriggerConfig, Type: ErrorTypeTrigger, Message: "invalid trigger configuration"})
	Register(ErrorDefinition{Code: CodeDeprecatedTriggerType, Type: ErrorTypeTrigger, Message: "deprecated trigger type"})

	// Condition errors
	Register(ErrorDefinition{Code: CodeInvalidCondition, Type: ErrorTypeCondition, Message: "invalid condition"})
	Register(ErrorDefinition{Code: CodeMissingConditionType, Type: ErrorTypeCondition, Message: "missing condition type"})
	Register(ErrorDefinition{Code: CodeUnknownConditionType, Type: ErrorTypeCondition, Message: "unknown condition type"})
	Register(ErrorDefinition{Code: CodeMissingConditionField, Type: ErrorTypeCondition, Message: "missing required condition field"})
	Register(ErrorDefinition{Code: CodeInvalidConditionConfig, Type: ErrorTypeCondition, Message: "invalid condition configuration"})
	Register(ErrorDefinition{Code: CodeDeprecatedConditionType, Type: ErrorTypeCondition, Message: "deprecated condition type"})

	// Action errors
	Register(ErrorDefinition{Code: CodeInvalidAction, Type: ErrorTypeAction, Message: "invalid action"})
	Register(ErrorDefinition{Code: CodeMissingActionService, Type: ErrorTypeAction, Message: "missing action service"})
	Register(ErrorDefinition{Code: CodeInvalidServiceFormat, Type: ErrorTypeAction, Message: "invalid service format"})
	Register(ErrorDefinition{Code: CodeInvalidActionConfig, Type: ErrorTypeAction, Message: "invalid action configuration"})
	Register(ErrorDefinition{Code: CodeDeprecatedActionType, Type: ErrorTypeAction, Message: "deprecated action type"})

	// Documentation errors
	Register(ErrorDefinition{Code: CodeMissingReadme, Type: ErrorTypeDocumentation, Message: "missing README file"})
	Register(ErrorDefinition{Code: CodeMissingChangelog, Type: ErrorTypeDocumentation, Message: "missing CHANGELOG file"})
	Register(ErrorDefinition{Code: CodeVersionMismatch, Type: ErrorTypeDocumentation, Message: "version mismatch"})
	Register(ErrorDefinition{Code: CodeMissingSourceURL, Type: ErrorTypeDocumentation, Message: "missing source_url"})
	Register(ErrorDefinition{Code: CodeInvalidSourceURL, Type: ErrorTypeDocumentation, Message: "invalid source_url"})
	Register(ErrorDefinition{Code: CodeMissingDocumentation, Type: ErrorTypeDocumentation, Message: "missing documentation"})
}

// Convenience factory functions for common validation errors.

// ErrYAMLSyntax creates a YAML syntax error.
func ErrYAMLSyntax(cause error) *Error {
	return Create(CodeYAMLSyntax).WithCause(cause)
}

// ErrFileNotFound creates a file not found error.
func ErrFileNotFound(path string) *Error {
	return Create(CodeFileNotFound).WithPath(path)
}

// ErrFileReadError creates a file read error.
func ErrFileReadError(path string, cause error) *Error {
	return Create(CodeFileReadError).WithPath(path).WithCause(cause)
}

// ErrMissingBlueprint creates a missing blueprint section error.
func ErrMissingBlueprint() *Error {
	return Create(CodeMissingBlueprint)
}

// ErrMissingField creates a missing required field error.
func ErrMissingField(path, field string) *Error {
	return Create(CodeMissingRequiredField).WithPath(path).WithMessagef("missing required field: %s", field)
}

// ErrInvalidFieldType creates an invalid field type error.
func ErrInvalidFieldType(path, expected, actual string) *Error {
	return Create(CodeInvalidFieldType).WithPath(path).WithMessagef("expected %s, got %s", expected, actual)
}

// ErrUndefinedInput creates an undefined input reference error.
func ErrUndefinedInput(path, inputName string) *Error {
	return Create(CodeUndefinedInput).WithPath(path).WithMessagef("undefined input: %s", inputName).WithDetails(map[string]any{
		"input_name": inputName,
	})
}

// ErrUnusedInput creates an unused input warning.
func ErrUnusedInput(inputName string) *Error {
	return Create(CodeUnusedInput).WithMessagef("input '%s' is defined but never used", inputName).WithDetails(map[string]any{
		"input_name": inputName,
	})
}

// ErrInvalidTemplate creates an invalid template error.
func ErrInvalidTemplate(path, message string) *Error {
	return Create(CodeInvalidTemplate).WithPath(path).WithMessage(message)
}

// ErrMissingTriggerType creates a missing trigger type error.
func ErrMissingTriggerType(path string) *Error {
	return Create(CodeMissingTriggerType).WithPath(path)
}

// ErrInvalidTrigger creates an invalid trigger configuration error.
func ErrInvalidTrigger(path, message string) *Error {
	return Create(CodeInvalidTriggerConfig).WithPath(path).WithMessage(message)
}

// ErrMissingConditionType creates a missing condition type error.
func ErrMissingConditionType(path string) *Error {
	return Create(CodeMissingConditionType).WithPath(path)
}

// ErrInvalidCondition creates an invalid condition configuration error.
func ErrInvalidCondition(path, message string) *Error {
	return Create(CodeInvalidConditionConfig).WithPath(path).WithMessage(message)
}

// ErrMissingActionService creates a missing action service error.
func ErrMissingActionService(path string) *Error {
	return Create(CodeMissingActionService).WithPath(path)
}

// ErrInvalidAction creates an invalid action configuration error.
func ErrInvalidAction(path, message string) *Error {
	return Create(CodeInvalidActionConfig).WithPath(path).WithMessage(message)
}

// ErrMissingReadme creates a missing README error.
func ErrMissingReadme() *Error {
	return Create(CodeMissingReadme)
}

// ErrMissingChangelog creates a missing CHANGELOG error.
func ErrMissingChangelog() *Error {
	return Create(CodeMissingChangelog)
}

// ErrVersionMismatch creates a version mismatch error.
func ErrVersionMismatch(expected, actual string) *Error {
	return Create(CodeVersionMismatch).WithMessagef("version mismatch: expected %s, got %s", expected, actual).WithDetails(map[string]any{
		"expected": expected,
		"actual":   actual,
	})
}
