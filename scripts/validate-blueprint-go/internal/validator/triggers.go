package validator

import (
	"fmt"
	"strings"
)

// ValidateTriggers validates trigger definitions
func (v *BlueprintValidator) ValidateTriggers() {
	triggers, ok := v.Data["trigger"]
	if !ok {
		return
	}

	triggerList, ok := triggers.([]interface{})
	if !ok {
		// Single trigger (not a list)
		if triggerMap, ok := triggers.(map[string]interface{}); ok {
			v.validateSingleTrigger(triggerMap, "trigger")
		}
		return
	}

	for i, trigger := range triggerList {
		if triggerMap, ok := trigger.(map[string]interface{}); ok {
			v.validateSingleTrigger(triggerMap, fmt.Sprintf("trigger[%d]", i))
		}
	}
}

// validateSingleTrigger validates a single trigger
func (v *BlueprintValidator) validateSingleTrigger(trigger map[string]interface{}, path string) {
	// Check for platform or trigger type
	platform, hasPlatform := trigger["platform"]
	triggerType, hasTrigger := trigger["trigger"]

	if !hasPlatform && !hasTrigger {
		v.AddErrorf("%s: Missing 'platform' or 'trigger' key", path)
		return
	}

	var platformStr string
	if hasPlatform {
		if str, ok := platform.(string); ok {
			platformStr = str
		}
	} else if hasTrigger {
		if str, ok := triggerType.(string); ok {
			platformStr = str
		}
	}

	// Check for template triggers using variables
	if platformStr == "template" {
		if valueTemplate, ok := trigger["value_template"].(string); ok {
			// Template triggers cannot reference automation variables directly
			if ContainsVariableRef(valueTemplate) && !ContainsInputRef(valueTemplate) {
				v.AddWarningf(
					"%s: Template trigger references variables. Trigger templates are evaluated separately and may not have access to blueprint variables.",
					path)
			}
		}
	}

	// Check entity_id is static (no templates in trigger entity_id)
	if entityID, ok := trigger["entity_id"].(string); ok {
		if strings.Contains(entityID, "{{") || strings.Contains(entityID, "{%") {
			v.AddErrorf(
				"%s: entity_id cannot contain templates. Trigger entity_id must be a static string or !input reference.",
				path)
		}
	}

	// Check for 'for' with template (can be problematic)
	if forVal, ok := trigger["for"]; ok {
		if forStr, ok := forVal.(string); ok {
			if strings.Contains(forStr, "{{") {
				// Template in 'for' is valid but may reference unavailable variables
				if ContainsVariableRef(forStr) && !ContainsInputRef(forStr) {
					v.AddWarningf(
						"%s: 'for' template references variables. Variables may not be available in trigger context.",
						path)
				}
			}
		}
	}
}
