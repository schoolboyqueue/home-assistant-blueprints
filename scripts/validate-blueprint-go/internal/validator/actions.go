package validator

import (
	"fmt"
	"strings"
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
func (v *BlueprintValidator) validateActionList(actions interface{}, path string) {
	switch a := actions.(type) {
	case []interface{}:
		for i, action := range a {
			v.validateActionList(action, fmt.Sprintf("%s[%d]", path, i))
		}
	case map[string]interface{}:
		v.validateSingleAction(a, path)
	}
}

// validateSingleAction validates a single action
func (v *BlueprintValidator) validateSingleAction(action map[string]interface{}, path string) {
	// Check for service call
	if service, ok := action["service"].(string); ok {
		// Validate service format (skip templates and !input references)
		isValidFormat := strings.Contains(service, ".") ||
			strings.HasPrefix(service, "!input") ||
			strings.Contains(service, "{{") ||
			strings.Contains(service, "{%")
		if !isValidFormat {
			v.AddWarningf("%s: Service '%s' should be in 'domain.service' format", path, service)
		}

		// Check data is not nil when present
		if data, hasData := action["data"]; hasData && data == nil {
			v.AddErrorf("%s: 'data' block cannot be None/empty", path)
		}
	}

	// Check for choose blocks
	if choose, ok := action["choose"].([]interface{}); ok {
		for i, choice := range choose {
			if choiceMap, ok := choice.(map[string]interface{}); ok {
				if conditions, ok := choiceMap["conditions"]; ok {
					v.validateConditionList(conditions, fmt.Sprintf("%s.choose[%d].conditions", path, i))
				}
				if sequence, ok := choiceMap["sequence"]; ok {
					v.validateActionList(sequence, fmt.Sprintf("%s.choose[%d].sequence", path, i))
				}
			}
		}
	}

	// Check for if/then/else
	if _, hasIf := action["if"]; hasIf {
		if thenAction, ok := action["then"]; ok {
			v.validateActionList(thenAction, fmt.Sprintf("%s.then", path))
		} else {
			v.AddErrorf("%s: 'if' requires 'then'", path)
		}
		if elseAction, ok := action["else"]; ok {
			v.validateActionList(elseAction, fmt.Sprintf("%s.else", path))
		}
	}

	// Check for repeat
	if repeat, ok := action["repeat"].(map[string]interface{}); ok {
		if sequence, ok := repeat["sequence"]; ok {
			v.validateActionList(sequence, fmt.Sprintf("%s.repeat.sequence", path))
		} else {
			v.AddErrorf("%s.repeat: Missing 'sequence'", path)
		}
	}

	// Collect input refs from action
	v.CollectInputRefsFromMap(action)
}
