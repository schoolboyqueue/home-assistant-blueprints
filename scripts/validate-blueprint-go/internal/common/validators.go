// Package common provides reusable validation primitives for Home Assistant Blueprint validation.
// This package re-exports the shared validators package and adds any validate-blueprint-go specific utilities.
package common

import (
	"github.com/home-assistant-blueprints/shared/validators"
)

// Re-export type aliases from shared validators
type (
	// RawData represents an untyped map structure from YAML parsing.
	RawData = validators.RawData
	// AnyList represents an untyped list from YAML parsing.
	AnyList = validators.AnyList
	// ValidationSeverity indicates the severity of a validation issue.
	ValidationSeverity = validators.ValidationSeverity
	// ValidationIssue represents a single validation error or warning.
	ValidationIssue = validators.ValidationIssue
	// ValidationResult collects validation issues from multiple checks.
	ValidationResult = validators.ValidationResult
	// TraversalFunc is a callback function for traversing YAML structures.
	TraversalFunc = validators.TraversalFunc
	// MapVisitorFunc is a callback for visiting map entries specifically.
	MapVisitorFunc = validators.MapVisitorFunc
)

// Re-export constants
const (
	SeverityError   = validators.SeverityError
	SeverityWarning = validators.SeverityWarning
)

// Re-export type extraction functions
var (
	GetMap    = validators.GetMap
	GetString = validators.GetString
	GetList   = validators.GetList
	GetBool   = validators.GetBool
	GetInt    = validators.GetInt
)

// Re-export optional type extraction functions
var (
	TryGetMap    = validators.TryGetMap
	TryGetString = validators.TryGetString
	TryGetList   = validators.TryGetList
	TryGetBool   = validators.TryGetBool
	TryGetInt    = validators.TryGetInt
)

// Re-export path building utilities
var (
	JoinPath  = validators.JoinPath
	IndexPath = validators.IndexPath
	KeyPath   = validators.KeyPath
)

// Re-export validation result functions
var (
	NewValidationResult = validators.NewValidationResult
)

// Re-export common validation functions
var (
	ValidateRequired     = validators.ValidateRequired
	ValidateRequiredKeys = validators.ValidateRequiredKeys
	ValidateEnumValue    = validators.ValidateEnumValue
	ValidateEnumMap      = validators.ValidateEnumMap
	ValidatePositiveInt  = validators.ValidatePositiveInt
	ValidateNotNil       = validators.ValidateNotNil
	ValidateNotEmpty     = validators.ValidateNotEmpty
)

// Re-export template validation functions
var (
	ContainsTemplate           = validators.ContainsTemplate
	ContainsInputRef           = validators.ContainsInputRef
	ContainsVariableRef        = validators.ContainsVariableRef
	ValidateBalancedDelimiters = validators.ValidateBalancedDelimiters
	ValidateNoInputInTemplate  = validators.ValidateNoInputInTemplate
	ValidateNoTemplateInField  = validators.ValidateNoTemplateInField
)

// Re-export traversal utilities
var (
	TraverseValue  = validators.TraverseValue
	CollectStrings = validators.CollectStrings
	TraverseMaps   = validators.TraverseMaps
)

// Re-export input reference collection
var (
	ExtractInputRef           = validators.ExtractInputRef
	CollectInputRefsFromValue = validators.CollectInputRefsFromValue
)

// Re-export list/map validation helpers
var (
	ValidateListItems  = validators.ValidateListItems
	ValidateMapEntries = validators.ValidateMapEntries
)

// Re-export conditional validation
var (
	ValidateIf        = validators.ValidateIf
	ValidateIfPresent = validators.ValidateIfPresent
)

// Re-export service/selector validation
var (
	ValidateServiceFormat = validators.ValidateServiceFormat
	ValidateSelector      = validators.ValidateSelector
	ValidateEntityDomain  = validators.ValidateEntityDomain
)
