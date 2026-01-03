package validator

import (
	"fmt"
	"strconv"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/common"
)

// ToFloat converts various types to float64
func ToFloat(val interface{}) (float64, error) {
	switch v := val.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, fmt.Errorf("cannot convert to float")
	}
}

// Abs returns the absolute value of a float64
func Abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// ContainsVariableRef checks if template contains variable references (not !input)
// Delegates to common.ContainsVariableRef for consistency.
func ContainsVariableRef(template string) bool {
	return common.ContainsVariableRef(template)
}

// ContainsInputRef checks if template contains !input references
// Delegates to common.ContainsInputRef for consistency.
func ContainsInputRef(template string) bool {
	return common.ContainsInputRef(template)
}

// CollectInputRefs collects !input references from a string
func (v *BlueprintValidator) CollectInputRefs(value string) {
	if inputName := common.ExtractInputRef(value); inputName != "" {
		v.UsedInputs[inputName] = true
	}
}

// CollectInputRefsFromMap recursively collects !input references from a map
// Uses common.TraverseValue for consistent traversal.
func (v *BlueprintValidator) CollectInputRefsFromMap(m RawData) {
	common.TraverseValue(m, "", func(value interface{}, _ string) bool {
		if s, ok := value.(string); ok {
			if inputName := common.ExtractInputRef(s); inputName != "" {
				v.UsedInputs[inputName] = true
			}
		}
		return true
	})
}

// MergeValidationResult merges validation issues from a common.ValidationResult
// into the BlueprintValidator's error/warning slices.
func (v *BlueprintValidator) MergeValidationResult(result *common.ValidationResult) {
	if result == nil {
		return
	}
	for _, issue := range result.Issues {
		if issue.IsError() {
			v.AddError(issue.String())
		} else {
			v.AddWarning(issue.String())
		}
	}
}

// AddValidationIssue adds a single validation issue to the appropriate slice.
func (v *BlueprintValidator) AddValidationIssue(issue common.ValidationIssue) {
	if issue.IsError() {
		v.AddError(issue.String())
	} else {
		v.AddWarning(issue.String())
	}
}

// AddIssueFromError adds an error string as a validation error if non-empty.
// This is a helper for integrating common validation functions that return error strings.
func (v *BlueprintValidator) AddIssueFromError(errMsg string) {
	if errMsg != "" {
		v.AddError(errMsg)
	}
}

// AddIssueFromWarning adds a warning string as a validation warning if non-empty.
// This is a helper for integrating common validation functions that return warning strings.
func (v *BlueprintValidator) AddIssueFromWarning(warnMsg string) {
	if warnMsg != "" {
		v.AddWarning(warnMsg)
	}
}
