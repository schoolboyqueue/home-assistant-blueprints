// Package errors provides a centralized error handling system with typed errors,
// error factories, and a registry pattern for consistent error creation and handling.
// This package extends the shared errors package with ha-ws-client-go specific error types.
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
// This extends the base ErrorType with ha-ws-client-go specific types.
type ErrorType = errors.ErrorType

// Base error types from shared package
const (
	ErrorTypeUnknown    = errors.ErrorTypeUnknown
	ErrorTypeValidation = errors.ErrorTypeValidation
	ErrorTypeParsing    = errors.ErrorTypeParsing
	ErrorTypeInternal   = errors.ErrorTypeInternal
)

// ha-ws-client-go specific error types
const (
	// ErrorTypeNetwork covers connection, timeout, and network-related errors.
	ErrorTypeNetwork ErrorType = errors.ErrorTypeAppBase + iota
	// ErrorTypeTimeout covers operation timeout errors.
	ErrorTypeTimeout
	// ErrorTypeNotFound covers entity, resource, or item not found errors.
	ErrorTypeNotFound
	// ErrorTypeAuth covers authentication and authorization errors.
	ErrorTypeAuth
	// ErrorTypeAPI covers Home Assistant API errors.
	ErrorTypeAPI
	// ErrorTypeCanceled covers context cancellation errors.
	ErrorTypeCanceled
	// ErrorTypeSubscription covers subscription-related errors.
	ErrorTypeSubscription
)

func init() {
	// Register custom error type names
	errors.RegisterTypeName(ErrorTypeNetwork, "network")
	errors.RegisterTypeName(ErrorTypeTimeout, "timeout")
	errors.RegisterTypeName(ErrorTypeNotFound, "not_found")
	errors.RegisterTypeName(ErrorTypeAuth, "auth")
	errors.RegisterTypeName(ErrorTypeAPI, "api")
	errors.RegisterTypeName(ErrorTypeCanceled, "canceled")
	errors.RegisterTypeName(ErrorTypeSubscription, "subscription")
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
	IsValidation      = errors.IsValidation
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

// ha-ws-client-go specific type check functions

// IsNetwork checks if an error is a network error.
func IsNetwork(err error) bool {
	return errors.IsType(err, ErrorTypeNetwork)
}

// IsTimeout checks if an error is a timeout error.
func IsTimeout(err error) bool {
	return errors.IsType(err, ErrorTypeTimeout)
}

// IsNotFound checks if an error is a not found error.
func IsNotFound(err error) bool {
	return errors.IsType(err, ErrorTypeNotFound)
}

// IsAuth checks if an error is an authentication error.
func IsAuth(err error) bool {
	return errors.IsType(err, ErrorTypeAuth)
}

// IsAPI checks if an error is an API error.
func IsAPI(err error) bool {
	return errors.IsType(err, ErrorTypeAPI)
}

// IsCanceled checks if an error is a cancellation error.
func IsCanceled(err error) bool {
	return errors.IsType(err, ErrorTypeCanceled)
}

// IsSubscription checks if an error is a subscription error.
func IsSubscription(err error) bool {
	return errors.IsType(err, ErrorTypeSubscription)
}
