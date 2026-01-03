package validator

import (
	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/common"
)

// ValidateInputs validates input definitions
func (v *BlueprintValidator) ValidateInputs() {
	blueprint, ok := common.TryGetMap(v.Data, "blueprint")
	if !ok {
		return
	}

	inputs, hasInput := blueprint["input"]
	if !hasInput {
		return
	}

	inputsMap, ok, errMsg := common.GetMap(inputs, "blueprint.input")
	if !ok {
		v.AddCategorizedError(CategoryInputs, "blueprint.input", errMsg)
		return
	}

	v.validateInputDict(inputsMap, "blueprint.input")
}

// validateInputDict recursively validates input definitions
// Uses common path building utilities for consistency.
func (v *BlueprintValidator) validateInputDict(inputs map[string]interface{}, path string) {
	for key, value := range inputs {
		currentPath := common.KeyPath(path, key)

		valueMap, ok, errMsg := common.GetMap(value, currentPath)
		if !ok {
			v.AddCategorizedError(CategoryInputs, currentPath, errMsg)
			continue
		}

		// Check if this is an input group or actual input
		if nestedInput, hasNested := valueMap["input"]; hasNested {
			// This is a group
			nestedMap, ok, errMsg := common.GetMap(nestedInput, common.JoinPath(currentPath, "input"))
			if !ok {
				v.AddCategorizedError(CategoryInputs, common.JoinPath(currentPath, "input"), errMsg)
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
// Uses common type extraction and validation patterns.
func (v *BlueprintValidator) validateSingleInput(inputDef map[string]interface{}, path, inputName string) {
	// Track default value
	if defaultVal, ok := inputDef["default"]; ok {
		v.InputDefaults[inputName] = defaultVal
	}

	// Check for selector
	selector, ok := inputDef["selector"]
	if !ok {
		v.AddCategorizedWarningf(CategoryInputs, path, "No selector defined (inputs should have selectors)")
		return
	}

	selectorPath := common.JoinPath(path, "selector")
	selectorMap, ok, errMsg := common.GetMap(selector, selectorPath)
	if !ok {
		v.AddCategorizedError(CategoryInputs, selectorPath, errMsg)
		return
	}

	// Track selector
	v.InputSelectors[inputName] = selectorMap

	// Validate selector type
	for selectorType := range selectorMap {
		// Use common selector validation
		if warnMsg := common.ValidateSelector(selectorType, ValidSelectorTypes, selectorPath); warnMsg != "" {
			v.AddCategorizedWarning(CategoryInputs, selectorPath, warnMsg)
		}

		// Track entity selector inputs
		if selectorType == "entity" {
			v.EntityInputs[inputName] = true
			if entitySelector, ok := common.TryGetMap(selectorMap, "entity"); ok {
				if domain, ok := common.TryGetString(entitySelector, "domain"); ok && domain == "input_datetime" {
					v.InputDatetimeInputs[inputName] = true
				}
			}
		}

		// Validate select selector options
		if selectorType == "select" {
			if selectConfig, ok := common.TryGetMap(selectorMap, "select"); ok {
				v.validateSelectOptions(selectConfig, path)
			}
		}
	}
}

// validateSelectOptions validates select selector options
// Uses common path building and type checking utilities.
func (v *BlueprintValidator) validateSelectOptions(selectConfig map[string]interface{}, path string) {
	options, ok := selectConfig["options"]
	if !ok {
		return
	}

	optionsPath := common.JoinPath(path, "selector.select.options")
	optionsList, ok, errMsg := common.GetList(options, optionsPath)
	if !ok {
		v.AddCategorizedError(CategoryInputs, optionsPath, errMsg)
		return
	}

	for i, option := range optionsList {
		optionPath := common.IndexPath(optionsPath, i)

		switch opt := option.(type) {
		case nil:
			v.AddCategorizedError(CategoryInputs, optionPath, "Option cannot be None. Select options must be strings or label/value dicts with non-empty values.")
		case map[string]interface{}:
			value := opt["value"]
			label := opt["label"]

			if value == nil {
				v.AddCategorizedErrorf(CategoryInputs, optionPath, "Option value is None. Label/value options must have a non-empty 'value' field. Label: '%v'", label)
			} else if valueStr, ok := value.(string); !ok {
				v.AddCategorizedErrorf(CategoryInputs, optionPath, "Option value must be a string. Label: '%v'", label)
			} else if valueStr == "" {
				v.AddCategorizedErrorf(CategoryInputs, optionPath, "Option value cannot be empty string. Home Assistant treats empty values as None during import. Label: '%v'", label)
			}
		case string:
			if opt == "" {
				v.AddCategorizedWarning(CategoryInputs, optionPath, "Empty string option. Consider using a meaningful value.")
			}
		default:
			v.AddCategorizedError(CategoryInputs, optionPath, "Option must be a string or label/value dict")
		}
	}
}

// ValidateInputReferences validates that all !input references point to defined inputs
func (v *BlueprintValidator) ValidateInputReferences() {
	// Find undefined inputs
	for inputName := range v.UsedInputs {
		if !v.DefinedInputs[inputName] {
			v.AddCategorizedErrorf(CategoryReferences, "",
				"Undefined input reference: '!input %s' - no matching input defined in blueprint.input",
				inputName)
		}
	}
}
