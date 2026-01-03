package testfixtures

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================
// HAState Tests
// =====================================

func TestNewHAState(t *testing.T) {
	t.Parallel()
	state := NewHAState("light.kitchen", "on")
	assert.Equal(t, "light.kitchen", state.EntityID)
	assert.Equal(t, "on", state.State)
}

func TestNewHAStateWithAttrs(t *testing.T) {
	t.Parallel()
	attrs := map[string]any{"brightness": 255}
	state := NewHAStateWithAttrs("light.kitchen", "on", attrs)
	assert.Equal(t, "light.kitchen", state.EntityID)
	assert.Equal(t, "on", state.State)
	assert.Equal(t, 255, state.Attributes["brightness"])
}

func TestNewHAStateWithTimestamps(t *testing.T) {
	t.Parallel()
	state := NewHAStateWithTimestamps("light.kitchen", "on")
	assert.NotEmpty(t, state.LastChanged)
	assert.NotEmpty(t, state.LastUpdated)
}

func TestLightState(t *testing.T) {
	t.Parallel()

	t.Run("on state", func(t *testing.T) {
		state := LightState("light.test", "on")
		assert.Equal(t, "light.test", state.EntityID)
		assert.Equal(t, "on", state.State)
		assert.Equal(t, 255, state.Attributes["brightness"])
	})

	t.Run("off state", func(t *testing.T) {
		state := LightState("light.test", "off")
		assert.Equal(t, "off", state.State)
		_, hasBrightness := state.Attributes["brightness"]
		assert.False(t, hasBrightness)
	})
}

// =====================================
// HAMessage Tests
// =====================================

func TestNewSuccessMessage(t *testing.T) {
	t.Parallel()
	msg := NewSuccessMessage(1, "pong")
	assert.Equal(t, 1, msg.ID)
	assert.Equal(t, "result", msg.Type)
	require.NotNil(t, msg.Success)
	assert.True(t, *msg.Success)
	assert.Equal(t, "pong", msg.Result)
}

func TestNewErrorMessage(t *testing.T) {
	t.Parallel()
	msg := NewErrorMessage(1, "not_found", "Entity not found")
	assert.Equal(t, 1, msg.ID)
	assert.Equal(t, "result", msg.Type)
	require.NotNil(t, msg.Success)
	assert.False(t, *msg.Success)
	require.NotNil(t, msg.Error)
	assert.Equal(t, "not_found", msg.Error.Code)
	assert.Equal(t, "Entity not found", msg.Error.Message)
}

func TestNewEventMessage(t *testing.T) {
	t.Parallel()
	vars := map[string]any{"key": "value"}
	msg := NewEventMessage(1, vars)
	assert.Equal(t, 1, msg.ID)
	assert.Equal(t, "event", msg.Type)
	require.NotNil(t, msg.Event)
	assert.Equal(t, "value", msg.Event.Variables["key"])
}

func TestNewAuthMessages(t *testing.T) {
	t.Parallel()

	t.Run("auth_required", func(t *testing.T) {
		msg := NewAuthRequiredMessage()
		assert.Equal(t, "auth_required", msg.Type)
	})

	t.Run("auth_ok", func(t *testing.T) {
		msg := NewAuthOKMessage()
		assert.Equal(t, "auth_ok", msg.Type)
	})

	t.Run("auth_invalid", func(t *testing.T) {
		msg := NewAuthInvalidMessage("bad token")
		assert.Equal(t, "auth_invalid", msg.Type)
		assert.Equal(t, "bad token", msg.Message)
	})
}

// =====================================
// Blueprint Fixture Tests
// =====================================

func TestMinimalBlueprint(t *testing.T) {
	t.Parallel()
	bp := MinimalBlueprint()
	require.Contains(t, bp, "blueprint")
	require.Contains(t, bp, "trigger")
	require.Contains(t, bp, "action")

	blueprint, ok := bp["blueprint"].(Map)
	require.True(t, ok)
	assert.Equal(t, "Test Blueprint", blueprint["name"])
	assert.Equal(t, "automation", blueprint["domain"])
}

func TestStateTrigger(t *testing.T) {
	t.Parallel()
	trigger := StateTrigger("light.test")
	assert.Equal(t, "state", trigger["platform"])
	assert.Equal(t, "light.test", trigger["entity_id"])
}

func TestNumericStateTrigger(t *testing.T) {
	t.Parallel()

	t.Run("with above", func(t *testing.T) {
		trigger := NumericStateTrigger("sensor.temp", 20, nil)
		assert.Equal(t, 20, trigger["above"])
		_, hasBelow := trigger["below"]
		assert.False(t, hasBelow)
	})

	t.Run("with below", func(t *testing.T) {
		trigger := NumericStateTrigger("sensor.temp", nil, 30)
		_, hasAbove := trigger["above"]
		assert.False(t, hasAbove)
		assert.Equal(t, 30, trigger["below"])
	})

	t.Run("with both", func(t *testing.T) {
		trigger := NumericStateTrigger("sensor.temp", 20, 30)
		assert.Equal(t, 20, trigger["above"])
		assert.Equal(t, 30, trigger["below"])
	})
}

func TestAndCondition(t *testing.T) {
	t.Parallel()
	cond := AndCondition(
		StateCondition("light.one", "on"),
		StateCondition("light.two", "on"),
	)
	assert.Equal(t, "and", cond["condition"])
	conditions, ok := cond["conditions"].(List)
	require.True(t, ok)
	assert.Len(t, conditions, 2)
}

func TestServiceCallFull(t *testing.T) {
	t.Parallel()
	action := ServiceCallFull("light.turn_on", "light.test", Map{"brightness": 255})
	assert.Equal(t, "light.turn_on", action["service"])

	target, ok := action["target"].(Map)
	require.True(t, ok)
	assert.Equal(t, "light.test", target["entity_id"])

	data, ok := action["data"].(Map)
	require.True(t, ok)
	assert.Equal(t, 255, data["brightness"])
}

func TestAutomationBlueprint(t *testing.T) {
	t.Parallel()
	bp := AutomationBlueprint(
		"Test Automation",
		Map{"entity": EntityInput("Target Entity")},
		[]Map{StateTrigger(CommonEntityIDs.Light)},
		[]Map{ServiceCall(CommonServices.LightTurnOff)},
	)

	blueprint, ok := bp["blueprint"].(Map)
	require.True(t, ok)
	assert.Equal(t, "Test Automation", blueprint["name"])
	assert.Equal(t, "automation", blueprint["domain"])

	triggers, ok := bp["trigger"].(List)
	require.True(t, ok)
	assert.Len(t, triggers, 1)

	actions, ok := bp["action"].(List)
	require.True(t, ok)
	assert.Len(t, actions, 1)
}

func TestInputFixtures(t *testing.T) {
	t.Parallel()

	t.Run("entity input", func(t *testing.T) {
		input := EntityInput("Target")
		assert.Equal(t, "Target", input["name"])
		selector, ok := input["selector"].(Map)
		require.True(t, ok)
		_, hasEntity := selector["entity"]
		assert.True(t, hasEntity)
	})

	t.Run("number input with default", func(t *testing.T) {
		input := NumberInputWithDefault("Brightness", 0, 100, 50)
		assert.Equal(t, "Brightness", input["name"])
		assert.Equal(t, 50.0, input["default"])
	})
}

func TestInputRef(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "!input my_entity", InputRef("my_entity"))
}

func TestInputRefInTemplate(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "{{ states(!input my_entity) }}", InputRefInTemplate("my_entity"))
}

// =====================================
// Common Constants Tests
// =====================================

func TestCommonEntityIDs(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "light.living_room", CommonEntityIDs.Light)
	assert.Equal(t, "sensor.temperature", CommonEntityIDs.Sensor)
	assert.Equal(t, "binary_sensor.motion", CommonEntityIDs.BinarySensor)
}

func TestCommonServices(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "light.turn_on", CommonServices.LightTurnOn)
	assert.Equal(t, "light.turn_off", CommonServices.LightTurnOff)
}

func TestCommonErrors(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "not_found", CommonErrors.NotFound.Code)
	assert.Equal(t, "unauthorized", CommonErrors.Unauthorized.Code)
}
