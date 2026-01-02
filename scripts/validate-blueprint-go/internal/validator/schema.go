package validator

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// ValidateStructure validates root-level structure
func (v *BlueprintValidator) ValidateStructure() {
	// Check required root keys
	for _, key := range RequiredRootKeys {
		if _, ok := v.Data[key]; !ok {
			v.AddErrorf("Missing required root key: '%s'", key)
		}
	}

	// Warn about variables not at root level
	if blueprint, ok := v.Data["blueprint"].(map[string]interface{}); ok {
		if _, hasVars := blueprint["variables"]; hasVars {
			v.AddError("'variables' must be at root level, not nested under 'blueprint'")
		}
	}

	// Check for variables at root
	if variables, ok := v.Data["variables"]; ok && variables != nil {
		if _, isMap := variables.(map[string]interface{}); !isMap {
			v.AddError("'variables' must be a dictionary")
		}
	}
}

// ValidateBlueprintSection validates blueprint metadata section
func (v *BlueprintValidator) ValidateBlueprintSection() {
	blueprint, ok := v.Data["blueprint"]
	if !ok {
		return
	}

	blueprintMap, ok := blueprint.(map[string]interface{})
	if !ok {
		v.AddError("'blueprint' must be a dictionary")
		return
	}

	// Check required blueprint keys
	for _, key := range RequiredBlueprintKeys {
		if _, ok := blueprintMap[key]; !ok {
			v.AddErrorf("Missing required blueprint key: '%s'", key)
		}
	}

	// Validate domain
	if domain, ok := blueprintMap["domain"].(string); ok {
		validDomains := []string{"automation", "script"}
		isValid := slices.Contains(validDomains, domain)
		if !isValid {
			v.AddErrorf("Invalid domain '%s', must be one of: %v", domain, validDomains)
		}
	}
}

// ValidateMode validates automation mode
func (v *BlueprintValidator) ValidateMode() {
	mode, ok := v.Data["mode"]
	if !ok {
		return // Default mode is 'single', which is valid
	}

	modeStr, ok := mode.(string)
	if !ok {
		v.AddError("'mode' must be a string")
		return
	}

	isValid := slices.Contains(ValidModes, modeStr)
	if !isValid {
		v.AddErrorf("Invalid mode '%s', must be one of: %v", modeStr, ValidModes)
	}

	// Check for max when using queued/parallel
	if modeStr == "queued" || modeStr == "parallel" {
		if maxVal, ok := v.Data["max"]; ok {
			if maxInt, ok := maxVal.(int); !ok || maxInt < 1 {
				v.AddErrorf("'max' must be a positive integer when mode is '%s'", modeStr)
			}
		}
	}
}

// ValidateVersionSync validates version sync between name and blueprint_version
func (v *BlueprintValidator) ValidateVersionSync() {
	blueprint, ok := v.Data["blueprint"].(map[string]interface{})
	if !ok {
		return
	}

	name, hasName := blueprint["name"].(string)
	variables, hasVars := v.Data["variables"].(map[string]interface{})

	if !hasName || !hasVars {
		return
	}

	blueprintVersion, hasVersion := variables["blueprint_version"]
	if !hasVersion {
		return
	}

	versionStr := fmt.Sprintf("%v", blueprintVersion)

	// Check if version in name matches blueprint_version
	versionPattern := regexp.MustCompile(`v?(\d+\.\d+(?:\.\d+)?)`)
	nameVersionMatch := versionPattern.FindString(name)

	if nameVersionMatch != "" && !strings.Contains(name, versionStr) {
		v.AddWarningf(
			"Version mismatch: blueprint name contains '%s' but blueprint_version is '%s'",
			nameVersionMatch, versionStr)
	}
}
