// Package errors provides a centralized error handling system with typed errors,
// error factories, and a registry pattern for consistent error creation and handling.
// This package is designed for the validate-blueprint-go validator.
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of an error for machine-readable classification.
type ErrorType int

const (
	// ErrorTypeUnknown is the default error type for unclassified errors.
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeSyntax covers YAML syntax and parsing errors.
	ErrorTypeSyntax
	// ErrorTypeSchema covers structural issues (missing keys, wrong types).
	ErrorTypeSchema
	// ErrorTypeValidation covers general validation errors.
	ErrorTypeValidation
	// ErrorTypeParsing covers file reading and parsing errors.
	ErrorTypeParsing
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
	// ErrorTypeInternal covers internal/unexpected errors.
	ErrorTypeInternal
)

// errorTypeNames maps error types to their string representations.
var errorTypeNames = map[ErrorType]string{
	ErrorTypeUnknown:       "unknown",
	ErrorTypeSyntax:        "syntax",
	ErrorTypeSchema:        "schema",
	ErrorTypeValidation:    "validation",
	ErrorTypeParsing:       "parsing",
	ErrorTypeReference:     "reference",
	ErrorTypeTemplate:      "template",
	ErrorTypeInput:         "input",
	ErrorTypeTrigger:       "trigger",
	ErrorTypeCondition:     "condition",
	ErrorTypeAction:        "action",
	ErrorTypeDocumentation: "documentation",
	ErrorTypeInternal:      "internal",
}

// String returns the string representation of the error type.
func (t ErrorType) String() string {
	if name, ok := errorTypeNames[t]; ok {
		return name
	}
	return "unknown"
}

// Error represents a typed error with additional context.
// It implements the error interface and supports error wrapping.
type Error struct {
	// Type is the category of the error for machine-readable classification.
	Type ErrorType
	// Code is an optional machine-readable error code (e.g., "missing_blueprint").
	Code string
	// Message is the human-readable error message.
	Message string
	// Path is the location in the blueprint (e.g., "blueprint.input.my_input").
	Path string
	// Cause is the underlying error, if any.
	Cause error
	// Details contains additional context about the error.
	Details map[string]any
}

// Error implements the error interface.
func (e *Error) Error() string {
	var prefix string
	if e.Code != "" {
		prefix = fmt.Sprintf("[%s] ", e.Code)
	}
	if e.Path != "" {
		if e.Cause != nil {
			return fmt.Sprintf("%s%s: %s: %v", prefix, e.Path, e.Message, e.Cause)
		}
		return fmt.Sprintf("%s%s: %s", prefix, e.Path, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s%s: %v", prefix, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s%s", prefix, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *Error) Unwrap() error {
	return e.Cause
}

// Is reports whether the error matches the target error.
func (e *Error) Is(target error) bool {
	var targetErr *Error
	if errors.As(target, &targetErr) {
		if e.Type == targetErr.Type {
			if e.Code != "" && targetErr.Code != "" {
				return e.Code == targetErr.Code
			}
			return true
		}
		return false
	}
	return false
}

// WithDetails returns a copy of the error with additional details.
func (e *Error) WithDetails(details map[string]any) *Error {
	newErr := *e
	newErr.Details = make(map[string]any)
	for k, v := range e.Details {
		newErr.Details[k] = v
	}
	for k, v := range details {
		newErr.Details[k] = v
	}
	return &newErr
}

// WithCause returns a copy of the error with the specified cause.
func (e *Error) WithCause(cause error) *Error {
	newErr := *e
	newErr.Cause = cause
	return &newErr
}

// WithPath returns a copy of the error with the specified path.
func (e *Error) WithPath(path string) *Error {
	newErr := *e
	newErr.Path = path
	return &newErr
}

// WithMessage returns a copy of the error with a new message.
func (e *Error) WithMessage(msg string) *Error {
	newErr := *e
	newErr.Message = msg
	return &newErr
}

// WithMessagef returns a copy of the error with a formatted message.
func (e *Error) WithMessagef(format string, args ...any) *Error {
	newErr := *e
	newErr.Message = fmt.Sprintf(format, args...)
	return &newErr
}

// New creates a new Error with the specified type and message.
func New(errType ErrorType, message string) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Details: make(map[string]any),
	}
}

// Newf creates a new Error with the specified type and formatted message.
func Newf(errType ErrorType, format string, args ...any) *Error {
	return &Error{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
		Details: make(map[string]any),
	}
}

// NewWithPath creates a new Error with the specified type, path, and message.
func NewWithPath(errType ErrorType, path, message string) *Error {
	return &Error{
		Type:    errType,
		Path:    path,
		Message: message,
		Details: make(map[string]any),
	}
}

// Wrap wraps an existing error with a typed Error.
func Wrap(errType ErrorType, cause error, message string) *Error {
	return &Error{
		Type:    errType,
		Message: message,
		Cause:   cause,
		Details: make(map[string]any),
	}
}

// Wrapf wraps an existing error with a typed Error and formatted message.
func Wrapf(errType ErrorType, cause error, format string, args ...any) *Error {
	return &Error{
		Type:    errType,
		Message: fmt.Sprintf(format, args...),
		Cause:   cause,
		Details: make(map[string]any),
	}
}

// GetType extracts the ErrorType from an error.
func GetType(err error) ErrorType {
	var typedErr *Error
	if errors.As(err, &typedErr) {
		return typedErr.Type
	}
	return ErrorTypeUnknown
}

// GetCode extracts the error code from an error.
func GetCode(err error) string {
	var typedErr *Error
	if errors.As(err, &typedErr) {
		return typedErr.Code
	}
	return ""
}

// GetPath extracts the path from an error.
func GetPath(err error) string {
	var typedErr *Error
	if errors.As(err, &typedErr) {
		return typedErr.Path
	}
	return ""
}

// IsType checks if an error is of a specific ErrorType.
func IsType(err error, errType ErrorType) bool {
	return GetType(err) == errType
}

// IsSyntax checks if an error is a syntax error.
func IsSyntax(err error) bool {
	return IsType(err, ErrorTypeSyntax)
}

// IsSchema checks if an error is a schema error.
func IsSchema(err error) bool {
	return IsType(err, ErrorTypeSchema)
}

// IsValidation checks if an error is a validation error.
func IsValidation(err error) bool {
	return IsType(err, ErrorTypeValidation)
}

// IsReference checks if an error is a reference error.
func IsReference(err error) bool {
	return IsType(err, ErrorTypeReference)
}

// IsTemplate checks if an error is a template error.
func IsTemplate(err error) bool {
	return IsType(err, ErrorTypeTemplate)
}

// IsInput checks if an error is an input error.
func IsInput(err error) bool {
	return IsType(err, ErrorTypeInput)
}

// IsTrigger checks if an error is a trigger error.
func IsTrigger(err error) bool {
	return IsType(err, ErrorTypeTrigger)
}

// IsCondition checks if an error is a condition error.
func IsCondition(err error) bool {
	return IsType(err, ErrorTypeCondition)
}

// IsAction checks if an error is an action error.
func IsAction(err error) bool {
	return IsType(err, ErrorTypeAction)
}

// IsDocumentation checks if an error is a documentation error.
func IsDocumentation(err error) bool {
	return IsType(err, ErrorTypeDocumentation)
}

// IsWarning returns true if this error should be treated as a warning rather than an error.
// This is determined by the IsWarningFlag field.
func (e *Error) IsWarning() bool {
	if val, ok := e.Details["is_warning"].(bool); ok {
		return val
	}
	return false
}

// AsWarning marks this error as a warning.
func (e *Error) AsWarning() *Error {
	return e.WithDetails(map[string]any{"is_warning": true})
}
