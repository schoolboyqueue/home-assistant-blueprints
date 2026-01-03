package validator

import (
	"slices"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/common"
)

// ValidateConditions validates condition definitions
func (v *BlueprintValidator) ValidateConditions() {
	conditions, ok := v.Data["condition"]
	if !ok {
		return // Conditions are optional
	}

	v.validateConditionList(conditions, "condition")
}

// validateConditionList validates a list of conditions
// Uses common path building for consistency.
func (v *BlueprintValidator) validateConditionList(conditions interface{}, path string) {
	switch cond := conditions.(type) {
	case []interface{}:
		for i, c := range cond {
			v.validateConditionList(c, common.IndexPath(path, i))
		}
	case map[string]interface{}:
		v.validateSingleCondition(cond, path)
	}
}

// validateSingleCondition validates a single condition
// Uses common enum validation for condition types.
func (v *BlueprintValidator) validateSingleCondition(condition map[string]interface{}, path string) {
	// Check for condition type
	condType, hasCondition := condition["condition"].(string)
	if !hasCondition {
		// Shorthand condition (e.g., just entity_id without condition key)
		if _, hasEntityID := condition["entity_id"]; hasEntityID {
			return // Valid shorthand
		}
		v.AddCategorizedWarning(CategoryConditions, path, "Missing 'condition' key")
		return
	}

	// Validate condition type using common enum validation pattern
	isValid := slices.Contains(ValidConditionTypes, condType)
	if !isValid {
		v.AddCategorizedWarningf(CategoryConditions, path, "Unknown condition type '%s'", condType)
	}

	// Validate nested conditions for and/or/not
	if condType == "and" || condType == "or" || condType == "not" {
		if conditions, ok := condition["conditions"]; ok {
			v.validateConditionList(conditions, common.JoinPath(path, "conditions"))
		}
	}
}
