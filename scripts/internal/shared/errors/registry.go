// Package errors provides a centralized error handling system.
// This file contains the error registry and factory functions for consistent error creation.
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
// If an error with the same code already exists, it will be overwritten.
func (r *Registry) Register(def ErrorDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.definitions[def.Code] = def
}

// Get retrieves an error definition by code.
// Returns nil if not found.
func (r *Registry) Get(code string) *ErrorDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if def, ok := r.definitions[code]; ok {
		return &def
	}
	return nil
}

// Create creates a new Error from a registered definition.
// Returns nil if the code is not registered.
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

// CreateWithMessage creates a new Error from a registered definition with a custom message.
// Returns nil if the code is not registered.
func (r *Registry) CreateWithMessage(code, message string) *Error {
	def := r.Get(code)
	if def == nil {
		return nil
	}
	return &Error{
		Type:    def.Type,
		Code:    def.Code,
		Message: message,
		Details: make(map[string]any),
	}
}

// CreateWithCause creates a new Error from a registered definition wrapping a cause.
// Returns nil if the code is not registered.
func (r *Registry) CreateWithCause(code string, cause error) *Error {
	def := r.Get(code)
	if def == nil {
		return nil
	}
	return &Error{
		Type:    def.Type,
		Code:    def.Code,
		Message: def.Message,
		Cause:   cause,
		Details: make(map[string]any),
	}
}

// CreateWithPath creates a new Error from a registered definition with a path.
// Returns nil if the code is not registered.
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
// Applications can use this or create their own Registry instances.
var DefaultRegistry = NewRegistry()

// Register adds an error definition to the default registry.
func Register(def ErrorDefinition) {
	DefaultRegistry.Register(def)
}

// Create creates a new Error from the default registry.
func Create(code string) *Error {
	return DefaultRegistry.Create(code)
}

// CreateWithMessage creates a new Error from the default registry with a custom message.
func CreateWithMessage(code, message string) *Error {
	return DefaultRegistry.CreateWithMessage(code, message)
}

// CreateWithCause creates a new Error from the default registry wrapping a cause.
func CreateWithCause(code string, cause error) *Error {
	return DefaultRegistry.CreateWithCause(code, cause)
}

// CreateWithPath creates a new Error from the default registry with a path.
func CreateWithPath(code, path string) *Error {
	return DefaultRegistry.CreateWithPath(code, path)
}

// Common error codes that are shared across applications.
const (
	// Validation errors
	CodeMissingArgument = "missing_argument"
	CodeInvalidArgument = "invalid_argument"
	CodeInvalidFormat   = "invalid_format"
	CodeInvalidJSON     = "invalid_json"

	// File errors
	CodeFileNotFound   = "file_not_found"
	CodeFileReadError  = "file_read_error"
	CodeFileParseError = "file_parse_error"
)

// init registers the common error definitions.
func init() {
	// Validation errors
	Register(ErrorDefinition{
		Code:    CodeMissingArgument,
		Type:    ErrorTypeValidation,
		Message: "missing required argument",
	})
	Register(ErrorDefinition{
		Code:    CodeInvalidArgument,
		Type:    ErrorTypeValidation,
		Message: "invalid argument",
	})
	Register(ErrorDefinition{
		Code:    CodeInvalidFormat,
		Type:    ErrorTypeValidation,
		Message: "invalid format",
	})
	Register(ErrorDefinition{
		Code:    CodeInvalidJSON,
		Type:    ErrorTypeParsing,
		Message: "invalid JSON",
	})

	// File errors
	Register(ErrorDefinition{
		Code:    CodeFileNotFound,
		Type:    ErrorTypeParsing,
		Message: "file not found",
	})
	Register(ErrorDefinition{
		Code:    CodeFileReadError,
		Type:    ErrorTypeParsing,
		Message: "failed to read file",
	})
	Register(ErrorDefinition{
		Code:    CodeFileParseError,
		Type:    ErrorTypeParsing,
		Message: "failed to parse file",
	})
}

// Convenience factory functions for common errors.

// ErrMissingArgument creates a missing argument error.
func ErrMissingArgument(usage string) *Error {
	return Create(CodeMissingArgument).WithMessagef("missing argument: %s", usage)
}

// ErrInvalidArgument creates an invalid argument error.
func ErrInvalidArgument(message string) *Error {
	return CreateWithMessage(CodeInvalidArgument, message)
}

// ErrInvalidJSON creates an invalid JSON error.
func ErrInvalidJSON(cause error) *Error {
	return CreateWithCause(CodeInvalidJSON, cause)
}

// ErrFileNotFound creates a file not found error.
func ErrFileNotFound(path string) *Error {
	return Create(CodeFileNotFound).WithPath(path)
}

// ErrFileReadError creates a file read error.
func ErrFileReadError(path string, cause error) *Error {
	return Create(CodeFileReadError).WithPath(path).WithCause(cause)
}
