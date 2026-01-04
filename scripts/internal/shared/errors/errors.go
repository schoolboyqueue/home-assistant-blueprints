// Package errors provides a centralized error handling system with typed errors,
// error factories, and a registry pattern for consistent error creation and handling.
// This package provides the base infrastructure that can be extended by specific
// applications with their own error types and codes.
package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of an error for machine-readable classification.
// Applications should define their own constants starting from ErrorTypeAppBase.
type ErrorType int

const (
	// ErrorTypeUnknown is the default error type for unclassified errors.
	ErrorTypeUnknown ErrorType = iota
	// ErrorTypeValidation covers input validation and argument errors.
	ErrorTypeValidation
	// ErrorTypeParsing covers JSON, YAML, and other parsing errors.
	ErrorTypeParsing
	// ErrorTypeInternal covers internal/unexpected errors.
	ErrorTypeInternal

	// ErrorTypeAppBase is the starting point for application-specific error types.
	// Applications should define their error types as:
	//   const MyErrorType = errors.ErrorTypeAppBase + iota
	ErrorTypeAppBase ErrorType = 100
)

// baseTypeNames maps base error types to their string representations.
var baseTypeNames = map[ErrorType]string{
	ErrorTypeUnknown:    "unknown",
	ErrorTypeValidation: "validation",
	ErrorTypeParsing:    "parsing",
	ErrorTypeInternal:   "internal",
}

// customTypeNames holds application-registered type names.
var customTypeNames = make(map[ErrorType]string)

// RegisterTypeName registers a custom error type name.
// This allows applications to define their own error types with proper string representations.
func RegisterTypeName(t ErrorType, name string) {
	customTypeNames[t] = name
}

// String returns the string representation of the error type.
func (t ErrorType) String() string {
	if name, ok := baseTypeNames[t]; ok {
		return name
	}
	if name, ok := customTypeNames[t]; ok {
		return name
	}
	return "unknown"
}

// Error represents a typed error with additional context.
// It implements the error interface and supports error wrapping.
type Error struct {
	// Type is the category of the error for machine-readable classification.
	Type ErrorType
	// Code is an optional machine-readable error code (e.g., "entity_not_found").
	Code string
	// Message is the human-readable error message.
	Message string
	// Path is the location in the data (e.g., "blueprint.input.my_input").
	// This is optional and mainly used by validation errors.
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
// It matches if the target is an *Error with the same Type and Code,
// or if the target is the same error instance.
func (e *Error) Is(target error) bool {
	var targetErr *Error
	if errors.As(target, &targetErr) {
		// Match by type and code if both are set
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

// IsWarning returns true if this error should be treated as a warning rather than an error.
// This is determined by the "is_warning" field in Details.
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

// NewWithCode creates a new Error with the specified type, code, and message.
func NewWithCode(errType ErrorType, code, message string) *Error {
	return &Error{
		Type:    errType,
		Code:    code,
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
// Returns ErrorTypeUnknown if the error is not an *Error.
func GetType(err error) ErrorType {
	var typedErr *Error
	if errors.As(err, &typedErr) {
		return typedErr.Type
	}
	return ErrorTypeUnknown
}

// GetCode extracts the error code from an error.
// Returns an empty string if the error is not an *Error or has no code.
func GetCode(err error) string {
	var typedErr *Error
	if errors.As(err, &typedErr) {
		return typedErr.Code
	}
	return ""
}

// GetPath extracts the path from an error.
// Returns an empty string if the error is not an *Error or has no path.
func GetPath(err error) string {
	var typedErr *Error
	if errors.As(err, &typedErr) {
		return typedErr.Path
	}
	return ""
}

// GetDetails extracts the details from an error.
// Returns nil if the error is not an *Error or has no details.
func GetDetails(err error) map[string]any {
	var typedErr *Error
	if errors.As(err, &typedErr) {
		return typedErr.Details
	}
	return nil
}

// IsType checks if an error is of a specific ErrorType.
func IsType(err error, errType ErrorType) bool {
	return GetType(err) == errType
}

// IsValidation checks if an error is a validation error.
func IsValidation(err error) bool {
	return IsType(err, ErrorTypeValidation)
}

// IsParsing checks if an error is a parsing error.
func IsParsing(err error) bool {
	return IsType(err, ErrorTypeParsing)
}

// IsInternal checks if an error is an internal error.
func IsInternal(err error) bool {
	return IsType(err, ErrorTypeInternal)
}
