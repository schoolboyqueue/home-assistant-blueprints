// Package testfixtures provides shared test fixtures and factory functions
// for creating common Home Assistant blueprint YAML structures.
// This reduces repetitive test code and ensures consistency across tests.
package testfixtures

// Map is a shorthand for the common map type used in YAML structures
type Map = map[string]interface{}

// List is a shorthand for the common list type used in YAML structures
type List = []interface{}

// =====================================
// Blueprint Section Fixtures
// =====================================

// MinimalBlueprint returns a minimal valid blueprint structure
func MinimalBlueprint() Map {
	return Map{
		"blueprint": Map{
			"name":        "Test Blueprint",
			"description": "A test blueprint",
			"domain":      "automation",
			"input":       Map{},
		},
		"trigger": List{},
		"action":  List{},
	}
}

// BlueprintWithName returns a blueprint section with a custom name
func BlueprintWithName(name string) Map {
	return Map{
		"name":        name,
		"description": "A test blueprint",
		"domain":      "automation",
		"input":       Map{},
	}
}

// BlueprintSection returns a complete blueprint section
func BlueprintSection(name, description, domain string) Map {
	return Map{
		"name":        name,
		"description": description,
		"domain":      domain,
		"input":       Map{},
	}
}

// =====================================
// Trigger Fixtures
// =====================================

// StateTrigger creates a state trigger with the given entity_id
func StateTrigger(entityID string) Map {
	return Map{
		"platform":  "state",
		"entity_id": entityID,
	}
}

// StateTriggerWithFor creates a state trigger with a for clause
func StateTriggerWithFor(entityID string, forDuration interface{}) Map {
	return Map{
		"platform":  "state",
		"entity_id": entityID,
		"for":       forDuration,
	}
}

// StateTriggerWithToFrom creates a state trigger with to/from states
func StateTriggerWithToFrom(entityID, to, from string) Map {
	return Map{
		"platform":  "state",
		"entity_id": entityID,
		"to":        to,
		"from":      from,
	}
}

// TimeTrigger creates a time trigger
func TimeTrigger(at string) Map {
	return Map{
		"trigger": "time",
		"at":      at,
	}
}

// TemplateTrigger creates a template trigger
func TemplateTrigger(valueTemplate string) Map {
	return Map{
		"platform":       "template",
		"value_template": valueTemplate,
	}
}

// DeviceTrigger creates a device trigger
func DeviceTrigger(deviceID, domain, triggerType string) Map {
	return Map{
		"platform":  "device",
		"device_id": deviceID,
		"domain":    domain,
		"type":      triggerType,
	}
}

// NumericStateTrigger creates a numeric_state trigger
func NumericStateTrigger(entityID string, above, below interface{}) Map {
	t := Map{
		"platform":  "numeric_state",
		"entity_id": entityID,
	}
	if above != nil {
		t["above"] = above
	}
	if below != nil {
		t["below"] = below
	}
	return t
}

// TriggerWithPlatform creates a trigger with only the platform specified
func TriggerWithPlatform(platform string) Map {
	return Map{
		"platform": platform,
	}
}

// InvalidTrigger creates a trigger missing both platform and trigger key
func InvalidTrigger() Map {
	return Map{
		"entity_id": "light.test",
	}
}

// =====================================
// Condition Fixtures
// =====================================

// StateCondition creates a state condition
func StateCondition(entityID, state string) Map {
	return Map{
		"condition": "state",
		"entity_id": entityID,
		"state":     state,
	}
}

// NumericStateCondition creates a numeric_state condition
func NumericStateCondition(entityID string, above, below interface{}) Map {
	c := Map{
		"condition": "numeric_state",
		"entity_id": entityID,
	}
	if above != nil {
		c["above"] = above
	}
	if below != nil {
		c["below"] = below
	}
	return c
}

// TemplateCondition creates a template condition
func TemplateCondition(valueTemplate string) Map {
	return Map{
		"condition":      "template",
		"value_template": valueTemplate,
	}
}

// TimeCondition creates a time condition
func TimeCondition(after, before string) Map {
	c := Map{
		"condition": "time",
	}
	if after != "" {
		c["after"] = after
	}
	if before != "" {
		c["before"] = before
	}
	return c
}

// AndCondition creates an AND condition with nested conditions
func AndCondition(conditions ...Map) Map {
	condList := make(List, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	return Map{
		"condition":  "and",
		"conditions": condList,
	}
}

// OrCondition creates an OR condition with nested conditions
func OrCondition(conditions ...Map) Map {
	condList := make(List, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	return Map{
		"condition":  "or",
		"conditions": condList,
	}
}

// NotCondition creates a NOT condition with nested conditions
func NotCondition(conditions ...Map) Map {
	condList := make(List, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	return Map{
		"condition":  "not",
		"conditions": condList,
	}
}

// ZoneCondition creates a zone condition
func ZoneCondition(entityID, zone string) Map {
	return Map{
		"condition": "zone",
		"entity_id": entityID,
		"zone":      zone,
	}
}

// TriggerCondition creates a trigger condition
func TriggerCondition(id string) Map {
	return Map{
		"condition": "trigger",
		"id":        id,
	}
}

// SunCondition creates a sun condition
func SunCondition(afterEvent string) Map {
	return Map{
		"condition": "sun",
		"after":     afterEvent,
	}
}

// DeviceCondition creates a device condition
func DeviceCondition(deviceID, domain, condType string) Map {
	return Map{
		"condition": "device",
		"device_id": deviceID,
		"domain":    domain,
		"type":      condType,
	}
}

// ShorthandCondition creates a shorthand condition (entity_id + state without condition key)
func ShorthandCondition(entityID, state string) Map {
	return Map{
		"entity_id": entityID,
		"state":     state,
	}
}

// =====================================
// Action Fixtures
// =====================================

// ServiceCall creates a service call action
func ServiceCall(service string) Map {
	return Map{
		"service": service,
	}
}

// ServiceCallWithTarget creates a service call with target
func ServiceCallWithTarget(service, entityID string) Map {
	return Map{
		"service": service,
		"target": Map{
			"entity_id": entityID,
		},
	}
}

// ServiceCallWithData creates a service call with data
func ServiceCallWithData(service string, data Map) Map {
	return Map{
		"service": service,
		"data":    data,
	}
}

// ServiceCallFull creates a service call with target and data
func ServiceCallFull(service, entityID string, data Map) Map {
	return Map{
		"service": service,
		"target": Map{
			"entity_id": entityID,
		},
		"data": data,
	}
}

// DelayAction creates a delay action
func DelayAction(duration string) Map {
	return Map{
		"delay": duration,
	}
}

// DelayActionDict creates a delay action with a dictionary duration
func DelayActionDict(hours, minutes, seconds int) Map {
	return Map{
		"delay": Map{
			"hours":   hours,
			"minutes": minutes,
			"seconds": seconds,
		},
	}
}

// WaitTemplateAction creates a wait_template action
func WaitTemplateAction(template string) Map {
	return Map{
		"wait_template": template,
	}
}

// ChooseAction creates a choose action
func ChooseAction(options ...Map) Map {
	optList := make(List, len(options))
	for i, o := range options {
		optList[i] = o
	}
	return Map{
		"choose": optList,
	}
}

// ChooseActionWithDefault creates a choose action with a default sequence
func ChooseActionWithDefault(options, defaultSequence []Map) Map {
	optList := make(List, len(options))
	for i, o := range options {
		optList[i] = o
	}
	defList := make(List, len(defaultSequence))
	for i, d := range defaultSequence {
		defList[i] = d
	}
	return Map{
		"choose":  optList,
		"default": defList,
	}
}

// ChooseOption creates a choose option with conditions and sequence
func ChooseOption(conditions, sequence []Map) Map {
	condList := make(List, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	seqList := make(List, len(sequence))
	for i, s := range sequence {
		seqList[i] = s
	}
	return Map{
		"conditions": condList,
		"sequence":   seqList,
	}
}

// IfThenAction creates an if/then action
func IfThenAction(conditions, then []Map) Map {
	condList := make(List, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	thenList := make(List, len(then))
	for i, t := range then {
		thenList[i] = t
	}
	return Map{
		"if":   condList,
		"then": thenList,
	}
}

// IfThenElseAction creates an if/then/else action
func IfThenElseAction(conditions, then, elseActions []Map) Map {
	condList := make(List, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	thenList := make(List, len(then))
	for i, t := range then {
		thenList[i] = t
	}
	elseList := make(List, len(elseActions))
	for i, e := range elseActions {
		elseList[i] = e
	}
	return Map{
		"if":   condList,
		"then": thenList,
		"else": elseList,
	}
}

// RepeatCountAction creates a repeat action with count
func RepeatCountAction(count int, sequence []Map) Map {
	seqList := make(List, len(sequence))
	for i, s := range sequence {
		seqList[i] = s
	}
	return Map{
		"repeat": Map{
			"count":    count,
			"sequence": seqList,
		},
	}
}

// RepeatWhileAction creates a repeat while action
func RepeatWhileAction(conditions, sequence []Map) Map {
	condList := make(List, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	seqList := make(List, len(sequence))
	for i, s := range sequence {
		seqList[i] = s
	}
	return Map{
		"repeat": Map{
			"while":    condList,
			"sequence": seqList,
		},
	}
}

// RepeatUntilAction creates a repeat until action
func RepeatUntilAction(conditions, sequence []Map) Map {
	condList := make(List, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	seqList := make(List, len(sequence))
	for i, s := range sequence {
		seqList[i] = s
	}
	return Map{
		"repeat": Map{
			"until":    condList,
			"sequence": seqList,
		},
	}
}

// =====================================
// Input Fixtures
// =====================================

// EntityInput creates an entity input definition
func EntityInput(name string) Map {
	return Map{
		"name": name,
		"selector": Map{
			"entity": Map{},
		},
	}
}

// EntityInputWithDomain creates an entity input with domain filter
func EntityInputWithDomain(name, domain string) Map {
	return Map{
		"name": name,
		"selector": Map{
			"entity": Map{
				"domain": domain,
			},
		},
	}
}

// EntityInputWithDefault creates an entity input with a default value
func EntityInputWithDefault(name, defaultEntity string) Map {
	return Map{
		"name":    name,
		"default": defaultEntity,
		"selector": Map{
			"entity": Map{},
		},
	}
}

// TextInput creates a text input definition
func TextInput(name string) Map {
	return Map{
		"name": name,
		"selector": Map{
			"text": Map{},
		},
	}
}

// TextInputWithDefault creates a text input with a default value
func TextInputWithDefault(name, defaultValue string) Map {
	return Map{
		"name":    name,
		"default": defaultValue,
		"selector": Map{
			"text": Map{},
		},
	}
}

// NumberInput creates a number input definition
func NumberInput(name string, minVal, maxVal float64) Map {
	return Map{
		"name": name,
		"selector": Map{
			"number": Map{
				"min": minVal,
				"max": maxVal,
			},
		},
	}
}

// NumberInputWithDefault creates a number input with a default value
func NumberInputWithDefault(name string, minVal, maxVal, defaultVal float64) Map {
	return Map{
		"name":    name,
		"default": defaultVal,
		"selector": Map{
			"number": Map{
				"min": minVal,
				"max": maxVal,
			},
		},
	}
}

// BooleanInput creates a boolean input definition
func BooleanInput(name string) Map {
	return Map{
		"name": name,
		"selector": Map{
			"boolean": Map{},
		},
	}
}

// BooleanInputWithDefault creates a boolean input with a default value
func BooleanInputWithDefault(name string, defaultVal bool) Map {
	return Map{
		"name":    name,
		"default": defaultVal,
		"selector": Map{
			"boolean": Map{},
		},
	}
}

// SelectInput creates a select input with options
func SelectInput(name string, options []string) Map {
	optList := make(List, len(options))
	for i, o := range options {
		optList[i] = o
	}
	return Map{
		"name": name,
		"selector": Map{
			"select": Map{
				"options": optList,
			},
		},
	}
}

// SelectInputWithLabelValue creates a select input with label/value options
func SelectInputWithLabelValue(name string, options []Map) Map {
	optList := make(List, len(options))
	for i, o := range options {
		optList[i] = o
	}
	return Map{
		"name": name,
		"selector": Map{
			"select": Map{
				"options": optList,
			},
		},
	}
}

// SelectOption creates a select option with label and value
func SelectOption(label, value string) Map {
	return Map{
		"label": label,
		"value": value,
	}
}

// TimeInput creates a time input definition
func TimeInput(name string) Map {
	return Map{
		"name": name,
		"selector": Map{
			"time": Map{},
		},
	}
}

// DateTimeInput creates a datetime input definition
func DateTimeInput(name string) Map {
	return Map{
		"name": name,
		"selector": Map{
			"datetime": Map{},
		},
	}
}

// InputGroup creates an input group containing other inputs
func InputGroup(name string, inputs Map) Map {
	return Map{
		"name":  name,
		"input": inputs,
	}
}

// InputWithoutSelector creates an input without a selector (triggers warning)
func InputWithoutSelector(name string) Map {
	return Map{
		"name": name,
	}
}

// =====================================
// Template String Fixtures
// =====================================

// ValidTemplates provides commonly used valid templates
var ValidTemplates = struct {
	States        string
	IsState       string
	Float         string
	Int           string
	WithFilter    string
	Conditional   string
	StateAttr     string
	InputRef      string
	InputInStates string
	MultiLine     string
	NestedBraces  string
}{
	States:        "{{ states('light.test') }}",
	IsState:       "{{ is_state('light.test', 'on') }}",
	Float:         "{{ states('sensor.temp') | float(0) }}",
	Int:           "{{ states('sensor.count') | int(0) }}",
	WithFilter:    "{{ items | join(', ') }}",
	Conditional:   "{{ 'on' if value > 5 else 'off' }}",
	StateAttr:     "{{ state_attr('light.test', 'brightness') }}",
	InputRef:      "!input my_input",
	InputInStates: "{{ states(!input my_entity) }}",
	MultiLine:     "{% if is_state('light.test', 'on') %}on{% else %}off{% endif %}",
	NestedBraces:  "{{ {'key': value} | to_json }}",
}

// InvalidTemplates provides commonly used invalid templates
var InvalidTemplates = struct {
	UnbalancedOpen  string
	UnbalancedClose string
	UnbalancedBlock string
	InputInTemplate string
	MultipleErrors  string
}{
	UnbalancedOpen:  "{{ states('light.test')",
	UnbalancedClose: "states('light.test') }}",
	UnbalancedBlock: "{% if true",
	InputInTemplate: "{{ !input sensor }}",
	MultipleErrors:  "{{ !input x }} {% incomplete",
}

// =====================================
// Variables Section Fixtures
// =====================================

// VariablesSection creates a variables section
func VariablesSection(vars Map) Map {
	return vars
}

// VariablesWithVersion creates a variables section with blueprint_version
func VariablesWithVersion(version string, otherVars Map) Map {
	vars := Map{
		"blueprint_version": version,
	}
	for k, v := range otherVars {
		vars[k] = v
	}
	return vars
}

// =====================================
// Complete Blueprint Fixtures
// =====================================

// AutomationBlueprint creates a complete automation blueprint
func AutomationBlueprint(name string, inputs Map, triggers, actions []Map) Map {
	trigList := make(List, len(triggers))
	for i, t := range triggers {
		trigList[i] = t
	}
	actList := make(List, len(actions))
	for i, a := range actions {
		actList[i] = a
	}
	return Map{
		"blueprint": Map{
			"name":        name,
			"description": "Auto-generated test blueprint",
			"domain":      "automation",
			"input":       inputs,
		},
		"trigger": trigList,
		"action":  actList,
	}
}

// AutomationBlueprintWithConditions creates a complete automation with conditions
func AutomationBlueprintWithConditions(name string, inputs Map, triggers, conditions, actions []Map) Map {
	bp := AutomationBlueprint(name, inputs, triggers, actions)
	condList := make(List, len(conditions))
	for i, c := range conditions {
		condList[i] = c
	}
	bp["condition"] = condList
	return bp
}

// AutomationBlueprintWithVariables creates a complete automation with variables
func AutomationBlueprintWithVariables(name string, inputs, variables Map, triggers, actions []Map) Map {
	bp := AutomationBlueprint(name, inputs, triggers, actions)
	bp["variables"] = variables
	return bp
}

// ScriptBlueprint creates a complete script blueprint
func ScriptBlueprint(name string, inputs Map, sequence []Map) Map {
	seqList := make(List, len(sequence))
	for i, s := range sequence {
		seqList[i] = s
	}
	return Map{
		"blueprint": Map{
			"name":        name,
			"description": "Auto-generated test script blueprint",
			"domain":      "script",
			"input":       inputs,
		},
		"sequence": seqList,
	}
}

// =====================================
// Input Reference Fixtures
// =====================================

// InputRef creates an input reference string
func InputRef(inputName string) string {
	return "!input " + inputName
}

// InputRefInTemplate creates an input reference inside a template
func InputRefInTemplate(inputName string) string {
	return "{{ states(!input " + inputName + ") }}"
}

// =====================================
// Common Test Data
// =====================================

// CommonEntityIDs provides common test entity IDs
var CommonEntityIDs = struct {
	Light         string
	Switch        string
	Sensor        string
	BinarySensor  string
	Person        string
	Zone          string
	InputBoolean  string
	InputNumber   string
	InputDateTime string
}{
	Light:         "light.living_room",
	Switch:        "switch.bedroom",
	Sensor:        "sensor.temperature",
	BinarySensor:  "binary_sensor.motion",
	Person:        "person.me",
	Zone:          "zone.home",
	InputBoolean:  "input_boolean.test",
	InputNumber:   "input_number.threshold",
	InputDateTime: "input_datetime.alarm",
}

// CommonServices provides common test service names
var CommonServices = struct {
	LightTurnOn   string
	LightTurnOff  string
	LightToggle   string
	SwitchTurnOn  string
	SwitchTurnOff string
	NotifyMobile  string
	ScriptTurnOn  string
}{
	LightTurnOn:   "light.turn_on",
	LightTurnOff:  "light.turn_off",
	LightToggle:   "light.toggle",
	SwitchTurnOn:  "switch.turn_on",
	SwitchTurnOff: "switch.turn_off",
	NotifyMobile:  "notify.mobile",
	ScriptTurnOn:  "script.turn_on",
}

// =====================================
// Mode Fixtures
// =====================================

// ModeSection creates a mode section
func ModeSection(mode string) Map {
	return Map{
		"mode": mode,
	}
}

// ModeSectionWithMax creates a mode section with max setting
func ModeSectionWithMax(mode string, maxConcurrent int) Map {
	return Map{
		"mode": mode,
		"max":  maxConcurrent,
	}
}
