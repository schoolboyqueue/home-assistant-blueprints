// validate-blueprint - A comprehensive Home Assistant Blueprint validator
//
// This tool performs comprehensive validation of Home Assistant blueprint files:
// 1. YAML syntax validation
// 2. Blueprint schema validation (required keys, structure)
// 3. Input/selector validation
// 4. Template syntax checking
// 5. Service call structure validation
// 6. Version sync validation
// 7. Trigger validation
// 8. Condition validation
// 9. Mode validation
// 10. Input reference validation
// And more...
//
// Usage:
//
//	validate-blueprint <blueprint.yaml>
//	validate-blueprint --all
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/yaml.v3"
)

// Version information - set by ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// ValidModes are the valid automation modes
var ValidModes = []string{"single", "restart", "queued", "parallel"}

// ValidConditionTypes are the valid condition types
var ValidConditionTypes = []string{
	"and", "or", "not", "state", "numeric_state", "template",
	"time", "zone", "trigger", "sun", "device",
}

// ValidSelectorTypes are the valid input selector types
var ValidSelectorTypes = map[string]bool{
	"action": true, "addon": true, "area": true, "attribute": true,
	"boolean": true, "color_rgb": true, "color_temp": true, "condition": true,
	"conversation_agent": true, "country": true, "date": true, "datetime": true,
	"device": true, "duration": true, "entity": true, "file": true,
	"floor": true, "icon": true, "label": true, "language": true,
	"location": true, "media": true, "navigation": true, "number": true,
	"object": true, "select": true, "state": true, "target": true,
	"template": true, "text": true, "theme": true, "time": true,
	"trigger": true, "ui_action": true, "ui_color": true,
}

// RequiredBlueprintKeys are required in the blueprint section
var RequiredBlueprintKeys = []string{"name", "description", "domain", "input"}

// RequiredRootKeys are required at the root level
var RequiredRootKeys = []string{"blueprint", "trigger", "action"}

// HysteresisPattern defines a pattern pair for hysteresis validation
type HysteresisPattern struct {
	OnPattern   *regexp.Regexp
	OffPattern  string
	Description string
}

// HysteresisPatterns for detecting hysteresis configuration issues
var HysteresisPatterns = []HysteresisPattern{
	{regexp.MustCompile(`(.*)_on$`), "${1}_off", "threshold"},
	{regexp.MustCompile(`(.*)_high$`), "${1}_low", "boundary"},
	{regexp.MustCompile(`(.*)_upper$`), "${1}_lower", "limit"},
	{regexp.MustCompile(`(.*)_start$`), "${1}_stop", "trigger point"},
	{regexp.MustCompile(`(.*)_enable$`), "${1}_disable", "activation point"},
	{regexp.MustCompile(`delta_on$`), "delta_off", "delta threshold"},
}

// Jinja2Builtins are built-in Jinja2/HA template functions that shouldn't trigger undefined warnings
var Jinja2Builtins = map[string]bool{
	// Python/Jinja2 built-in constants
	"true": true, "false": true, "none": true,
	"True": true, "False": true, "None": true,
	// Jinja2 control keywords
	"if": true, "else": true, "elif": true, "endif": true,
	"for": true, "endfor": true, "in": true, "not": true,
	"and": true, "or": true, "is": true, "set": true, "endset": true,
	"macro": true, "endmacro": true, "call": true, "endcall": true,
	"filter": true, "endfilter": true, "block": true, "endblock": true,
	"extends": true, "include": true, "import": true, "from": true,
	"as": true, "with": true, "endwith": true, "do": true,
	"continue": true, "break": true,
	// Jinja2 tests
	"defined": true, "undefined": true, "number": true, "string": true,
	"mapping": true, "iterable": true, "callable": true, "sequence": true,
	"sameas": true, "escaped": true, "even": true, "odd": true,
	"divisibleby": true, "lower": true, "upper": true,
	// Jinja2 built-in filters
	"abs": true, "attr": true, "batch": true, "capitalize": true,
	"center": true, "count": true, "default": true, "dictsort": true,
	"escape": true, "filesizeformat": true, "first": true, "float": true,
	"forceescape": true, "format": true, "groupby": true, "indent": true,
	"int": true, "items": true, "join": true, "last": true, "length": true,
	"list": true, "map": true, "max": true, "min": true, "pprint": true,
	"random": true, "reject": true, "rejectattr": true, "replace": true,
	"reverse": true, "round": true, "safe": true, "select": true,
	"selectattr": true, "slice": true, "sort": true, "split": true,
	"striptags": true, "sum": true, "title": true, "tojson": true,
	"trim": true, "truncate": true, "unique": true, "urlencode": true,
	"urlize": true, "wordcount": true, "wordwrap": true, "xmlattr": true,
	// Home Assistant specific functions
	"states": true, "is_state": true, "state_attr": true, "is_state_attr": true,
	"has_value": true, "expand": true, "device_entities": true, "area_entities": true,
	"integration_entities": true, "device_attr": true, "device_id": true,
	"area_name": true, "area_id": true, "floor_id": true, "floor_name": true,
	"label_id": true, "label_name": true, "labels": true, "relative_time": true,
	"time_since": true, "timedelta": true, "strptime": true, "strftime": true,
	"as_timestamp": true, "as_datetime": true, "as_local": true, "as_timedelta": true,
	"today_at": true, "now": true, "utcnow": true, "distance": true,
	"closest": true, "iif": true, "log": true, "sin": true, "cos": true,
	"tan": true, "asin": true, "acos": true, "atan": true, "atan2": true,
	"sqrt": true, "e": true, "pi": true, "tau": true, "inf": true,
	"average": true, "median": true, "statistical_mode": true, "pack": true,
	"unpack": true, "ord": true, "base64_encode": true, "base64_decode": true,
	"slugify": true, "regex_match": true, "regex_search": true, "regex_replace": true,
	"regex_findall": true, "regex_findall_index": true, "from_json": true,
	"to_json": true, "value_json": true, "trigger": true, "this": true,
	"context": true, "repeat": true, "wait": true, "namespace": true,
	// Common loop variables
	"item": true, "loop": true, "index": true, "index0": true,
	"cycle": true, "depth": true, "depth0": true,
	"previtem": true, "nextitem": true, "changed": true,
	// Datetime attributes
	"year": true, "month": true, "day": true, "hour": true, "minute": true,
	"second": true, "microsecond": true, "weekday": true, "isoweekday": true,
	"isocalendar": true, "isoformat": true, "date": true, "time": true,
	"timestamp": true, "tzinfo": true, "tzname": true, "utcoffset": true,
	"dst": true, "timetuple": true,
	// State object attributes
	"state": true, "attributes": true, "entity_id": true, "domain": true,
	"object_id": true, "name": true, "last_changed": true, "last_updated": true,
	"last_reported": true, "context_id": true,
	// Trigger object attributes
	"platform": true, "event": true, "to_state": true, "from_state": true,
	"idx": true, "id": true, "description": true, "alias": true,
	// Additional common attributes
	"friendly_name": true, "icon": true, "unit_of_measurement": true,
	"device_class": true, "brightness": true, "color_temp": true, "hs_color": true,
	"rgb_color": true, "xy_color": true, "temperature": true, "humidity": true,
	"pressure": true, "position": true, "current_position": true,
	"current_temperature": true, "target_temperature": true, "hvac_mode": true,
	"hvac_action": true, "fan_mode": true, "swing_mode": true, "preset_mode": true,
	"speed": true, "percentage": true, "battery_level": true, "battery": true,
	"power": true, "voltage": true, "current": true, "energy": true,
	"elevation": true, "azimuth": true, "rising": true, "setting": true,
	"next_rising": true, "next_setting": true,
}

// BlueprintValidator validates Home Assistant Blueprint YAML files
type BlueprintValidator struct {
	FilePath            string
	Errors              []string
	Warnings            []string
	Data                map[string]interface{}
	DefinedInputs       map[string]bool
	UsedInputs          map[string]bool
	InputDefaults       map[string]interface{}
	InputSelectors      map[string]map[string]interface{}
	EntityInputs        map[string]bool
	InputDatetimeInputs map[string]bool
	DefinedVariables    map[string]bool
	JoinVariables       map[string]bool
	NonzeroDefaultVars  map[string]bool
}

// NewBlueprintValidator creates a new validator instance
func NewBlueprintValidator(filePath string) *BlueprintValidator {
	return &BlueprintValidator{
		FilePath:            filePath,
		Errors:              []string{},
		Warnings:            []string{},
		Data:                make(map[string]interface{}),
		DefinedInputs:       make(map[string]bool),
		UsedInputs:          make(map[string]bool),
		InputDefaults:       make(map[string]interface{}),
		InputSelectors:      make(map[string]map[string]interface{}),
		EntityInputs:        make(map[string]bool),
		InputDatetimeInputs: make(map[string]bool),
		DefinedVariables:    make(map[string]bool),
		JoinVariables:       make(map[string]bool),
		NonzeroDefaultVars:  make(map[string]bool),
	}
}

// Validate runs all validation checks
func (v *BlueprintValidator) Validate() bool {
	fmt.Printf("Validating: %s\n", v.FilePath)

	if !v.loadYAML() {
		return false
	}

	v.validateStructure()
	v.validateBlueprintSection()
	v.validateMode()
	v.validateInputs()
	v.validateHysteresisBoundaries()
	v.validateVariables()
	v.validateVersionSync()
	v.validateTriggers()
	v.validateConditions()
	v.validateActions()
	v.validateTemplates()
	v.validateInputReferences()
	v.checkReadmeExists()
	v.checkChangelogExists()

	return v.reportResults()
}

// loadYAML loads and parses the YAML file
func (v *BlueprintValidator) loadYAML() bool {
	content, err := os.ReadFile(v.FilePath)
	if err != nil {
		v.Errors = append(v.Errors, fmt.Sprintf("Failed to load file: %v", err))
		return false
	}

	// Custom handling for !input tags - replace with placeholder
	contentStr := string(content)
	inputRegex := regexp.MustCompile(`!input\s+(\S+)`)
	contentStr = inputRegex.ReplaceAllString(contentStr, `"!input $1"`)

	if err := yaml.Unmarshal([]byte(contentStr), &v.Data); err != nil {
		v.Errors = append(v.Errors, fmt.Sprintf("YAML syntax error: %v", err))
		return false
	}

	return true
}

// validateStructure validates root-level structure
func (v *BlueprintValidator) validateStructure() {
	// Check required root keys
	for _, key := range RequiredRootKeys {
		if _, ok := v.Data[key]; !ok {
			v.Errors = append(v.Errors, fmt.Sprintf("Missing required root key: '%s'", key))
		}
	}

	// Warn about variables not at root level
	if blueprint, ok := v.Data["blueprint"].(map[string]interface{}); ok {
		if _, hasVars := blueprint["variables"]; hasVars {
			v.Errors = append(v.Errors, "'variables' must be at root level, not nested under 'blueprint'")
		}
	}

	// Check for variables at root
	if variables, ok := v.Data["variables"]; ok && variables != nil {
		if _, isMap := variables.(map[string]interface{}); !isMap {
			v.Errors = append(v.Errors, "'variables' must be a dictionary")
		}
	}
}

// validateBlueprintSection validates blueprint metadata section
func (v *BlueprintValidator) validateBlueprintSection() {
	blueprint, ok := v.Data["blueprint"]
	if !ok {
		return
	}

	blueprintMap, ok := blueprint.(map[string]interface{})
	if !ok {
		v.Errors = append(v.Errors, "'blueprint' must be a dictionary")
		return
	}

	// Check required blueprint keys
	for _, key := range RequiredBlueprintKeys {
		if _, ok := blueprintMap[key]; !ok {
			v.Errors = append(v.Errors, fmt.Sprintf("Missing required blueprint key: '%s'", key))
		}
	}

	// Validate domain
	if domain, ok := blueprintMap["domain"].(string); ok {
		validDomains := []string{"automation", "script"}
		isValid := slices.Contains(validDomains, domain)
		if !isValid {
			v.Errors = append(v.Errors, fmt.Sprintf("Invalid domain '%s', must be one of: %v", domain, validDomains))
		}
	}
}

// validateMode validates automation mode
func (v *BlueprintValidator) validateMode() {
	mode, ok := v.Data["mode"]
	if !ok {
		return // Default mode is 'single', which is valid
	}

	modeStr, ok := mode.(string)
	if !ok {
		v.Errors = append(v.Errors, "'mode' must be a string")
		return
	}

	isValid := slices.Contains(ValidModes, modeStr)
	if !isValid {
		v.Errors = append(v.Errors, fmt.Sprintf("Invalid mode '%s', must be one of: %v", modeStr, ValidModes))
	}

	// Check for max when using queued/parallel
	if modeStr == "queued" || modeStr == "parallel" {
		if maxVal, ok := v.Data["max"]; ok {
			if maxInt, ok := maxVal.(int); !ok || maxInt < 1 {
				v.Errors = append(v.Errors, fmt.Sprintf("'max' must be a positive integer when mode is '%s'", modeStr))
			}
		}
	}
}

// validateInputs validates input definitions
func (v *BlueprintValidator) validateInputs() {
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
		v.Errors = append(v.Errors, "'blueprint.input' must be a dictionary")
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
			v.Errors = append(v.Errors, fmt.Sprintf("%s: Input must be a dictionary", currentPath))
			continue
		}

		// Check if this is an input group or actual input
		if nestedInput, hasNested := valueMap["input"]; hasNested {
			// This is a group
			nestedMap, ok := nestedInput.(map[string]interface{})
			if !ok {
				v.Errors = append(v.Errors, fmt.Sprintf("%s.input: Must be a dictionary", currentPath))
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
		v.Warnings = append(v.Warnings, fmt.Sprintf("%s: No selector defined (inputs should have selectors)", path))
		return
	}

	selectorMap, ok := selector.(map[string]interface{})
	if !ok {
		v.Errors = append(v.Errors, fmt.Sprintf("%s.selector: Must be a dictionary", path))
		return
	}

	// Track selector
	v.InputSelectors[inputName] = selectorMap

	// Validate selector type
	for selectorType := range selectorMap {
		if !ValidSelectorTypes[selectorType] {
			v.Warnings = append(v.Warnings, fmt.Sprintf("%s.selector: Unknown selector type '%s'", path, selectorType))
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
		v.Errors = append(v.Errors, fmt.Sprintf("%s.selector.select.options: Must be a list", path))
		return
	}

	for i, option := range optionsList {
		optionPath := fmt.Sprintf("%s.selector.select.options[%d]", path, i)

		switch opt := option.(type) {
		case nil:
			v.Errors = append(v.Errors, fmt.Sprintf("%s: Option cannot be None. Select options must be strings or label/value dicts with non-empty values.", optionPath))
		case map[string]interface{}:
			value := opt["value"]
			label := opt["label"]

			if value == nil {
				v.Errors = append(v.Errors, fmt.Sprintf("%s: Option value is None. Label/value options must have a non-empty 'value' field. Label: '%v'", optionPath, label))
			} else if valueStr, ok := value.(string); !ok {
				v.Errors = append(v.Errors, fmt.Sprintf("%s: Option value must be a string. Label: '%v'", optionPath, label))
			} else if valueStr == "" {
				v.Errors = append(v.Errors, fmt.Sprintf("%s: Option value cannot be empty string. Home Assistant treats empty values as None during import. Label: '%v'", optionPath, label))
			}
		case string:
			if opt == "" {
				v.Warnings = append(v.Warnings, fmt.Sprintf("%s: Empty string option. Consider using a meaningful value.", optionPath))
			}
		default:
			v.Errors = append(v.Errors, fmt.Sprintf("%s: Option must be a string or label/value dict", optionPath))
		}
	}
}

// validateHysteresisBoundaries validates hysteresis boundary pairs
func (v *BlueprintValidator) validateHysteresisBoundaries() {
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

			onValue, onErr := toFloat(onDefault)
			offValue, offErr := toFloat(offDefault)

			if onErr != nil || offErr != nil {
				continue
			}

			switch {
			case onValue < offValue:
				v.Errors = append(v.Errors, fmt.Sprintf(
					"Hysteresis %s inversion: '%s' (default=%v) should be greater than '%s' (default=%v). With ON < OFF, the system will chatter rapidly.",
					pattern.Description, inputName, onValue, offName, offValue))
			case onValue == offValue:
				v.Warnings = append(v.Warnings, fmt.Sprintf(
					"Hysteresis %s has no gap: '%s' and '%s' both default to %v. Without a gap between ON and OFF thresholds, there's no hysteresis protection against oscillation.",
					pattern.Description, inputName, offName, onValue))
			default:
				gap := onValue - offValue
				if onValue != 0 && gap/abs(onValue) < 0.1 {
					v.Warnings = append(v.Warnings, fmt.Sprintf(
						"Hysteresis %s gap may be too small: '%s' (default=%v) minus '%s' (default=%v) = %v. A larger gap provides better oscillation protection.",
						pattern.Description, inputName, onValue, offName, offValue, gap))
				}
			}
		}
	}
}

// validateVariables validates variables section
func (v *BlueprintValidator) validateVariables() {
	variables, ok := v.Data["variables"]
	if !ok {
		v.Warnings = append(v.Warnings, "No variables section defined")
		return
	}

	variablesMap, ok := variables.(map[string]interface{})
	if !ok {
		v.Errors = append(v.Errors, "'variables' must be a dictionary")
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
			v.collectInputRefs(valueStr)

			// Check for bare boolean literals
			v.checkBareBooleanLiterals(name, valueStr)

			// Check for unsafe math operations
			v.checkUnsafeMathOperations(name, valueStr)

			// Check for Python-style list methods
			v.checkPythonStyleMethods(name, valueStr)
		}
		definedVars[name] = true
	}

	// Check for blueprint_version
	if _, ok := variablesMap["blueprint_version"]; !ok {
		v.Warnings = append(v.Warnings, "No 'blueprint_version' variable defined")
	}
}

// validateVersionSync validates version sync between name and blueprint_version
func (v *BlueprintValidator) validateVersionSync() {
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
		v.Warnings = append(v.Warnings, fmt.Sprintf(
			"Version mismatch: blueprint name contains '%s' but blueprint_version is '%s'",
			nameVersionMatch, versionStr))
	}
}

// validateTriggers validates trigger definitions
func (v *BlueprintValidator) validateTriggers() {
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
		v.Errors = append(v.Errors, fmt.Sprintf("%s: Missing 'platform' or 'trigger' key", path))
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
			if containsVariableRef(valueTemplate) && !containsInputRef(valueTemplate) {
				v.Warnings = append(v.Warnings, fmt.Sprintf(
					"%s: Template trigger references variables. Trigger templates are evaluated separately and may not have access to blueprint variables.",
					path))
			}
		}
	}

	// Check entity_id is static (no templates in trigger entity_id)
	if entityID, ok := trigger["entity_id"].(string); ok {
		if strings.Contains(entityID, "{{") || strings.Contains(entityID, "{%") {
			v.Errors = append(v.Errors, fmt.Sprintf(
				"%s: entity_id cannot contain templates. Trigger entity_id must be a static string or !input reference.",
				path))
		}
	}

	// Check for 'for' with template (can be problematic)
	if forVal, ok := trigger["for"]; ok {
		if forStr, ok := forVal.(string); ok {
			if strings.Contains(forStr, "{{") {
				// Template in 'for' is valid but may reference unavailable variables
				if containsVariableRef(forStr) && !containsInputRef(forStr) {
					v.Warnings = append(v.Warnings, fmt.Sprintf(
						"%s: 'for' template references variables. Variables may not be available in trigger context.",
						path))
				}
			}
		}
	}
}

// validateConditions validates condition definitions
func (v *BlueprintValidator) validateConditions() {
	conditions, ok := v.Data["condition"]
	if !ok {
		return // Conditions are optional
	}

	v.validateConditionList(conditions, "condition")
}

// validateConditionList validates a list of conditions
func (v *BlueprintValidator) validateConditionList(conditions interface{}, path string) {
	switch cond := conditions.(type) {
	case []interface{}:
		for i, c := range cond {
			v.validateConditionList(c, fmt.Sprintf("%s[%d]", path, i))
		}
	case map[string]interface{}:
		v.validateSingleCondition(cond, path)
	}
}

// validateSingleCondition validates a single condition
func (v *BlueprintValidator) validateSingleCondition(condition map[string]interface{}, path string) {
	// Check for condition type
	condType, hasCondition := condition["condition"].(string)
	if !hasCondition {
		// Shorthand condition (e.g., just entity_id without condition key)
		if _, hasEntityID := condition["entity_id"]; hasEntityID {
			return // Valid shorthand
		}
		v.Warnings = append(v.Warnings, fmt.Sprintf("%s: Missing 'condition' key", path))
		return
	}

	// Validate condition type
	isValid := slices.Contains(ValidConditionTypes, condType)
	if !isValid {
		v.Warnings = append(v.Warnings, fmt.Sprintf("%s: Unknown condition type '%s'", path, condType))
	}

	// Validate nested conditions for and/or/not
	if condType == "and" || condType == "or" || condType == "not" {
		if conditions, ok := condition["conditions"]; ok {
			v.validateConditionList(conditions, fmt.Sprintf("%s.conditions", path))
		}
	}
}

// validateActions validates action definitions
func (v *BlueprintValidator) validateActions() {
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
			v.Warnings = append(v.Warnings, fmt.Sprintf("%s: Service '%s' should be in 'domain.service' format", path, service))
		}

		// Check data is not nil when present
		if data, hasData := action["data"]; hasData && data == nil {
			v.Errors = append(v.Errors, fmt.Sprintf("%s: 'data' block cannot be None/empty", path))
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
			v.Errors = append(v.Errors, fmt.Sprintf("%s: 'if' requires 'then'", path))
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
			v.Errors = append(v.Errors, fmt.Sprintf("%s.repeat: Missing 'sequence'", path))
		}
	}

	// Collect input refs from action
	v.collectInputRefsFromMap(action)
}

// validateTemplates validates Jinja2 templates throughout the blueprint
func (v *BlueprintValidator) validateTemplates() {
	v.validateTemplatesInValue(v.Data, "")
}

// validateTemplatesInValue recursively validates templates in a value
func (v *BlueprintValidator) validateTemplatesInValue(value interface{}, path string) {
	switch val := value.(type) {
	case string:
		v.validateTemplateString(val, path)
	case map[string]interface{}:
		for k, v2 := range val {
			newPath := path
			if newPath == "" {
				newPath = k
			} else {
				newPath = fmt.Sprintf("%s.%s", path, k)
			}
			v.validateTemplatesInValue(v2, newPath)
		}
	case []interface{}:
		for i, v2 := range val {
			v.validateTemplatesInValue(v2, fmt.Sprintf("%s[%d]", path, i))
		}
	}
}

// validateTemplateString validates a template string
func (v *BlueprintValidator) validateTemplateString(template, path string) {
	// Check for !input inside {{ }} blocks
	inputInTemplatePattern := regexp.MustCompile(`\{\{[^}]*!input[^}]*\}\}`)
	if inputInTemplatePattern.MatchString(template) {
		v.Errors = append(v.Errors, fmt.Sprintf(
			"%s: Cannot use !input tags inside {{ }} blocks. Assign the input to a variable first.",
			path))
	}

	// Check for balanced Jinja2 delimiters
	if strings.Count(template, "{{") != strings.Count(template, "}}") {
		v.Errors = append(v.Errors, fmt.Sprintf("%s: Unbalanced {{ }} delimiters", path))
	}
	if strings.Count(template, "{%") != strings.Count(template, "%}") {
		v.Errors = append(v.Errors, fmt.Sprintf("%s: Unbalanced {%% %%} delimiters", path))
	}
}

// validateInputReferences validates that all !input references point to defined inputs
func (v *BlueprintValidator) validateInputReferences() {
	// Find undefined inputs
	for inputName := range v.UsedInputs {
		if !v.DefinedInputs[inputName] {
			v.Errors = append(v.Errors, fmt.Sprintf(
				"Undefined input reference: '!input %s' - no matching input defined in blueprint.input",
				inputName))
		}
	}
}

// checkReadmeExists checks if README.md exists in the blueprint directory
func (v *BlueprintValidator) checkReadmeExists() {
	dir := filepath.Dir(v.FilePath)
	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		v.Warnings = append(v.Warnings, fmt.Sprintf("No README.md found in %s/ directory", filepath.Base(dir)))
	}
}

// checkChangelogExists checks if CHANGELOG.md exists in the blueprint directory
func (v *BlueprintValidator) checkChangelogExists() {
	dir := filepath.Dir(v.FilePath)
	changelogPath := filepath.Join(dir, "CHANGELOG.md")
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		v.Warnings = append(v.Warnings, fmt.Sprintf("No CHANGELOG.md found in %s/ directory", filepath.Base(dir)))
	}
}

// checkBareBooleanLiterals checks for bare boolean literals in templates
func (v *BlueprintValidator) checkBareBooleanLiterals(varName, value string) {
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

		// Skip if inside {{ }} blocks
		if strings.Contains(line, "{{") && strings.Contains(line, "}}") {
			continue
		}

		if stripped == "true" && !foundBareTrue {
			foundBareTrue = true
			v.Warnings = append(v.Warnings, fmt.Sprintf(
				"Variable '%s': Bare 'true' outputs STRING \"true\", not boolean. Use '{{ true }}' to output actual boolean.",
				varName))
		} else if stripped == "false" && !foundBareFalse {
			foundBareFalse = true
			v.Warnings = append(v.Warnings, fmt.Sprintf(
				"Variable '%s': Bare 'false' outputs STRING \"false\", not boolean. The string \"false\" is TRUTHY (non-empty). Use '{{ false }}' instead.",
				varName))
		}
	}
}

// checkUnsafeMathOperations checks for potentially unsafe math operations
func (v *BlueprintValidator) checkUnsafeMathOperations(varName, value string) {
	// Check for log() with potentially non-positive arguments
	logPattern := regexp.MustCompile(`log\s*\(\s*(\w+)\s*\)`)
	for _, match := range logPattern.FindAllStringSubmatch(value, -1) {
		varRef := match[1]
		guardPattern := regexp.MustCompile(fmt.Sprintf(`%s\s*>\s*0|%s\s+is\s+number`, regexp.QuoteMeta(varRef), regexp.QuoteMeta(varRef)))
		if !guardPattern.MatchString(value) {
			v.Warnings = append(v.Warnings, fmt.Sprintf(
				"Variable '%s': log(%s) may fail if %s <= 0. Consider adding a guard like 'if %s > 0'.",
				varName, varRef, varRef, varRef))
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
				v.Warnings = append(v.Warnings, fmt.Sprintf(
					"Variable '%s': sqrt() with potentially negative argument. Consider using sqrt(max(0, value)).",
					varName))
				break
			}
		}
	}
}

// checkPythonStyleMethods checks for Python-style list methods
func (v *BlueprintValidator) checkPythonStyleMethods(varName, value string) {
	// Check for patterns like [a,b].min() which should be [a,b] | min
	pythonMethodPattern := regexp.MustCompile(`\[[^\]]+\]\.(min|max|sum|sort|reverse)\(\)`)
	if pythonMethodPattern.MatchString(value) {
		v.Errors = append(v.Errors, fmt.Sprintf(
			"Variable '%s': Python-style list method detected. Use Jinja2 filter syntax instead (e.g., '[a,b] | min' not '[a,b].min()').",
			varName))
	}
}

// collectInputRefs collects !input references from a string
func (v *BlueprintValidator) collectInputRefs(value string) {
	if after, ok := strings.CutPrefix(value, "!input "); ok {
		inputName := after
		inputName = strings.TrimSpace(inputName)
		v.UsedInputs[inputName] = true
	}
}

// collectInputRefsFromMap recursively collects !input references from a map
func (v *BlueprintValidator) collectInputRefsFromMap(m map[string]interface{}) {
	for _, value := range m {
		switch val := value.(type) {
		case string:
			v.collectInputRefs(val)
		case map[string]interface{}:
			v.collectInputRefsFromMap(val)
		case []interface{}:
			for _, item := range val {
				switch itemVal := item.(type) {
				case string:
					v.collectInputRefs(itemVal)
				case map[string]interface{}:
					v.collectInputRefsFromMap(itemVal)
				}
			}
		}
	}
}

// reportResults prints validation results and returns success status
func (v *BlueprintValidator) reportResults() bool {
	fmt.Println()

	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	if len(v.Errors) > 0 {
		fmt.Printf("%s ERRORS:\n", red("X"))
		for _, err := range v.Errors {
			fmt.Printf("  %s %s\n", red("*"), err)
		}
		fmt.Println()
	}

	if len(v.Warnings) > 0 {
		fmt.Printf("%s WARNINGS:\n", yellow("!"))
		for _, warning := range v.Warnings {
			fmt.Printf("  %s %s\n", yellow("*"), warning)
		}
		fmt.Println()
	}

	if len(v.Errors) == 0 && len(v.Warnings) == 0 {
		fmt.Printf("%s Blueprint is valid!\n", green("OK"))
		return true
	}

	if len(v.Errors) == 0 {
		fmt.Printf("%s Blueprint is valid (with %d warnings)\n", green("OK"), len(v.Warnings))
		return true
	}

	fmt.Printf("%s Blueprint validation failed with %d errors\n", red("FAIL"), len(v.Errors))
	return false
}

// Helper functions

func toFloat(val interface{}) (float64, error) {
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

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func containsVariableRef(template string) bool {
	// Check if template contains variable references (not !input)
	varPattern := regexp.MustCompile(`\{\{[^}]*\b[a-z_][a-z0-9_]*\b[^}]*\}\}`)
	return varPattern.MatchString(template)
}

func containsInputRef(template string) bool {
	return strings.Contains(template, "!input")
}

// showHelp displays comprehensive usage information
func showHelp() {
	fmt.Print(`Usage: validate-blueprint <blueprint.yaml>
       validate-blueprint --all
       validate-blueprint --version
       validate-blueprint --help

validate-blueprint - A comprehensive Home Assistant Blueprint validator

Description:
  This tool performs comprehensive validation of Home Assistant blueprint files,
  checking for common errors and best practices.

Commands:
  <blueprint.yaml>     Validate a single blueprint file
  --all                Validate all blueprints in the repository
  --version, -v        Show version information
  --help, -h, help     Show this help message

Validation Checks:
  1.  YAML syntax validation
  2.  Blueprint schema validation (required keys, structure)
  3.  Input/selector validation
  4.  Template syntax checking
  5.  Service call structure validation
  6.  Version sync validation
  7.  Trigger validation
  8.  Condition validation
  9.  Mode validation
  10. Input reference validation
  11. Hysteresis boundary validation
  12. Variable definition validation
  13. README.md and CHANGELOG.md existence check

Valid Selectors:
  action, addon, area, attribute, boolean, color_rgb, color_temp, condition,
  conversation_agent, country, date, datetime, device, duration, entity, file,
  floor, icon, label, language, location, media, navigation, number, object,
  select, state, target, template, text, theme, time, trigger, ui_action, ui_color

Valid Modes:
  single, restart, queued, parallel

Valid Condition Types:
  and, or, not, state, numeric_state, template, time, zone, trigger, sun, device

Blueprint File Patterns (for --all):
  The tool searches for files matching these patterns:
  - *_pro.yaml
  - *_pro_blueprint.yaml
  - blueprint.yaml

Exit Codes:
  0    Validation passed (may have warnings)
  1    Validation failed (has errors) or invalid usage

Examples:
  validate-blueprint my_automation.yaml
  validate-blueprint blueprints/automation/motion_light.yaml
  validate-blueprint --all

`)
}

// isHelpRequested checks if any argument is a help flag
func isHelpRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--help" || arg == "-h" || arg == "help" {
			return true
		}
	}
	return false
}

// isVersionRequested checks if any argument is a version flag
func isVersionRequested(args []string) bool {
	for _, arg := range args {
		if arg == "--version" || arg == "-v" || arg == "version" {
			return true
		}
	}
	return false
}

// showVersion displays version information
func showVersion() {
	fmt.Printf("validate-blueprint %s\n", Version)
	fmt.Printf("  Build time: %s\n", BuildTime)
	fmt.Printf("  Git commit: %s\n", GitCommit)
}

// findAllBlueprints finds all blueprint YAML files in the repository
func findAllBlueprints(basePath string) ([]string, error) {
	var blueprints []string
	patterns := []string{"*_pro.yaml", "*_pro_blueprint.yaml", "blueprint.yaml"}
	excludeDirs := map[string]bool{
		".git": true, "node_modules": true, "venv": true, ".venv": true, "__pycache__": true,
	}

	err := filepath.Walk(basePath, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr // Return the actual error instead of nil
		}

		// Skip excluded directories
		if info.IsDir() && excludeDirs[info.Name()] {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches any pattern
		for _, pattern := range patterns {
			matched, matchErr := filepath.Match(pattern, info.Name())
			if matchErr != nil {
				continue // Skip invalid patterns
			}
			if matched {
				blueprints = append(blueprints, path)
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(blueprints)
	return blueprints, nil
}

// validateSingle validates a single blueprint file
func validateSingle(blueprintPath string) bool {
	validator := NewBlueprintValidator(blueprintPath)
	return validator.Validate()
}

// validateAll validates all blueprints in the repository
func validateAll() bool {
	// Navigate up from scripts/validate-blueprint-go/ to the repo root
	execPath, err := os.Executable()
	if err != nil {
		// Fall back to current directory
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			execPath = "."
		} else {
			execPath = cwd
		}
	}

	// Try to find repo root by looking for common markers
	repoRoot := filepath.Dir(filepath.Dir(filepath.Dir(execPath)))

	// If running from source, use relative path
	if _, statErr := os.Stat(filepath.Join(repoRoot, "blueprints")); os.IsNotExist(statErr) {
		// Try current working directory
		cwd, cwdErr := os.Getwd()
		if cwdErr == nil {
			if _, checkErr := os.Stat(filepath.Join(cwd, "blueprints")); checkErr == nil {
				repoRoot = cwd
			} else {
				// Go up directories looking for blueprints folder
				for range 5 {
					parent := filepath.Dir(repoRoot)
					if parent == repoRoot {
						break
					}
					repoRoot = parent
					if _, lookupErr := os.Stat(filepath.Join(repoRoot, "blueprints")); lookupErr == nil {
						break
					}
				}
			}
		}
	}

	blueprints, err := findAllBlueprints(repoRoot)
	if err != nil {
		fmt.Printf("Error finding blueprints: %v\n", err)
		return false
	}

	if len(blueprints) == 0 {
		fmt.Println("No blueprints found in repository")
		return false
	}

	fmt.Printf("Found %d blueprint(s) to validate\n\n", len(blueprints))

	type result struct {
		path    string
		success bool
	}
	var results []result

	for _, bp := range blueprints {
		validator := NewBlueprintValidator(bp)
		success := validator.Validate()
		results = append(results, result{path: bp, success: success})
		fmt.Println(strings.Repeat("-", 80))
		fmt.Println()
	}

	// Summary
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	passed := 0
	for _, r := range results {
		if r.success {
			passed++
		}
	}
	failed := len(results) - passed

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for _, r := range results {
		relPath, relErr := filepath.Rel(repoRoot, r.path)
		if relErr != nil || relPath == "" {
			relPath = r.path
		}
		if r.success {
			fmt.Printf("%s %s\n", green("OK"), relPath)
		} else {
			fmt.Printf("%s %s\n", red("X"), relPath)
		}
	}

	fmt.Println()
	fmt.Printf("Total: %d | Passed: %d | Failed: %d\n", len(results), passed, failed)

	return failed == 0
}

func main() {
	// Check for version first
	if isVersionRequested(os.Args[1:]) {
		showVersion()
		os.Exit(0)
	}

	// Check for help
	if len(os.Args) < 2 || isHelpRequested(os.Args[1:]) {
		showHelp()
		if len(os.Args) < 2 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	var success bool
	if os.Args[1] == "--all" {
		success = validateAll()
	} else {
		success = validateSingle(os.Args[1])
	}

	if success {
		os.Exit(0)
	}
	os.Exit(1)
}
