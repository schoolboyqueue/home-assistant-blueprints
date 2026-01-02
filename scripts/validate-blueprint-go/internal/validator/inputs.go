package validator

import "fmt"

// ValidateInputs validates input definitions
func (v *BlueprintValidator) ValidateInputs() {
	blueprint, ok := v.Data["blueprint"].(map[string]interface{})
	if !ok {
		return
	}

	inputs, ok := blueprint["input"]
	if !ok {
		return
	}

	inputsMap, ok := inputs.(map[string]interface{})
	if !ok {
		v.AddError("'blueprint.input' must be a dictionary")
		return
	}

	v.validateInputDict(inputsMap, "blueprint.input")
}

// validateInputDict recursively validates input definitions
func (v *BlueprintValidator) validateInputDict(inputs map[string]interface{}, path string) {
	for key, value := range inputs {
		currentPath := fmt.Sprintf("%s.%s", path, key)

		valueMap, ok := value.(map[string]interface{})
		if !ok {
			v.AddErrorf("%s: Input must be a dictionary", currentPath)
			continue
		}

		// Check if this is an input group or actual input
		if nestedInput, hasNested := valueMap["input"]; hasNested {
			// This is a group
			nestedMap, ok := nestedInput.(map[string]interface{})
			if !ok {
				v.AddErrorf("%s.input: Must be a dictionary", currentPath)
			} else {
				v.validateInputDict(nestedMap, currentPath)
			}
		} else {
			// This is an actual input definition
			v.DefinedInputs[key] = true
			v.validateSingleInput(valueMap, currentPath, key)
		}
	}
}

// validateSingleInput validates a single input definition
func (v *BlueprintValidator) validateSingleInput(inputDef map[string]interface{}, path, inputName string) {
	// Track default value
	if defaultVal, ok := inputDef["default"]; ok {
		v.InputDefaults[inputName] = defaultVal
	}

	// Check for selector
	selector, ok := inputDef["selector"]
	if !ok {
		v.AddWarningf("%s: No selector defined (inputs should have selectors)", path)
		return
	}

	selectorMap, ok := selector.(map[string]interface{})
	if !ok {
		v.AddErrorf("%s.selector: Must be a dictionary", path)
		return
	}

	// Track selector
	v.InputSelectors[inputName] = selectorMap

	// Validate selector type
	for selectorType := range selectorMap {
		if !ValidSelectorTypes[selectorType] {
			v.AddWarningf("%s.selector: Unknown selector type '%s'", path, selectorType)
		}

		// Track entity selector inputs
		if selectorType == "entity" {
			v.EntityInputs[inputName] = true
			if entitySelector, ok := selectorMap["entity"].(map[string]interface{}); ok {
				if domain, ok := entitySelector["domain"].(string); ok && domain == "input_datetime" {
					v.InputDatetimeInputs[inputName] = true
				}
			}
		}

		// Validate select selector options
		if selectorType == "select" {
			if selectConfig, ok := selectorMap["select"].(map[string]interface{}); ok {
				v.validateSelectOptions(selectConfig, path)
			}
		}
	}
}

// validateSelectOptions validates select selector options
func (v *BlueprintValidator) validateSelectOptions(selectConfig map[string]interface{}, path string) {
	options, ok := selectConfig["options"]
	if !ok {
		return
	}

	optionsList, ok := options.([]interface{})
	if !ok {
		v.AddErrorf("%s.selector.select.options: Must be a list", path)
		return
	}

	for i, option := range optionsList {
		optionPath := fmt.Sprintf("%s.selector.select.options[%d]", path, i)

		switch opt := option.(type) {
		case nil:
			v.AddErrorf("%s: Option cannot be None. Select options must be strings or label/value dicts with non-empty values.", optionPath)
		case map[string]interface{}:
			value := opt["value"]
			label := opt["label"]

			if value == nil {
				v.AddErrorf("%s: Option value is None. Label/value options must have a non-empty 'value' field. Label: '%v'", optionPath, label)
			} else if valueStr, ok := value.(string); !ok {
				v.AddErrorf("%s: Option value must be a string. Label: '%v'", optionPath, label)
			} else if valueStr == "" {
				v.AddErrorf("%s: Option value cannot be empty string. Home Assistant treats empty values as None during import. Label: '%v'", optionPath, label)
			}
		case string:
			if opt == "" {
				v.AddWarningf("%s: Empty string option. Consider using a meaningful value.", optionPath)
			}
		default:
			v.AddErrorf("%s: Option must be a string or label/value dict", optionPath)
		}
	}
}

// ValidateInputReferences validates that all !input references point to defined inputs
func (v *BlueprintValidator) ValidateInputReferences() {
	// Find undefined inputs
	for inputName := range v.UsedInputs {
		if !v.DefinedInputs[inputName] {
			v.AddErrorf(
				"Undefined input reference: '!input %s' - no matching input defined in blueprint.input",
				inputName)
		}
	}
}
