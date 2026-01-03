package validator

import (
	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/common"
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
			v.validateSingleTrigger(triggerMap, common.IndexPath("trigger", i))
		}
	}
}

// validateSingleTrigger validates a single trigger
// Uses common template and field validation utilities.
func (v *BlueprintValidator) validateSingleTrigger(trigger map[string]interface{}, path string) {
	// Check for platform or trigger type
	platform, hasPlatform := trigger["platform"]
	triggerType, hasTrigger := trigger["trigger"]

	if !hasPlatform && !hasTrigger {
		v.AddCategorizedError(CategoryTriggers, path, "Missing 'platform' or 'trigger' key")
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
		if valueTemplate, ok := common.TryGetString(trigger, "value_template"); ok {
			// Template triggers cannot reference automation variables directly
			if common.ContainsVariableRef(valueTemplate) && !common.ContainsInputRef(valueTemplate) {
				v.AddCategorizedWarning(CategoryTriggers, path,
					"Template trigger references variables. Trigger templates are evaluated separately and may not have access to blueprint variables.")
			}
		}
	}

	// Check entity_id is static (no templates in trigger entity_id)
	if entityID, ok := common.TryGetString(trigger, "entity_id"); ok {
		if err := common.ValidateNoTemplateInField(entityID, path, "entity_id"); err != "" {
			v.AddCategorizedError(CategoryTriggers, path, err)
		}
	}

	// Check for 'for' with template (can be problematic)
	if forVal, ok := trigger["for"]; ok {
		if forStr, ok := forVal.(string); ok {
			if common.ContainsTemplate(forStr) {
				// Template in 'for' is valid but may reference unavailable variables
				if common.ContainsVariableRef(forStr) && !common.ContainsInputRef(forStr) {
					v.AddCategorizedWarning(CategoryTriggers, path,
						"'for' template references variables. Variables may not be available in trigger context.")
				}
			}
		}
	}
}
