package validator

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/common"
)

// ValidateHysteresisBoundaries validates hysteresis boundary pairs
func (v *BlueprintValidator) ValidateHysteresisBoundaries() {
	for _, pattern := range HysteresisPatterns {
		for inputName := range v.DefinedInputs {
			matches := pattern.OnPattern.FindStringSubmatch(inputName)
			if matches == nil {
				continue
			}

			// Construct the expected OFF input name
			var offName string
			if pattern.OffPattern == "delta_off" {
				offName = strings.Replace(inputName, "_on", "_off", 1)
			} else {
				offName = pattern.OnPattern.ReplaceAllString(inputName, pattern.OffPattern)
			}

			if !v.DefinedInputs[offName] {
				continue
			}

			// Found a hysteresis pair - validate the relationship
			onDefault := v.InputDefaults[inputName]
			offDefault := v.InputDefaults[offName]

			if onDefault == nil || offDefault == nil {
				continue
			}

			onValue, onErr := ToFloat(onDefault)
			offValue, offErr := ToFloat(offDefault)

			if onErr != nil || offErr != nil {
				continue
			}

			switch {
			case onValue < offValue:
				v.AddErrorf(
					"Hysteresis %s inversion: '%s' (default=%v) should be greater than '%s' (default=%v). With ON < OFF, the system will chatter rapidly.",
					pattern.Description, inputName, onValue, offName, offValue)
			case onValue == offValue:
				v.AddWarningf(
					"Hysteresis %s has no gap: '%s' and '%s' both default to %v. Without a gap between ON and OFF thresholds, there's no hysteresis protection against oscillation.",
					pattern.Description, inputName, offName, onValue)
			default:
				gap := onValue - offValue
				if onValue != 0 && gap/Abs(onValue) < 0.1 {
					v.AddWarningf(
						"Hysteresis %s gap may be too small: '%s' (default=%v) minus '%s' (default=%v) = %v. A larger gap provides better oscillation protection.",
						pattern.Description, inputName, onValue, offName, offValue, gap)
				}
			}
		}
	}
}

// ValidateVariables validates variables section
// Uses common type extraction utilities.
func (v *BlueprintValidator) ValidateVariables() {
	variables, ok := v.Data["variables"]
	if !ok {
		v.AddWarning("No variables section defined")
		return
	}

	variablesMap, ok, errMsg := common.GetMap(variables, "variables")
	if !ok {
		v.AddError(errMsg)
		return
	}

	// Track defined variables
	for name := range variablesMap {
		v.DefinedVariables[name] = true
	}

	// Pre-pass: collect variables with non-zero defaults
	defaultPattern := regexp.MustCompile(`\|\s*(?:float|int)\s*\(\s*(\d+\.?\d*)`)
	for name, value := range variablesMap {
		if valueStr, ok := value.(string); ok {
			if matches := defaultPattern.FindStringSubmatch(valueStr); matches != nil {
				if defaultVal, err := strconv.ParseFloat(matches[1], 64); err == nil && defaultVal > 0 {
					v.NonzeroDefaultVars[name] = true
				}
			}
		}
	}

	// Track join variables and collect input refs
	joinPattern := regexp.MustCompile(`\|\s*join\b|join\s*\(`)
	definedVars := make(map[string]bool)

	for name, value := range variablesMap {
		if valueStr, ok := value.(string); ok {
			if joinPattern.MatchString(valueStr) {
				v.JoinVariables[name] = true
			}

			// Collect !input references
			v.CollectInputRefs(valueStr)

			// Check for bare boolean literals
			v.CheckBareBooleanLiterals(name, valueStr)

			// Check for unsafe math operations
			v.CheckUnsafeMathOperations(name, valueStr)

			// Check for Python-style list methods
			v.CheckPythonStyleMethods(name, valueStr)
		}
		definedVars[name] = true
	}

	// Check for blueprint_version
	if _, ok := variablesMap["blueprint_version"]; !ok {
		v.AddWarning("No 'blueprint_version' variable defined")
	}
}

// CheckBareBooleanLiterals checks for bare boolean literals in templates
// Uses common template detection utilities.
func (v *BlueprintValidator) CheckBareBooleanLiterals(varName, value string) {
	lines := strings.Split(value, "\n")
	foundBareTrue := false
	foundBareFalse := false

	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if stripped == "" {
			continue
		}

		if stripped != "true" && stripped != "false" {
			continue
		}

		// Skip if inside {{ }} blocks using common template detection
		if common.ContainsTemplate(line) {
			continue
		}

		if stripped == "true" && !foundBareTrue {
			foundBareTrue = true
			v.AddWarningf(
				"Variable '%s': Bare 'true' outputs STRING \"true\", not boolean. Use '{{ true }}' to output actual boolean.",
				varName)
		} else if stripped == "false" && !foundBareFalse {
			foundBareFalse = true
			v.AddWarningf(
				"Variable '%s': Bare 'false' outputs STRING \"false\", not boolean. The string \"false\" is TRUTHY (non-empty). Use '{{ false }}' instead.",
				varName)
		}
	}
}

// CheckUnsafeMathOperations checks for potentially unsafe math operations
func (v *BlueprintValidator) CheckUnsafeMathOperations(varName, value string) {
	// Check for log() with potentially non-positive arguments
	logPattern := regexp.MustCompile(`log\s*\(\s*(\w+)\s*\)`)
	for _, match := range logPattern.FindAllStringSubmatch(value, -1) {
		varRef := match[1]
		guardPattern := regexp.MustCompile(fmt.Sprintf(`%s\s*>\s*0|%s\s+is\s+number`, regexp.QuoteMeta(varRef), regexp.QuoteMeta(varRef)))
		if !guardPattern.MatchString(value) {
			v.AddWarningf(
				"Variable '%s': log(%s) may fail if %s <= 0. Consider adding a guard like 'if %s > 0'.",
				varName, varRef, varRef, varRef)
		}
	}

	// Check for sqrt() with potentially negative arguments
	sqrtPattern := regexp.MustCompile(`sqrt\s*\(\s*([^)]+)\)`)
	for _, match := range sqrtPattern.FindAllStringSubmatch(value, -1) {
		arg := strings.TrimSpace(match[1])
		// Skip literal positive numbers
		if regexp.MustCompile(`^\d+\.?\d*$`).MatchString(arg) {
			continue
		}

		// Check if guarded with max(0, x) or abs()
		if !strings.Contains(value, "max(0,") && !strings.Contains(value, "abs(") {
			varMatch := regexp.MustCompile(`([a-zA-Z_]\w*)`).FindStringSubmatch(arg)
			if varMatch != nil {
				v.AddWarningf(
					"Variable '%s': sqrt() with potentially negative argument. Consider using sqrt(max(0, value)).",
					varName)
				break
			}
		}
	}
}

// CheckPythonStyleMethods checks for Python-style list methods
func (v *BlueprintValidator) CheckPythonStyleMethods(varName, value string) {
	// Check for patterns like [a,b].min() which should be [a,b] | min
	pythonMethodPattern := regexp.MustCompile(`\[[^\]]+\]\.(min|max|sum|sort|reverse)\(\)`)
	if pythonMethodPattern.MatchString(value) {
		v.AddErrorf(
			"Variable '%s': Python-style list method detected. Use Jinja2 filter syntax instead (e.g., '[a,b] | min' not '[a,b].min()').",
			varName)
	}
}
