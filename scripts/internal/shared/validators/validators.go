// Package validators provides reusable validation primitives for CLI tools.
// This package contains type-safe extraction utilities, path building helpers,
// and common validation patterns used across validator components.
package validators

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// --- Type Aliases ---
// These type aliases improve code readability and provide a migration path toward
// stronger typing without breaking existing code.

// RawData represents an untyped map structure from YAML/JSON parsing.
// This replaces map[string]interface{} throughout the codebase for better documentation.
type RawData = map[string]interface{}

// AnyList represents an untyped list from YAML/JSON parsing.
// This replaces []interface{} throughout the codebase for better documentation.
type AnyList = []interface{}

// --- Type Extraction Utilities ---
// These functions safely extract typed values from interface{} with descriptive error messages.

// GetMap extracts a map[string]interface{} from an interface{} value.
// Returns the map and true if successful, or nil and false with an error message if not.
func GetMap(value interface{}, path string) (result map[string]interface{}, ok bool, errMsg string) {
	if value == nil {
		return nil, false, fmt.Sprintf("%s: value is nil", path)
	}
	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, false, fmt.Sprintf("%s: must be a dictionary, got %T", path, value)
	}
	return m, true, ""
}

// GetString extracts a string from an interface{} value.
// Returns the string and true if successful, or empty string and false with an error message if not.
func GetString(value interface{}, path string) (result string, ok bool, errMsg string) {
	if value == nil {
		return "", false, fmt.Sprintf("%s: value is nil", path)
	}
	s, ok := value.(string)
	if !ok {
		return "", false, fmt.Sprintf("%s: must be a string, got %T", path, value)
	}
	return s, true, ""
}

// GetList extracts a []interface{} from an interface{} value.
// Returns the list and true if successful, or nil and false with an error message if not.
func GetList(value interface{}, path string) (result []interface{}, ok bool, errMsg string) {
	if value == nil {
		return nil, false, fmt.Sprintf("%s: value is nil", path)
	}
	l, ok := value.([]interface{})
	if !ok {
		return nil, false, fmt.Sprintf("%s: must be a list, got %T", path, value)
	}
	return l, true, ""
}

// GetBool extracts a bool from an interface{} value.
// Returns the bool and true if successful, or false and false with an error message if not.
func GetBool(value interface{}, path string) (result, ok bool, errMsg string) {
	if value == nil {
		return false, false, fmt.Sprintf("%s: value is nil", path)
	}
	b, ok := value.(bool)
	if !ok {
		return false, false, fmt.Sprintf("%s: must be a boolean, got %T", path, value)
	}
	return b, true, ""
}

// GetInt extracts an int from an interface{} value.
// Returns the int and true if successful, or 0 and false with an error message if not.
func GetInt(value interface{}, path string) (result int, ok bool, errMsg string) {
	if value == nil {
		return 0, false, fmt.Sprintf("%s: value is nil", path)
	}
	i, ok := value.(int)
	if !ok {
		return 0, false, fmt.Sprintf("%s: must be an integer, got %T", path, value)
	}
	return i, true, ""
}

// --- Optional Type Extraction ---
// These functions return the typed value if the key exists and is the correct type.
// They do not produce errors for missing keys or type mismatches.

// TryGetMap attempts to extract a map from a parent map at the given key.
// Returns the map and true if found and correct type, or nil and false otherwise.
func TryGetMap(parent map[string]interface{}, key string) (map[string]interface{}, bool) {
	if value, exists := parent[key]; exists {
		if m, ok := value.(map[string]interface{}); ok {
			return m, true
		}
	}
	return nil, false
}

// TryGetString attempts to extract a string from a parent map at the given key.
// Returns the string and true if found and correct type, or empty string and false otherwise.
func TryGetString(parent map[string]interface{}, key string) (string, bool) {
	if value, exists := parent[key]; exists {
		if s, ok := value.(string); ok {
			return s, true
		}
	}
	return "", false
}

// TryGetList attempts to extract a list from a parent map at the given key.
// Returns the list and true if found and correct type, or nil and false otherwise.
func TryGetList(parent map[string]interface{}, key string) ([]interface{}, bool) {
	if value, exists := parent[key]; exists {
		if l, ok := value.([]interface{}); ok {
			return l, true
		}
	}
	return nil, false
}

// TryGetBool attempts to extract a bool from a parent map at the given key.
// Returns the bool and true if found and correct type, or false and false otherwise.
func TryGetBool(parent map[string]interface{}, key string) (result, found bool) {
	if value, exists := parent[key]; exists {
		if b, ok := value.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// TryGetInt attempts to extract an int from a parent map at the given key.
// Returns the int and true if found and correct type, or 0 and false otherwise.
func TryGetInt(parent map[string]interface{}, key string) (int, bool) {
	if value, exists := parent[key]; exists {
		if i, ok := value.(int); ok {
			return i, true
		}
	}
	return 0, false
}

// --- Path Building Utilities ---

// JoinPath creates a path by joining parent and child with a dot separator.
// Handles the case where parent is empty.
func JoinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return fmt.Sprintf("%s.%s", parent, child)
}

// IndexPath creates a path for an array index.
func IndexPath(parent string, index int) string {
	return fmt.Sprintf("%s[%d]", parent, index)
}

// KeyPath creates a path for a map key.
func KeyPath(parent, key string) string {
	return JoinPath(parent, key)
}

// --- Validation Result Types ---

// ValidationSeverity indicates the severity of a validation issue.
type ValidationSeverity int

const (
	SeverityError ValidationSeverity = iota
	SeverityWarning
)

// ValidationIssue represents a single validation error or warning.
type ValidationIssue struct {
	Severity ValidationSeverity
	Path     string
	Message  string
}

// String returns a formatted string representation of the issue.
func (v ValidationIssue) String() string {
	if v.Path != "" {
		return fmt.Sprintf("%s: %s", v.Path, v.Message)
	}
	return v.Message
}

// IsError returns true if this issue is an error.
func (v ValidationIssue) IsError() bool {
	return v.Severity == SeverityError
}

// IsWarning returns true if this issue is a warning.
func (v ValidationIssue) IsWarning() bool {
	return v.Severity == SeverityWarning
}

// ValidationResult collects validation issues from multiple checks.
type ValidationResult struct {
	Issues []ValidationIssue
}

// NewValidationResult creates a new empty validation result.
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Issues: []ValidationIssue{},
	}
}

// AddError adds an error to the validation result.
func (r *ValidationResult) AddError(path, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Severity: SeverityError,
		Path:     path,
		Message:  message,
	})
}

// AddErrorf adds a formatted error to the validation result.
func (r *ValidationResult) AddErrorf(path, format string, args ...interface{}) {
	r.AddError(path, fmt.Sprintf(format, args...))
}

// AddWarning adds a warning to the validation result.
func (r *ValidationResult) AddWarning(path, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Severity: SeverityWarning,
		Path:     path,
		Message:  message,
	})
}

// AddWarningf adds a formatted warning to the validation result.
func (r *ValidationResult) AddWarningf(path, format string, args ...interface{}) {
	r.AddWarning(path, fmt.Sprintf(format, args...))
}

// Merge adds all issues from another ValidationResult.
func (r *ValidationResult) Merge(other *ValidationResult) {
	if other != nil {
		r.Issues = append(r.Issues, other.Issues...)
	}
}

// HasErrors returns true if there are any errors in the result.
func (r *ValidationResult) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.IsError() {
			return true
		}
	}
	return false
}

// HasWarnings returns true if there are any warnings in the result.
func (r *ValidationResult) HasWarnings() bool {
	for _, issue := range r.Issues {
		if issue.IsWarning() {
			return true
		}
	}
	return false
}

// Errors returns only the error issues.
func (r *ValidationResult) Errors() []ValidationIssue {
	var errors []ValidationIssue
	for _, issue := range r.Issues {
		if issue.IsError() {
			errors = append(errors, issue)
		}
	}
	return errors
}

// Warnings returns only the warning issues.
func (r *ValidationResult) Warnings() []ValidationIssue {
	var warnings []ValidationIssue
	for _, issue := range r.Issues {
		if issue.IsWarning() {
			warnings = append(warnings, issue)
		}
	}
	return warnings
}

// ErrorStrings returns error messages as strings.
func (r *ValidationResult) ErrorStrings() []string {
	var result []string
	for _, issue := range r.Errors() {
		result = append(result, issue.String())
	}
	return result
}

// WarningStrings returns warning messages as strings.
func (r *ValidationResult) WarningStrings() []string {
	var result []string
	for _, issue := range r.Warnings() {
		result = append(result, issue.String())
	}
	return result
}

// --- Common Validation Functions ---

// ValidateRequired checks that a key exists in a map.
// Returns an error if the key is missing, otherwise returns an empty string.
func ValidateRequired(m map[string]interface{}, key, path string) string {
	if _, ok := m[key]; !ok {
		return fmt.Sprintf("%s: missing required key '%s'", path, key)
	}
	return ""
}

// ValidateRequiredKeys checks that all specified keys exist in a map.
// Returns a list of error messages for missing keys.
func ValidateRequiredKeys(m map[string]interface{}, keys []string, path string) []string {
	var errors []string
	for _, key := range keys {
		if err := ValidateRequired(m, key, path); err != "" {
			errors = append(errors, err)
		}
	}
	return errors
}

// ValidateEnumValue checks that a string value is one of the allowed values.
// Returns an error message if invalid, or empty string if valid.
func ValidateEnumValue(value string, allowed []string, path, fieldName string) string {
	if !slices.Contains(allowed, value) {
		return fmt.Sprintf("%s: invalid %s '%s', must be one of: %v", path, fieldName, value, allowed)
	}
	return ""
}

// ValidateEnumMap checks that a string value is a key in the allowed map.
// Returns an error message if invalid, or empty string if valid.
func ValidateEnumMap(value string, allowed map[string]bool, path, fieldName string) string {
	if !allowed[value] {
		return fmt.Sprintf("%s: unknown %s '%s'", path, fieldName, value)
	}
	return ""
}

// ValidatePositiveInt checks that an integer value is positive.
// Returns an error message if not positive, or empty string if valid.
func ValidatePositiveInt(value int, path, fieldName string) string {
	if value < 1 {
		return fmt.Sprintf("%s: %s must be a positive integer", path, fieldName)
	}
	return ""
}

// ValidateNotNil checks that a value is not nil.
// Returns an error message if nil, or empty string if valid.
func ValidateNotNil(value interface{}, path, fieldName string) string {
	if value == nil {
		return fmt.Sprintf("%s: %s cannot be None/empty", path, fieldName)
	}
	return ""
}

// ValidateNotEmpty checks that a string is not empty.
// Returns an error message if empty, or empty string if valid.
func ValidateNotEmpty(value, path, fieldName string) string {
	if value == "" {
		return fmt.Sprintf("%s: %s cannot be empty", path, fieldName)
	}
	return ""
}

// --- Template Validation ---

// ContainsTemplate checks if a string contains Jinja2 template markers.
func ContainsTemplate(s string) bool {
	return strings.Contains(s, "{{") || strings.Contains(s, "{%")
}

// ContainsInputRef checks if a string contains !input references.
func ContainsInputRef(s string) bool {
	return strings.Contains(s, "!input")
}

// ContainsVariableRef checks if template contains variable references (not !input).
func ContainsVariableRef(template string) bool {
	varPattern := regexp.MustCompile(`\{\{[^}]*\b[a-z_][a-z0-9_]*\b[^}]*\}\}`)
	return varPattern.MatchString(template)
}

// ValidateBalancedDelimiters checks for balanced Jinja2 delimiters.
// Returns error messages for any unbalanced delimiters.
func ValidateBalancedDelimiters(template, path string) []string {
	var errors []string

	if strings.Count(template, "{{") != strings.Count(template, "}}") {
		errors = append(errors, fmt.Sprintf("%s: unbalanced {{ }} delimiters", path))
	}
	if strings.Count(template, "{%") != strings.Count(template, "%}") {
		errors = append(errors, fmt.Sprintf("%s: unbalanced {%% %%} delimiters", path))
	}

	return errors
}

// ValidateNoInputInTemplate checks that !input is not used inside {{ }} blocks.
// Returns an error message if !input is found inside template blocks, or empty string if valid.
func ValidateNoInputInTemplate(template, path string) string {
	inputInTemplatePattern := regexp.MustCompile(`\{\{[^}]*!input[^}]*\}\}`)
	if inputInTemplatePattern.MatchString(template) {
		return fmt.Sprintf("%s: cannot use !input tags inside {{ }} blocks. Assign the input to a variable first", path)
	}
	return ""
}

// ValidateNoTemplateInField checks that a field does not contain templates.
// Useful for fields like trigger entity_id that must be static.
func ValidateNoTemplateInField(value, path, fieldName string) string {
	if ContainsTemplate(value) {
		return fmt.Sprintf("%s: %s cannot contain templates. Must be a static string or !input reference", path, fieldName)
	}
	return ""
}

// --- Recursive Traversal Utilities ---

// TraversalFunc is a callback function for traversing YAML structures.
// It receives the current value, its path, and should return true to continue traversal.
type TraversalFunc func(value interface{}, path string) bool

// TraverseValue recursively traverses a YAML value structure, calling the visitor
// function for each value. The visitor can return false to stop traversal of children.
func TraverseValue(value interface{}, path string, visitor TraversalFunc) {
	// Call visitor first - if it returns false, don't traverse children
	if !visitor(value, path) {
		return
	}

	switch val := value.(type) {
	case map[string]interface{}:
		for k, v := range val {
			TraverseValue(v, JoinPath(path, k), visitor)
		}
	case []interface{}:
		for i, v := range val {
			TraverseValue(v, IndexPath(path, i), visitor)
		}
	}
}

// CollectStrings traverses a value and collects all string values.
func CollectStrings(value interface{}) []string {
	var collected []string
	TraverseValue(value, "", func(v interface{}, _ string) bool {
		if s, ok := v.(string); ok {
			collected = append(collected, s)
		}
		return true
	})
	return collected
}

// MapVisitorFunc is a callback for visiting map entries specifically.
type MapVisitorFunc func(m map[string]interface{}, path string)

// TraverseMaps recursively finds all map values and calls the visitor for each.
func TraverseMaps(value interface{}, path string, visitor MapVisitorFunc) {
	switch val := value.(type) {
	case map[string]interface{}:
		visitor(val, path)
		for k, v := range val {
			TraverseMaps(v, JoinPath(path, k), visitor)
		}
	case []interface{}:
		for i, v := range val {
			TraverseMaps(v, IndexPath(path, i), visitor)
		}
	}
}

// --- Input Reference Collection ---

// ExtractInputRef extracts !input reference names from a string.
// Returns the input name if found, or empty string if not an input reference.
func ExtractInputRef(value string) string {
	if after, ok := strings.CutPrefix(value, "!input "); ok {
		return strings.TrimSpace(after)
	}
	return ""
}

// CollectInputRefsFromValue recursively collects all !input references from a value.
// Returns a map of input names to true for all found references.
func CollectInputRefsFromValue(value interface{}) map[string]bool {
	refs := make(map[string]bool)

	TraverseValue(value, "", func(v interface{}, _ string) bool {
		if s, ok := v.(string); ok {
			if inputName := ExtractInputRef(s); inputName != "" {
				refs[inputName] = true
			}
		}
		return true
	})

	return refs
}

// --- List/Map Validation Helpers ---

// ValidateListItems validates each item in a list using the provided validator function.
// The validator receives each item, its index, and the path.
func ValidateListItems(list []interface{}, path string, validator func(item interface{}, index int, itemPath string) *ValidationResult) *ValidationResult {
	result := NewValidationResult()
	for i, item := range list {
		itemPath := IndexPath(path, i)
		if itemResult := validator(item, i, itemPath); itemResult != nil {
			result.Merge(itemResult)
		}
	}
	return result
}

// ValidateMapEntries validates each entry in a map using the provided validator function.
// The validator receives each key, value, and the path.
func ValidateMapEntries(m map[string]interface{}, path string, validator func(key string, value interface{}, entryPath string) *ValidationResult) *ValidationResult {
	result := NewValidationResult()
	for key, value := range m {
		entryPath := KeyPath(path, key)
		if entryResult := validator(key, value, entryPath); entryResult != nil {
			result.Merge(entryResult)
		}
	}
	return result
}

// --- Conditional Validation ---

// ValidateIf performs validation only if the condition is true.
// This is useful for conditional validation rules.
func ValidateIf(condition bool, validator func() string) string {
	if condition {
		return validator()
	}
	return ""
}

// ValidateIfPresent validates a value only if the key exists in the map.
func ValidateIfPresent(m map[string]interface{}, key string, validator func(value interface{}) string) string {
	if value, ok := m[key]; ok {
		return validator(value)
	}
	return ""
}

// --- Service Format Validation ---

// ValidateServiceFormat validates that a service string is in domain.service format
// or is a template/input reference.
func ValidateServiceFormat(service, path string) string {
	// Valid if contains a dot, is an input reference, or is a template
	isValidFormat := strings.Contains(service, ".") ||
		strings.HasPrefix(service, "!input") ||
		strings.Contains(service, "{{") ||
		strings.Contains(service, "{%")

	if !isValidFormat {
		return fmt.Sprintf("%s: service '%s' should be in 'domain.service' format", path, service)
	}
	return ""
}

// --- Domain and Selector Validation ---

// ValidateSelector checks if a selector type is valid.
// Returns a warning message if the selector type is unknown, or empty string if valid.
func ValidateSelector(selectorType string, validTypes map[string]bool, path string) string {
	if !validTypes[selectorType] {
		return fmt.Sprintf("%s: unknown selector type '%s'", path, selectorType)
	}
	return ""
}

// ValidateEntityDomain checks if an entity domain matches expected domain(s).
// Returns an error message if not matching, or empty string if valid.
func ValidateEntityDomain(entityDomain string, expectedDomains []string, path string) string {
	if !slices.Contains(expectedDomains, entityDomain) {
		return fmt.Sprintf("%s: unexpected entity domain '%s', expected one of: %v", path, entityDomain, expectedDomains)
	}
	return ""
}
