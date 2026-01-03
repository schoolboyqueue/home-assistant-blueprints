package testfixtures

import (
	"time"
)

// =====================================
// Common Test Data Constants
// =====================================

// CommonEntityIDs provides common test entity IDs used across tests.
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
	InputText     string
	InputSelect   string
	Automation    string
	Script        string
	Scene         string
	Climate       string
	Cover         string
	Fan           string
	MediaPlayer   string
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
	InputText:     "input_text.test_message",
	InputSelect:   "input_select.test_mode",
	Automation:    "automation.test_automation",
	Script:        "script.test_script",
	Scene:         "scene.movie_time",
	Climate:       "climate.living_room",
	Cover:         "cover.garage_door",
	Fan:           "fan.bedroom",
	MediaPlayer:   "media_player.living_room",
}

// CommonServices provides common test service names.
var CommonServices = struct {
	LightTurnOn   string
	LightTurnOff  string
	LightToggle   string
	SwitchTurnOn  string
	SwitchTurnOff string
	SwitchToggle  string
	NotifyMobile  string
	ScriptTurnOn  string
	HomeAssistant string
}{
	LightTurnOn:   "light.turn_on",
	LightTurnOff:  "light.turn_off",
	LightToggle:   "light.toggle",
	SwitchTurnOn:  "switch.turn_on",
	SwitchTurnOff: "switch.turn_off",
	SwitchToggle:  "switch.toggle",
	NotifyMobile:  "notify.mobile",
	ScriptTurnOn:  "script.turn_on",
	HomeAssistant: "homeassistant.turn_on",
}

// =====================================
// HAState Fixtures
// =====================================

// NewHAState creates a basic HAState fixture.
func NewHAState(entityID, state string) HAState {
	return HAState{
		EntityID: entityID,
		State:    state,
	}
}

// NewHAStateWithAttrs creates an HAState with attributes.
func NewHAStateWithAttrs(entityID, state string, attrs map[string]any) HAState {
	return HAState{
		EntityID:   entityID,
		State:      state,
		Attributes: attrs,
	}
}

// NewHAStateWithTimestamps creates an HAState with timestamps.
func NewHAStateWithTimestamps(entityID, state string) HAState {
	now := time.Now().UTC().Format(time.RFC3339)
	return HAState{
		EntityID:    entityID,
		State:       state,
		LastChanged: now,
		LastUpdated: now,
	}
}

// NewHAStateFull creates a fully populated HAState.
func NewHAStateFull(entityID, state string, attrs map[string]any, contextID string) HAState {
	now := time.Now().UTC().Format(time.RFC3339)
	return HAState{
		EntityID:    entityID,
		State:       state,
		Attributes:  attrs,
		LastChanged: now,
		LastUpdated: now,
		Context: &HAContext{
			ID: contextID,
		},
	}
}

// LightState creates a common light state fixture.
func LightState(entityID, state string) HAState {
	attrs := map[string]any{
		"friendly_name": "Test Light",
		"supported_features": 0,
	}
	if state == "on" {
		attrs["brightness"] = 255
		attrs["color_mode"] = "brightness"
	}
	return NewHAStateWithAttrs(entityID, state, attrs)
}

// SensorState creates a common sensor state fixture.
func SensorState(entityID, value, unit string) HAState {
	return NewHAStateWithAttrs(entityID, value, map[string]any{
		"friendly_name":      "Test Sensor",
		"unit_of_measurement": unit,
	})
}

// BinarySensorState creates a common binary sensor state fixture.
func BinarySensorState(entityID, state string) HAState {
	return NewHAStateWithAttrs(entityID, state, map[string]any{
		"friendly_name": "Test Binary Sensor",
		"device_class":  "motion",
	})
}

// =====================================
// HAMessage Fixtures
// =====================================

// BoolPtr returns a pointer to a bool value.
func BoolPtr(b bool) *bool {
	return &b
}

// NewSuccessMessage creates a successful HAMessage response.
func NewSuccessMessage(id int, result any) HAMessage {
	return HAMessage{
		ID:      id,
		Type:    "result",
		Success: BoolPtr(true),
		Result:  result,
	}
}

// NewErrorMessage creates an error HAMessage response.
func NewErrorMessage(id int, code, message string) HAMessage {
	return HAMessage{
		ID:      id,
		Type:    "result",
		Success: BoolPtr(false),
		Error: &HAError{
			Code:    code,
			Message: message,
		},
	}
}

// NewEventMessage creates an event HAMessage.
func NewEventMessage(id int, variables map[string]any) HAMessage {
	return HAMessage{
		ID:   id,
		Type: "event",
		Event: &HAEvent{
			Variables: variables,
		},
	}
}

// NewTemplateResultMessage creates a template result event message.
func NewTemplateResultMessage(id int, result string) HAMessage {
	return HAMessage{
		ID:   id,
		Type: "event",
		Event: &HAEvent{
			Result: result,
		},
	}
}

// NewAuthRequiredMessage creates an auth_required message.
func NewAuthRequiredMessage() HAMessage {
	return HAMessage{
		Type: "auth_required",
	}
}

// NewAuthOKMessage creates an auth_ok message.
func NewAuthOKMessage() HAMessage {
	return HAMessage{
		Type: "auth_ok",
	}
}

// NewAuthInvalidMessage creates an auth_invalid message.
func NewAuthInvalidMessage(message string) HAMessage {
	return HAMessage{
		Type:    "auth_invalid",
		Message: message,
	}
}

// NewPongMessage creates a pong response message.
func NewPongMessage(id int) HAMessage {
	return HAMessage{
		ID:   id,
		Type: "pong",
	}
}

// =====================================
// HAError Fixtures
// =====================================

// CommonErrors provides common error fixtures.
var CommonErrors = struct {
	NotFound        HAError
	Unauthorized    HAError
	InvalidFormat   HAError
	HomeAssistant   HAError
	Timeout         HAError
	ServiceNotFound HAError
}{
	NotFound: HAError{
		Code:    "not_found",
		Message: "Entity not found",
	},
	Unauthorized: HAError{
		Code:    "unauthorized",
		Message: "Unauthorized",
	},
	InvalidFormat: HAError{
		Code:    "invalid_format",
		Message: "Invalid format",
	},
	HomeAssistant: HAError{
		Code:    "home_assistant_error",
		Message: "Home Assistant error",
	},
	Timeout: HAError{
		Code:    "timeout",
		Message: "Timeout waiting for response",
	},
	ServiceNotFound: HAError{
		Code:    "service_not_found",
		Message: "Service not found",
	},
}

// =====================================
// HAEvent Fixtures
// =====================================

// NewTriggerEvent creates an event with trigger variables.
func NewTriggerEvent(platform, entityID string) HAEvent {
	return HAEvent{
		Variables: map[string]any{
			"trigger": map[string]any{
				"platform":  platform,
				"entity_id": entityID,
			},
		},
	}
}

// NewStateTriggerEvent creates a state change trigger event.
func NewStateTriggerEvent(entityID, fromState, toState string) HAEvent {
	return HAEvent{
		Variables: map[string]any{
			"trigger": map[string]any{
				"platform":   "state",
				"entity_id":  entityID,
				"from_state": fromState,
				"to_state":   toState,
			},
		},
	}
}

// =====================================
// TraceInfo Fixtures
// =====================================

// NewTraceInfo creates a TraceInfo fixture.
func NewTraceInfo(itemID, runID, state string) TraceInfo {
	return TraceInfo{
		ItemID: itemID,
		RunID:  runID,
		State:  state,
	}
}

// NewTraceInfoFull creates a fully populated TraceInfo.
func NewTraceInfoFull(itemID, runID, state, scriptExecution string) TraceInfo {
	now := time.Now().UTC().Format(time.RFC3339)
	return TraceInfo{
		ItemID:          itemID,
		RunID:           runID,
		State:           state,
		ScriptExecution: scriptExecution,
		Timestamp: &Timestamp{
			Start:  now,
			Finish: now,
		},
		Context: &HAContext{
			ID: "test-context-id",
		},
	}
}

// =====================================
// Registry Entry Fixtures
// =====================================

// NewEntityEntry creates an EntityEntry fixture.
func NewEntityEntry(entityID, name, platform string) EntityEntry {
	return EntityEntry{
		EntityID:     entityID,
		Name:         name,
		OriginalName: name,
		Platform:     platform,
	}
}

// NewDeviceEntry creates a DeviceEntry fixture.
func NewDeviceEntry(id, name, manufacturer, model string) DeviceEntry {
	return DeviceEntry{
		ID:           id,
		Name:         name,
		Manufacturer: manufacturer,
		Model:        model,
	}
}

// NewAreaEntry creates an AreaEntry fixture.
func NewAreaEntry(areaID, name string) AreaEntry {
	return AreaEntry{
		AreaID: areaID,
		Name:   name,
	}
}

// =====================================
// HAConfig Fixtures
// =====================================

// NewHAConfig creates an HAConfig fixture.
func NewHAConfig() HAConfig {
	return HAConfig{
		Version:      "2024.1.0",
		LocationName: "Test Home",
		TimeZone:     "America/New_York",
		UnitSystem: map[string]string{
			"length":      "mi",
			"temperature": "°F",
		},
		State:      "RUNNING",
		Components: []string{"homeassistant", "automation", "script"},
	}
}
