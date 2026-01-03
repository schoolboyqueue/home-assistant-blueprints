package validator

import (
	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/common"
)

// ValidateActions validates action definitions
func (v *BlueprintValidator) ValidateActions() {
	actions, ok := v.Data["action"]
	if !ok {
		return
	}

	v.validateActionList(actions, "action")
}

// validateActionList validates a list of actions
// Uses common path building for consistency.
func (v *BlueprintValidator) validateActionList(actions interface{}, path string) {
	switch a := actions.(type) {
	case []interface{}:
		for i, action := range a {
			v.validateActionList(action, common.IndexPath(path, i))
		}
	case map[string]interface{}:
		v.validateSingleAction(a, path)
	}
}

// validateSingleAction validates a single action
// Uses common service format validation and nil checking.
func (v *BlueprintValidator) validateSingleAction(action map[string]interface{}, path string) {
	// Check for service call
	if service, ok := common.TryGetString(action, "service"); ok {
		// Validate service format using common validator
		if warnMsg := common.ValidateServiceFormat(service, path); warnMsg != "" {
			v.AddCategorizedWarning(CategoryActions, path, warnMsg)
		}

		// Check data is not nil when present using common nil validation
		if data, hasData := action["data"]; hasData {
			if errMsg := common.ValidateNotNil(data, path, "'data' block"); errMsg != "" {
				v.AddCategorizedError(CategoryActions, path, errMsg)
			}
		}
	}

	// Check for choose blocks
	if choose, ok := common.TryGetList(action, "choose"); ok {
		choosePath := common.JoinPath(path, "choose")
		for i, choice := range choose {
			choicePath := common.IndexPath(choosePath, i)
			if choiceMap, ok := choice.(map[string]interface{}); ok {
				if conditions, ok := choiceMap["conditions"]; ok {
					v.validateConditionList(conditions, common.JoinPath(choicePath, "conditions"))
				}
				if sequence, ok := choiceMap["sequence"]; ok {
					v.validateActionList(sequence, common.JoinPath(choicePath, "sequence"))
				}
			}
		}
	}

	// Check for if/then/else
	if _, hasIf := action["if"]; hasIf {
		if thenAction, ok := action["then"]; ok {
			v.validateActionList(thenAction, common.JoinPath(path, "then"))
		} else {
			v.AddCategorizedError(CategoryActions, path, "'if' requires 'then'")
		}
		if elseAction, ok := action["else"]; ok {
			v.validateActionList(elseAction, common.JoinPath(path, "else"))
		}
	}

	// Check for repeat
	if repeat, ok := common.TryGetMap(action, "repeat"); ok {
		repeatPath := common.JoinPath(path, "repeat")
		if sequence, ok := repeat["sequence"]; ok {
			v.validateActionList(sequence, common.JoinPath(repeatPath, "sequence"))
		} else {
			v.AddCategorizedError(CategoryActions, repeatPath, "Missing 'sequence'")
		}
	}

	// Collect input refs from action
	v.CollectInputRefsFromMap(action)
}
