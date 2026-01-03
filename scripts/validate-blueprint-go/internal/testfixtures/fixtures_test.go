package testfixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMinimalBlueprint(t *testing.T) {
	t.Parallel()

	bp := MinimalBlueprint()

	require.Contains(t, bp, "blueprint")
	require.Contains(t, bp, "trigger")
	require.Contains(t, bp, "action")

	blueprint, ok := bp["blueprint"].(Map)
	require.True(t, ok, "blueprint should be a Map")
	assert.Equal(t, "Test Blueprint", blueprint["name"])
	assert.Equal(t, "automation", blueprint["domain"])
}

func TestStateTrigger(t *testing.T) {
	t.Parallel()

	trigger := StateTrigger("light.test")

	assert.Equal(t, "state", trigger["platform"])
	assert.Equal(t, "light.test", trigger["entity_id"])
}

func TestStateTriggerWithFor(t *testing.T) {
	t.Parallel()

	trigger := StateTriggerWithFor("light.test", "00:05:00")

	assert.Equal(t, "state", trigger["platform"])
	assert.Equal(t, "light.test", trigger["entity_id"])
	assert.Equal(t, "00:05:00", trigger["for"])
}

func TestTimeTrigger(t *testing.T) {
	t.Parallel()

	trigger := TimeTrigger("07:00:00")

	assert.Equal(t, "time", trigger["trigger"])
	assert.Equal(t, "07:00:00", trigger["at"])
}

func TestTemplateTrigger(t *testing.T) {
	t.Parallel()

	trigger := TemplateTrigger("{{ is_state('light.test', 'on') }}")

	assert.Equal(t, "template", trigger["platform"])
	assert.Equal(t, "{{ is_state('light.test', 'on') }}", trigger["value_template"])
}

func TestStateCondition(t *testing.T) {
	t.Parallel()

	cond := StateCondition("light.test", "on")

	assert.Equal(t, "state", cond["condition"])
	assert.Equal(t, "light.test", cond["entity_id"])
	assert.Equal(t, "on", cond["state"])
}

func TestAndCondition(t *testing.T) {
	t.Parallel()

	cond := AndCondition(
		StateCondition("light.one", "on"),
		StateCondition("light.two", "on"),
	)

	assert.Equal(t, "and", cond["condition"])
	conditions, ok := cond["conditions"].(List)
	require.True(t, ok, "conditions should be a List")
	assert.Len(t, conditions, 2)
}

func TestOrCondition(t *testing.T) {
	t.Parallel()

	cond := OrCondition(
		StateCondition("light.one", "on"),
		StateCondition("light.two", "off"),
	)

	assert.Equal(t, "or", cond["condition"])
	conditions, ok := cond["conditions"].(List)
	require.True(t, ok, "conditions should be a List")
	assert.Len(t, conditions, 2)
}

func TestNotCondition(t *testing.T) {
	t.Parallel()

	cond := NotCondition(
		StateCondition("light.test", "on"),
	)

	assert.Equal(t, "not", cond["condition"])
	conditions, ok := cond["conditions"].(List)
	require.True(t, ok, "conditions should be a List")
	assert.Len(t, conditions, 1)
}

func TestServiceCall(t *testing.T) {
	t.Parallel()

	action := ServiceCall("light.turn_on")

	assert.Equal(t, "light.turn_on", action["service"])
}

func TestServiceCallWithTarget(t *testing.T) {
	t.Parallel()

	action := ServiceCallWithTarget("light.turn_on", "light.living_room")

	assert.Equal(t, "light.turn_on", action["service"])
	target, ok := action["target"].(Map)
	require.True(t, ok, "target should be a Map")
	assert.Equal(t, "light.living_room", target["entity_id"])
}

func TestServiceCallWithData(t *testing.T) {
	t.Parallel()

	action := ServiceCallWithData("light.turn_on", Map{
		"brightness": 255,
	})

	assert.Equal(t, "light.turn_on", action["service"])
	data, ok := action["data"].(Map)
	require.True(t, ok, "data should be a Map")
	assert.Equal(t, 255, data["brightness"])
}

func TestDelayAction(t *testing.T) {
	t.Parallel()

	action := DelayAction("00:00:05")

	assert.Equal(t, "00:00:05", action["delay"])
}

func TestChooseAction(t *testing.T) {
	t.Parallel()

	option := ChooseOption(
		[]Map{StateCondition("light.test", "on")},
		[]Map{ServiceCall("light.turn_off")},
	)
	action := ChooseAction(option)

	assert.Contains(t, action, "choose")
	choose, ok := action["choose"].(List)
	require.True(t, ok, "choose should be a List")
	assert.Len(t, choose, 1)
}

func TestIfThenAction(t *testing.T) {
	t.Parallel()

	action := IfThenAction(
		[]Map{StateCondition("light.test", "on")},
		[]Map{ServiceCall("light.turn_off")},
	)

	assert.Contains(t, action, "if")
	assert.Contains(t, action, "then")
	assert.NotContains(t, action, "else")
}

func TestIfThenElseAction(t *testing.T) {
	t.Parallel()

	action := IfThenElseAction(
		[]Map{StateCondition("light.test", "on")},
		[]Map{ServiceCall("light.turn_off")},
		[]Map{ServiceCall("light.turn_on")},
	)

	assert.Contains(t, action, "if")
	assert.Contains(t, action, "then")
	assert.Contains(t, action, "else")
}

func TestRepeatCountAction(t *testing.T) {
	t.Parallel()

	action := RepeatCountAction(5, []Map{ServiceCall("light.toggle")})

	assert.Contains(t, action, "repeat")
	repeat, ok := action["repeat"].(Map)
	require.True(t, ok, "repeat should be a Map")
	assert.Equal(t, 5, repeat["count"])
	assert.Contains(t, repeat, "sequence")
}

func TestRepeatWhileAction(t *testing.T) {
	t.Parallel()

	action := RepeatWhileAction(
		[]Map{StateCondition("light.test", "on")},
		[]Map{DelayAction("00:00:01")},
	)

	assert.Contains(t, action, "repeat")
	repeat, ok := action["repeat"].(Map)
	require.True(t, ok, "repeat should be a Map")
	assert.Contains(t, repeat, "while")
	assert.Contains(t, repeat, "sequence")
}

func TestEntityInput(t *testing.T) {
	t.Parallel()

	input := EntityInput("My Entity")

	assert.Equal(t, "My Entity", input["name"])
	selector, ok := input["selector"].(Map)
	require.True(t, ok, "selector should be a Map")
	assert.Contains(t, selector, "entity")
}

func TestEntityInputWithDomain(t *testing.T) {
	t.Parallel()

	input := EntityInputWithDomain("Light Entity", "light")

	selector, ok := input["selector"].(Map)
	require.True(t, ok, "selector should be a Map")
	entity, ok := selector["entity"].(Map)
	require.True(t, ok, "entity should be a Map")
	assert.Equal(t, "light", entity["domain"])
}

func TestTextInput(t *testing.T) {
	t.Parallel()

	input := TextInput("My Text")

	assert.Equal(t, "My Text", input["name"])
	selector, ok := input["selector"].(Map)
	require.True(t, ok, "selector should be a Map")
	assert.Contains(t, selector, "text")
}

func TestNumberInput(t *testing.T) {
	t.Parallel()

	input := NumberInput("My Number", 0, 100)

	assert.Equal(t, "My Number", input["name"])
	selector, ok := input["selector"].(Map)
	require.True(t, ok, "selector should be a Map")
	number, ok := selector["number"].(Map)
	require.True(t, ok, "number should be a Map")
	assert.Equal(t, float64(0), number["min"])
	assert.Equal(t, float64(100), number["max"])
}

func TestSelectInput(t *testing.T) {
	t.Parallel()

	input := SelectInput("My Select", []string{"option1", "option2"})

	assert.Equal(t, "My Select", input["name"])
	selector, ok := input["selector"].(Map)
	require.True(t, ok, "selector should be a Map")
	selectConfig, ok := selector["select"].(Map)
	require.True(t, ok, "selectConfig should be a Map")
	options, ok := selectConfig["options"].(List)
	require.True(t, ok, "options should be a List")
	assert.Len(t, options, 2)
}

func TestSelectOption(t *testing.T) {
	t.Parallel()

	option := SelectOption("Label", "value")

	assert.Equal(t, "Label", option["label"])
	assert.Equal(t, "value", option["value"])
}

func TestInputGroup(t *testing.T) {
	t.Parallel()

	group := InputGroup("My Group", Map{
		"nested_input": EntityInput("Nested Entity"),
	})

	assert.Equal(t, "My Group", group["name"])
	assert.Contains(t, group, "input")
}

func TestAutomationBlueprint(t *testing.T) {
	t.Parallel()

	bp := AutomationBlueprint(
		"Test Automation",
		Map{"entity": EntityInput("Target Entity")},
		[]Map{StateTrigger("light.test")},
		[]Map{ServiceCall("light.turn_on")},
	)

	assert.Contains(t, bp, "blueprint")
	blueprint, ok := bp["blueprint"].(Map)
	require.True(t, ok, "blueprint should be a Map")
	assert.Equal(t, "Test Automation", blueprint["name"])
	assert.Equal(t, "automation", blueprint["domain"])

	triggers, ok := bp["trigger"].(List)
	require.True(t, ok, "trigger should be a List")
	assert.Len(t, triggers, 1)

	actions, ok := bp["action"].(List)
	require.True(t, ok, "action should be a List")
	assert.Len(t, actions, 1)
}

func TestAutomationBlueprintWithConditions(t *testing.T) {
	t.Parallel()

	bp := AutomationBlueprintWithConditions(
		"Test Automation",
		Map{},
		[]Map{StateTrigger("light.test")},
		[]Map{StateCondition("light.test", "on")},
		[]Map{ServiceCall("light.turn_off")},
	)

	assert.Contains(t, bp, "condition")
	conditions, ok := bp["condition"].(List)
	require.True(t, ok, "condition should be a List")
	assert.Len(t, conditions, 1)
}

func TestAutomationBlueprintWithVariables(t *testing.T) {
	t.Parallel()

	bp := AutomationBlueprintWithVariables(
		"Test Automation",
		Map{},
		Map{"my_var": "{{ states('sensor.temp') }}"},
		[]Map{StateTrigger("light.test")},
		[]Map{ServiceCall("light.turn_on")},
	)

	assert.Contains(t, bp, "variables")
	variables, ok := bp["variables"].(Map)
	require.True(t, ok, "variables should be a Map")
	assert.Contains(t, variables, "my_var")
}

func TestScriptBlueprint(t *testing.T) {
	t.Parallel()

	bp := ScriptBlueprint(
		"Test Script",
		Map{"entity": EntityInput("Target Entity")},
		[]Map{ServiceCall("light.turn_on")},
	)

	assert.Contains(t, bp, "blueprint")
	blueprint, ok := bp["blueprint"].(Map)
	require.True(t, ok, "blueprint should be a Map")
	assert.Equal(t, "Test Script", blueprint["name"])
	assert.Equal(t, "script", blueprint["domain"])

	sequence, ok := bp["sequence"].(List)
	require.True(t, ok, "sequence should be a List")
	assert.Len(t, sequence, 1)
}

func TestInputRef(t *testing.T) {
	t.Parallel()

	ref := InputRef("my_input")

	assert.Equal(t, "!input my_input", ref)
}

func TestInputRefInTemplate(t *testing.T) {
	t.Parallel()

	ref := InputRefInTemplate("my_entity")

	assert.Equal(t, "{{ states(!input my_entity) }}", ref)
}

func TestValidTemplates(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, ValidTemplates.States)
	assert.NotEmpty(t, ValidTemplates.IsState)
	assert.NotEmpty(t, ValidTemplates.Float)
	assert.NotEmpty(t, ValidTemplates.Conditional)
}

func TestInvalidTemplates(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, InvalidTemplates.UnbalancedOpen)
	assert.NotEmpty(t, InvalidTemplates.UnbalancedClose)
	assert.NotEmpty(t, InvalidTemplates.InputInTemplate)
}

func TestCommonEntityIDs(t *testing.T) {
	t.Parallel()

	assert.Contains(t, CommonEntityIDs.Light, "light.")
	assert.Contains(t, CommonEntityIDs.Switch, "switch.")
	assert.Contains(t, CommonEntityIDs.Sensor, "sensor.")
}

func TestCommonServices(t *testing.T) {
	t.Parallel()

	assert.Contains(t, CommonServices.LightTurnOn, "light.")
	assert.Contains(t, CommonServices.SwitchTurnOn, "switch.")
}

func TestModeSection(t *testing.T) {
	t.Parallel()

	mode := ModeSection("single")

	assert.Equal(t, "single", mode["mode"])
}

func TestModeSectionWithMax(t *testing.T) {
	t.Parallel()

	mode := ModeSectionWithMax("queued", 10)

	assert.Equal(t, "queued", mode["mode"])
	assert.Equal(t, 10, mode["max"])
}

func TestVariablesWithVersion(t *testing.T) {
	t.Parallel()

	vars := VariablesWithVersion("1.0", Map{
		"my_var": "value",
	})

	assert.Equal(t, "1.0", vars["blueprint_version"])
	assert.Equal(t, "value", vars["my_var"])
}

func TestNumericStateTrigger(t *testing.T) {
	t.Parallel()

	t.Run("with above only", func(t *testing.T) {
		t.Parallel()
		trigger := NumericStateTrigger("sensor.temp", 20, nil)
		assert.Equal(t, "numeric_state", trigger["platform"])
		assert.Equal(t, 20, trigger["above"])
		assert.NotContains(t, trigger, "below")
	})

	t.Run("with below only", func(t *testing.T) {
		t.Parallel()
		trigger := NumericStateTrigger("sensor.temp", nil, 30)
		assert.NotContains(t, trigger, "above")
		assert.Equal(t, 30, trigger["below"])
	})

	t.Run("with both", func(t *testing.T) {
		t.Parallel()
		trigger := NumericStateTrigger("sensor.temp", 20, 30)
		assert.Equal(t, 20, trigger["above"])
		assert.Equal(t, 30, trigger["below"])
	})
}

func TestTimeCondition(t *testing.T) {
	t.Parallel()

	t.Run("with after only", func(t *testing.T) {
		t.Parallel()
		cond := TimeCondition("07:00:00", "")
		assert.Equal(t, "time", cond["condition"])
		assert.Equal(t, "07:00:00", cond["after"])
		assert.NotContains(t, cond, "before")
	})

	t.Run("with both", func(t *testing.T) {
		t.Parallel()
		cond := TimeCondition("07:00:00", "23:00:00")
		assert.Equal(t, "07:00:00", cond["after"])
		assert.Equal(t, "23:00:00", cond["before"])
	})
}

func TestInvalidTrigger(t *testing.T) {
	t.Parallel()

	trigger := InvalidTrigger()

	assert.NotContains(t, trigger, "platform")
	assert.NotContains(t, trigger, "trigger")
	assert.Contains(t, trigger, "entity_id")
}

func TestBooleanInput(t *testing.T) {
	t.Parallel()

	input := BooleanInput("My Boolean")

	assert.Equal(t, "My Boolean", input["name"])
	selector, ok := input["selector"].(Map)
	require.True(t, ok, "selector should be a Map")
	assert.Contains(t, selector, "boolean")
}

func TestBooleanInputWithDefault(t *testing.T) {
	t.Parallel()

	input := BooleanInputWithDefault("My Boolean", true)

	assert.Equal(t, true, input["default"])
}
