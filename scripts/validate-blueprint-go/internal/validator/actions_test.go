package validator

import (
	"testing"

	"github.com/home-assistant-blueprints/testfixtures"
	"github.com/stretchr/testify/assert"
)

func TestValidateActions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		data           map[string]interface{}
		expectedErrors int
	}{
		{
			name:           "no actions",
			data:           testfixtures.Map{},
			expectedErrors: 0,
		},
		{
			name: "single action as map",
			data: testfixtures.Map{
				"action": testfixtures.ServiceCallWithTarget(
					testfixtures.CommonServices.LightTurnOn,
					"light.test",
				),
			},
			expectedErrors: 0,
		},
		{
			name: "action list",
			data: testfixtures.Map{
				"action": testfixtures.List{
					testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOn),
					testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOff),
				},
			},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.Data = tt.data
			v.ValidateActions()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateSingleAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		action           map[string]interface{}
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "valid service call",
			action: testfixtures.ServiceCallWithTarget(
				testfixtures.CommonServices.LightTurnOn,
				"light.test",
			),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "service with data",
			action: testfixtures.ServiceCallWithData(
				testfixtures.CommonServices.LightTurnOn,
				testfixtures.Map{"brightness": 255},
			),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "service with nil data (error)",
			action: testfixtures.Map{
				"service": testfixtures.CommonServices.LightTurnOn,
				"data":    nil,
			},
			expectedErrors:   1,
			expectedWarnings: 0,
		},
		{
			name: "invalid service format (warning)",
			action: testfixtures.Map{
				"service": "turn_on", // Missing domain
			},
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name: "service with input reference (valid)",
			action: testfixtures.Map{
				"service": testfixtures.InputRef("my_service"),
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "service with template (valid)",
			action: testfixtures.Map{
				"service": "{{ my_service }}",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "delay action",
			action:           testfixtures.DelayAction("00:00:05"),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "wait_template action",
			action:           testfixtures.WaitTemplateAction(testfixtures.ValidTemplates.IsState),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.validateSingleAction(tt.action, "action[0]")

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)
		})
	}
}

func TestValidateChooseAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		action         map[string]interface{}
		expectedErrors int
	}{
		{
			name: "valid choose block",
			action: testfixtures.ChooseAction(
				testfixtures.ChooseOption(
					[]testfixtures.Map{testfixtures.StateCondition("light.test", "on")},
					[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOff)},
				),
			),
			expectedErrors: 0,
		},
		{
			name: "choose with multiple options",
			action: testfixtures.ChooseAction(
				testfixtures.ChooseOption(
					[]testfixtures.Map{testfixtures.StateCondition("a", "on")},
					[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOn)},
				),
				testfixtures.ChooseOption(
					[]testfixtures.Map{testfixtures.StateCondition("b", "off")},
					[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOff)},
				),
			),
			expectedErrors: 0,
		},
		{
			name: "choose with default",
			action: testfixtures.ChooseActionWithDefault(
				[]testfixtures.Map{
					testfixtures.ChooseOption(
						[]testfixtures.Map{testfixtures.StateCondition("a", "on")},
						[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOn)},
					),
				},
				[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOff)},
			),
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.validateSingleAction(tt.action, "action[0]")

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateIfThenElseAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		action         map[string]interface{}
		expectedErrors int
	}{
		{
			name: "valid if/then",
			action: testfixtures.IfThenAction(
				[]testfixtures.Map{testfixtures.StateCondition("light.test", "on")},
				[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOff)},
			),
			expectedErrors: 0,
		},
		{
			name: "valid if/then/else",
			action: testfixtures.IfThenElseAction(
				[]testfixtures.Map{testfixtures.StateCondition("light.test", "on")},
				[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOff)},
				[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOn)},
			),
			expectedErrors: 0,
		},
		{
			name: "if without then (error)",
			action: testfixtures.Map{
				"if": testfixtures.List{
					testfixtures.StateCondition("light.test", "on"),
				},
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.validateSingleAction(tt.action, "action[0]")

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateRepeatAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		action         map[string]interface{}
		expectedErrors int
	}{
		{
			name: "valid repeat with sequence",
			action: testfixtures.RepeatCountAction(
				5,
				[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightToggle)},
			),
			expectedErrors: 0,
		},
		{
			name: "valid repeat while",
			action: testfixtures.RepeatWhileAction(
				[]testfixtures.Map{testfixtures.StateCondition("light.test", "on")},
				[]testfixtures.Map{testfixtures.DelayAction("00:00:01")},
			),
			expectedErrors: 0,
		},
		{
			name: "valid repeat until",
			action: testfixtures.RepeatUntilAction(
				[]testfixtures.Map{testfixtures.StateCondition("light.test", "off")},
				[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOff)},
			),
			expectedErrors: 0,
		},
		{
			name: "repeat without sequence (error)",
			action: testfixtures.Map{
				"repeat": testfixtures.Map{
					"count": 5,
				},
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.validateSingleAction(tt.action, "action[0]")

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestActionInputRefCollection(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	action := testfixtures.Map{
		"service": testfixtures.InputRef("my_service"),
		"target": testfixtures.Map{
			"entity_id": testfixtures.InputRef("target_entity"),
		},
		"data": testfixtures.Map{
			"brightness": testfixtures.InputRef("brightness_value"),
		},
	}

	v.validateSingleAction(action, "action[0]")

	// Check that input refs were collected
	assert.True(t, v.UsedInputs["my_service"])
	assert.True(t, v.UsedInputs["target_entity"])
	assert.True(t, v.UsedInputs["brightness_value"])
}

func TestNestedActionValidation(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.Data = testfixtures.Map{
		"action": testfixtures.List{
			testfixtures.ChooseAction(
				testfixtures.ChooseOption(
					[]testfixtures.Map{testfixtures.StateCondition("light.test", "on")},
					[]testfixtures.Map{
						testfixtures.IfThenElseAction(
							[]testfixtures.Map{testfixtures.TimeCondition("22:00:00", "")},
							[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightTurnOff)},
							[]testfixtures.Map{
								testfixtures.RepeatCountAction(
									3,
									[]testfixtures.Map{testfixtures.ServiceCall(testfixtures.CommonServices.LightToggle)},
								),
							},
						),
					},
				),
			),
		},
	}

	v.ValidateActions()

	// All nested structures should be validated without errors
	assert.Empty(t, v.Errors)
}
