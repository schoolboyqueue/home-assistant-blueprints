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

// CreateWithMessage creates a new Error from the default registry with a custom message.
func CreateWithMessage(code, message string) *Error {
	return DefaultRegistry.CreateWithMessage(code, message)
}

// CreateWithCause creates a new Error from the default registry wrapping a cause.
func CreateWithCause(code string, cause error) *Error {
	return DefaultRegistry.CreateWithCause(code, cause)
}

// Common error codes for the ha-ws-client-go package.
const (
	// Connection errors
	CodeConnectionClosed    = "connection_closed"
	CodeConnectionFailed    = "connection_failed"
	CodeConnectionTimeout   = "connection_timeout"
	CodeAuthFailed          = "auth_failed"
	CodeAuthRequired        = "auth_required"
	CodeWebsocketError      = "websocket_error"
	CodeMessageSendFailed   = "message_send_failed"
	CodeMessageReadFailed   = "message_read_failed"
	CodeMessageMarshalError = "message_marshal_error"

	// Entity errors
	CodeEntityNotFound    = "entity_not_found"
	CodeEntityInvalidID   = "entity_invalid_id"
	CodeEntityStateError  = "entity_state_error"
	CodeEntityUnavailable = "entity_unavailable"

	// Subscription errors
	CodeSubscriptionFailed  = "subscription_failed"
	CodeSubscriptionTimeout = "subscription_timeout"
	CodeUnsubscribeFailed   = "unsubscribe_failed"

	// Request errors
	CodeRequestCanceled = "request_canceled"
	CodeRequestTimeout  = "request_timeout"
	CodeRequestFailed   = "request_failed"

	// Validation errors
	CodeMissingArgument = "missing_argument"
	CodeInvalidArgument = "invalid_argument"
	CodeInvalidPattern  = "invalid_pattern"
	CodeInvalidJSON     = "invalid_json"
	CodeInvalidFormat   = "invalid_format"

	// API errors
	CodeAPIError         = "api_error"
	CodeServiceCallError = "service_call_error"
	CodeNoDataFound      = "no_data_found"
	CodeNoTracesFound    = "no_traces_found"
	CodeTemplateError    = "template_error"
)

// init registers the default error definitions.
func init() {
	// Connection errors
	Register(ErrorDefinition{
		Code:    CodeConnectionClosed,
		Type:    ErrorTypeNetwork,
		Message: "connection closed",
	})
	Register(ErrorDefinition{
		Code:    CodeConnectionFailed,
		Type:    ErrorTypeNetwork,
		Message: "failed to connect",
	})
	Register(ErrorDefinition{
		Code:    CodeConnectionTimeout,
		Type:    ErrorTypeTimeout,
		Message: "connection timed out",
	})
	Register(ErrorDefinition{
		Code:    CodeAuthFailed,
		Type:    ErrorTypeAuth,
		Message: "authentication failed",
	})
	Register(ErrorDefinition{
		Code:    CodeAuthRequired,
		Type:    ErrorTypeAuth,
		Message: "authentication required",
	})
	Register(ErrorDefinition{
		Code:    CodeWebsocketError,
		Type:    ErrorTypeNetwork,
		Message: "websocket error",
	})
	Register(ErrorDefinition{
		Code:    CodeMessageSendFailed,
		Type:    ErrorTypeNetwork,
		Message: "failed to send message",
	})
	Register(ErrorDefinition{
		Code:    CodeMessageReadFailed,
		Type:    ErrorTypeNetwork,
		Message: "failed to read message",
	})
	Register(ErrorDefinition{
		Code:    CodeMessageMarshalError,
		Type:    ErrorTypeParsing,
		Message: "failed to marshal message",
	})

	// Entity errors
	Register(ErrorDefinition{
		Code:    CodeEntityNotFound,
		Type:    ErrorTypeNotFound,
		Message: "entity not found",
	})
	Register(ErrorDefinition{
		Code:    CodeEntityInvalidID,
		Type:    ErrorTypeValidation,
		Message: "invalid entity ID format",
	})
	Register(ErrorDefinition{
		Code:    CodeEntityStateError,
		Type:    ErrorTypeAPI,
		Message: "failed to get entity state",
	})
	Register(ErrorDefinition{
		Code:    CodeEntityUnavailable,
		Type:    ErrorTypeAPI,
		Message: "entity unavailable",
	})

	// Subscription errors
	Register(ErrorDefinition{
		Code:    CodeSubscriptionFailed,
		Type:    ErrorTypeSubscription,
		Message: "subscription failed",
	})
	Register(ErrorDefinition{
		Code:    CodeSubscriptionTimeout,
		Type:    ErrorTypeTimeout,
		Message: "subscription timeout",
	})
	Register(ErrorDefinition{
		Code:    CodeUnsubscribeFailed,
		Type:    ErrorTypeSubscription,
		Message: "failed to unsubscribe",
	})

	// Request errors
	Register(ErrorDefinition{
		Code:    CodeRequestCanceled,
		Type:    ErrorTypeCanceled,
		Message: "request canceled",
	})
	Register(ErrorDefinition{
		Code:    CodeRequestTimeout,
		Type:    ErrorTypeTimeout,
		Message: "request timed out",
	})
	Register(ErrorDefinition{
		Code:    CodeRequestFailed,
		Type:    ErrorTypeAPI,
		Message: "request failed",
	})

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
		Code:    CodeInvalidPattern,
		Type:    ErrorTypeValidation,
		Message: "invalid pattern",
	})
	Register(ErrorDefinition{
		Code:    CodeInvalidJSON,
		Type:    ErrorTypeParsing,
		Message: "invalid JSON",
	})
	Register(ErrorDefinition{
		Code:    CodeInvalidFormat,
		Type:    ErrorTypeValidation,
		Message: "invalid format",
	})

	// API errors
	Register(ErrorDefinition{
		Code:    CodeAPIError,
		Type:    ErrorTypeAPI,
		Message: "API error",
	})
	Register(ErrorDefinition{
		Code:    CodeServiceCallError,
		Type:    ErrorTypeAPI,
		Message: "service call failed",
	})
	Register(ErrorDefinition{
		Code:    CodeNoDataFound,
		Type:    ErrorTypeNotFound,
		Message: "no data found",
	})
	Register(ErrorDefinition{
		Code:    CodeNoTracesFound,
		Type:    ErrorTypeNotFound,
		Message: "no traces found",
	})
	Register(ErrorDefinition{
		Code:    CodeTemplateError,
		Type:    ErrorTypeAPI,
		Message: "template error",
	})
}

// Convenience factory functions for common errors.

// ErrConnectionClosed creates a connection closed error.
func ErrConnectionClosed() *Error {
	return Create(CodeConnectionClosed)
}

// ErrConnectionFailed creates a connection failed error.
func ErrConnectionFailed(cause error) *Error {
	return CreateWithCause(CodeConnectionFailed, cause)
}

// ErrAuthFailed creates an authentication failed error.
func ErrAuthFailed(message string) *Error {
	return CreateWithMessage(CodeAuthFailed, message)
}

// ErrEntityNotFound creates an entity not found error.
func ErrEntityNotFound(entityID string) *Error {
	return Create(CodeEntityNotFound).WithMessagef("entity not found: %s", entityID).WithDetails(map[string]any{
		"entity_id": entityID,
	})
}

// ErrEntityInvalidID creates an invalid entity ID error.
func ErrEntityInvalidID(entityID string) *Error {
	return Create(CodeEntityInvalidID).WithMessagef("invalid entity ID format: %s", entityID).WithDetails(map[string]any{
		"entity_id": entityID,
	})
}

// ErrSubscriptionFailed creates a subscription failed error.
func ErrSubscriptionFailed(cause error) *Error {
	if cause != nil {
		return CreateWithCause(CodeSubscriptionFailed, cause)
	}
	return Create(CodeSubscriptionFailed)
}

// ErrSubscriptionTimeout creates a subscription timeout error.
func ErrSubscriptionTimeout() *Error {
	return Create(CodeSubscriptionTimeout)
}

// ErrRequestCanceled creates a request canceled error.
func ErrRequestCanceled(cause error) *Error {
	return CreateWithCause(CodeRequestCanceled, cause)
}

// ErrRequestTimeout creates a request timeout error.
func ErrRequestTimeout() *Error {
	return Create(CodeRequestTimeout)
}

// ErrMissingArgument creates a missing argument error.
func ErrMissingArgument(usage string) *Error {
	return Create(CodeMissingArgument).WithMessagef("missing argument: %s", usage)
}

// ErrInvalidArgument creates an invalid argument error.
func ErrInvalidArgument(message string) *Error {
	return CreateWithMessage(CodeInvalidArgument, message)
}

// ErrInvalidPattern creates an invalid pattern error.
func ErrInvalidPattern(cause error) *Error {
	return CreateWithCause(CodeInvalidPattern, cause)
}

// ErrInvalidJSON creates an invalid JSON error.
func ErrInvalidJSON(cause error) *Error {
	return CreateWithCause(CodeInvalidJSON, cause)
}

// ErrMessageMarshalFailed creates a message marshal error.
func ErrMessageMarshalFailed(cause error) *Error {
	return CreateWithCause(CodeMessageMarshalError, cause)
}

// ErrMessageSendFailed creates a message send error.
func ErrMessageSendFailed(cause error) *Error {
	return CreateWithCause(CodeMessageSendFailed, cause)
}

// ErrAPIError creates an API error with the given code and message.
func ErrAPIError(code, message string) *Error {
	return &Error{
		Type:    ErrorTypeAPI,
		Code:    code,
		Message: message,
		Details: make(map[string]any),
	}
}

// ErrNoTracesFound creates a no traces found error.
func ErrNoTracesFound(automationID string) *Error {
	return Create(CodeNoTracesFound).WithMessagef("no traces found for automation.%s", automationID).WithDetails(map[string]any{
		"automation_id": automationID,
	})
}

// ErrTemplateTimeout creates a template render timeout error.
func ErrTemplateTimeout() *Error {
	return Create(CodeTemplateError).WithMessage("template render timeout")
}

// ErrNoDataFound creates a no data found error.
func ErrNoDataFound(message string) *Error {
	return CreateWithMessage(CodeNoDataFound, message)
}
