package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/common"
)

// ValidateStructure validates root-level structure
// Uses common required key validation patterns.
func (v *BlueprintValidator) ValidateStructure() {
	// Check required root keys
	for _, key := range RequiredRootKeys {
		if _, ok := v.Data[key]; !ok {
			v.AddErrorf("Missing required root key: '%s'", key)
		}
	}

	// Warn about variables not at root level
	if blueprint, ok := common.TryGetMap(v.Data, "blueprint"); ok {
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
// Uses common type extraction and enum validation.
func (v *BlueprintValidator) ValidateBlueprintSection() {
	blueprint, ok := v.Data["blueprint"]
	if !ok {
		return
	}

	blueprintMap, ok, errMsg := common.GetMap(blueprint, "blueprint")
	if !ok {
		v.AddError(errMsg)
		return
	}

	// Check required blueprint keys
	for _, key := range RequiredBlueprintKeys {
		if _, ok := blueprintMap[key]; !ok {
			v.AddErrorf("Missing required blueprint key: '%s'", key)
		}
	}

	// Validate domain using common enum validation
	if domain, ok := common.TryGetString(blueprintMap, "domain"); ok {
		validDomains := []string{"automation", "script"}
		if errMsg := common.ValidateEnumValue(domain, validDomains, "blueprint", "domain"); errMsg != "" {
			v.AddError(errMsg)
		}
	}
}

// ValidateMode validates automation mode
// Uses common type extraction and enum validation.
func (v *BlueprintValidator) ValidateMode() {
	mode, ok := v.Data["mode"]
	if !ok {
		return // Default mode is 'single', which is valid
	}

	modeStr, ok, errMsg := common.GetString(mode, "mode")
	if !ok {
		v.AddError(errMsg)
		return
	}

	// Validate mode using common enum validation
	if errMsg := common.ValidateEnumValue(modeStr, ValidModes, "", "mode"); errMsg != "" {
		v.AddError(errMsg)
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
// Uses common type extraction utilities.
func (v *BlueprintValidator) ValidateVersionSync() {
	blueprint, ok := common.TryGetMap(v.Data, "blueprint")
	if !ok {
		return
	}

	name, hasName := common.TryGetString(blueprint, "name")
	variables, hasVars := common.TryGetMap(v.Data, "variables")

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
