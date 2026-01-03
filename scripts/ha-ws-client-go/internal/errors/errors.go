// Package errors provides a centralized error handling system with typed errors,
// error factories, and a registry pattern for consistent error creation and handling.
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
	// ErrorTypeNetwork covers connection, timeout, and network-related errors.
	ErrorTypeNetwork
	// ErrorTypeValidation covers input validation and argument errors.
	ErrorTypeValidation
	// ErrorTypeParsing covers JSON, YAML, and other parsing errors.
	ErrorTypeParsing
	// ErrorTypeTimeout covers operation timeout errors.
	ErrorTypeTimeout
	// ErrorTypeNotFound covers entity, resource, or item not found errors.
	ErrorTypeNotFound
	// ErrorTypeAuth covers authentication and authorization errors.
	ErrorTypeAuth
	// ErrorTypeAPI covers Home Assistant API errors.
	ErrorTypeAPI
	// ErrorTypeInternal covers internal/unexpected errors.
	ErrorTypeInternal
	// ErrorTypeCanceled covers context cancellation errors.
	ErrorTypeCanceled
	// ErrorTypeSubscription covers subscription-related errors.
	ErrorTypeSubscription
)

// errorTypeNames maps error types to their string representations.
var errorTypeNames = map[ErrorType]string{
	ErrorTypeUnknown:      "unknown",
	ErrorTypeNetwork:      "network",
	ErrorTypeValidation:   "validation",
	ErrorTypeParsing:      "parsing",
	ErrorTypeTimeout:      "timeout",
	ErrorTypeNotFound:     "not_found",
	ErrorTypeAuth:         "auth",
	ErrorTypeAPI:          "api",
	ErrorTypeInternal:     "internal",
	ErrorTypeCanceled:     "canceled",
	ErrorTypeSubscription: "subscription",
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
	// Code is an optional machine-readable error code (e.g., "entity_not_found").
	Code string
	// Message is the human-readable error message.
	Message string
	// Cause is the underlying error, if any.
	Cause error
	// Details contains additional context about the error.
	Details map[string]any
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Code != "" {
		if e.Cause != nil {
			return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
		}
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
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

// IsNetwork checks if an error is a network error.
func IsNetwork(err error) bool {
	return IsType(err, ErrorTypeNetwork)
}

// IsValidation checks if an error is a validation error.
func IsValidation(err error) bool {
	return IsType(err, ErrorTypeValidation)
}

// IsParsing checks if an error is a parsing error.
func IsParsing(err error) bool {
	return IsType(err, ErrorTypeParsing)
}

// IsTimeout checks if an error is a timeout error.
func IsTimeout(err error) bool {
	return IsType(err, ErrorTypeTimeout)
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	return IsType(err, ErrorTypeNotFound)
}

// IsAuth checks if an error is an authentication error.
func IsAuth(err error) bool {
	return IsType(err, ErrorTypeAuth)
}

// IsAPI checks if an error is an API error.
func IsAPI(err error) bool {
	return IsType(err, ErrorTypeAPI)
}

// IsInternal checks if an error is an internal error.
func IsInternal(err error) bool {
	return IsType(err, ErrorTypeInternal)
}

// IsCanceled checks if an error is a cancellation error.
func IsCanceled(err error) bool {
	return IsType(err, ErrorTypeCanceled)
}

// IsSubscription checks if an error is a subscription error.
func IsSubscription(err error) bool {
	return IsType(err, ErrorTypeSubscription)
}
