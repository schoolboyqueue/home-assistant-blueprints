// Package errors provides a centralized error handling system with typed errors,
// error factories, and a registry pattern for consistent error creation and handling.
// This package extends the shared errors package with validate-blueprint-go specific error types.
package errors

import (
	"github.com/home-assistant-blueprints/shared/errors"
)

// Re-export the base Error type and related types from the shared package
type (
	// Error is a typed error with additional context.
	Error = errors.Error
	// ErrorDefinition holds the definition of a registered error.
	ErrorDefinition = errors.ErrorDefinition
	// Registry holds registered error definitions.
	Registry = errors.Registry
)

// ErrorType represents the category of an error for machine-readable classification.
// This extends the base ErrorType with validate-blueprint-go specific types.
type ErrorType = errors.ErrorType

// Base error types from shared package
const (
	ErrorTypeUnknown    = errors.ErrorTypeUnknown
	ErrorTypeValidation = errors.ErrorTypeValidation
	ErrorTypeParsing    = errors.ErrorTypeParsing
	ErrorTypeInternal   = errors.ErrorTypeInternal
)

// validate-blueprint-go specific error types
const (
	// ErrorTypeSyntax covers YAML syntax and parsing errors.
	ErrorTypeSyntax ErrorType = errors.ErrorTypeAppBase + iota
	// ErrorTypeSchema covers structural issues (missing keys, wrong types).
	ErrorTypeSchema
	// ErrorTypeReference covers undefined or unused input references.
	ErrorTypeReference
	// ErrorTypeTemplate covers Jinja2 template syntax errors.
	ErrorTypeTemplate
	// ErrorTypeInput covers input definition and selector issues.
	ErrorTypeInput
	// ErrorTypeTrigger covers trigger validation errors.
	ErrorTypeTrigger
	// ErrorTypeCondition covers condition validation errors.
	ErrorTypeCondition
	// ErrorTypeAction covers action/service call validation errors.
	ErrorTypeAction
	// ErrorTypeDocumentation covers README/CHANGELOG and documentation issues.
	ErrorTypeDocumentation
)

func init() {
	// Register custom error type names
	errors.RegisterTypeName(ErrorTypeSyntax, "syntax")
	errors.RegisterTypeName(ErrorTypeSchema, "schema")
	errors.RegisterTypeName(ErrorTypeReference, "reference")
	errors.RegisterTypeName(ErrorTypeTemplate, "template")
	errors.RegisterTypeName(ErrorTypeInput, "input")
	errors.RegisterTypeName(ErrorTypeTrigger, "trigger")
	errors.RegisterTypeName(ErrorTypeCondition, "condition")
	errors.RegisterTypeName(ErrorTypeAction, "action")
	errors.RegisterTypeName(ErrorTypeDocumentation, "documentation")
}

// Re-export base functions from shared package
var (
	New               = errors.New
	Newf              = errors.Newf
	NewWithCode       = errors.NewWithCode
	NewWithPath       = errors.NewWithPath
	Wrap              = errors.Wrap
	Wrapf             = errors.Wrapf
	GetType           = errors.GetType
	GetCode           = errors.GetCode
	GetPath           = errors.GetPath
	GetDetails        = errors.GetDetails
	IsType            = errors.IsType
	IsParsing         = errors.IsParsing
	IsInternal        = errors.IsInternal
	NewRegistry       = errors.NewRegistry
	Register          = errors.Register
	Create            = errors.Create
	CreateWithMessage = errors.CreateWithMessage
	CreateWithCause   = errors.CreateWithCause
	CreateWithPath    = errors.CreateWithPath
)

// Use shared DefaultRegistry
var DefaultRegistry = errors.DefaultRegistry

// validate-blueprint-go specific type check functions

// IsSyntax checks if an error is a syntax error.
func IsSyntax(err error) bool {
	return errors.IsType(err, ErrorTypeSyntax)
}

// IsSchema checks if an error is a schema error.
func IsSchema(err error) bool {
	return errors.IsType(err, ErrorTypeSchema)
}

// IsValidation checks if an error is a validation error.
func IsValidation(err error) bool {
	return errors.IsType(err, ErrorTypeValidation)
}

// IsReference checks if an error is a reference error.
func IsReference(err error) bool {
	return errors.IsType(err, ErrorTypeReference)
}

// IsTemplate checks if an error is a template error.
func IsTemplate(err error) bool {
	return errors.IsType(err, ErrorTypeTemplate)
}

// IsInput checks if an error is an input error.
func IsInput(err error) bool {
	return errors.IsType(err, ErrorTypeInput)
}

// IsTrigger checks if an error is a trigger error.
func IsTrigger(err error) bool {
	return errors.IsType(err, ErrorTypeTrigger)
}

// IsCondition checks if an error is a condition error.
func IsCondition(err error) bool {
	return errors.IsType(err, ErrorTypeCondition)
}

// IsAction checks if an error is an action error.
func IsAction(err error) bool {
	return errors.IsType(err, ErrorTypeAction)
}

// IsDocumentation checks if an error is a documentation error.
func IsDocumentation(err error) bool {
	return errors.IsType(err, ErrorTypeDocumentation)
}
