package validator

import (
	"testing"

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
			data:           map[string]interface{}{},
			expectedErrors: 0,
		},
		{
			name: "single action as map",
			data: map[string]interface{}{
				"action": map[string]interface{}{
					"service": "light.turn_on",
					"target": map[string]interface{}{
						"entity_id": "light.test",
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "action list",
			data: map[string]interface{}{
				"action": []interface{}{
					map[string]interface{}{
						"service": "light.turn_on",
					},
					map[string]interface{}{
						"service": "light.turn_off",
					},
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
			action: map[string]interface{}{
				"service": "light.turn_on",
				"target": map[string]interface{}{
					"entity_id": "light.test",
				},
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "service with data",
			action: map[string]interface{}{
				"service": "light.turn_on",
				"data": map[string]interface{}{
					"brightness": 255,
				},
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "service with nil data (error)",
			action: map[string]interface{}{
				"service": "light.turn_on",
				"data":    nil,
			},
			expectedErrors:   1,
			expectedWarnings: 0,
		},
		{
			name: "invalid service format (warning)",
			action: map[string]interface{}{
				"service": "turn_on", // Missing domain
			},
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name: "service with input reference (valid)",
			action: map[string]interface{}{
				"service": "!input my_service",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "service with template (valid)",
			action: map[string]interface{}{
				"service": "{{ my_service }}",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "delay action",
			action: map[string]interface{}{
				"delay": "00:00:05",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "wait_template action",
			action: map[string]interface{}{
				"wait_template": "{{ is_state('light.test', 'on') }}",
			},
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
			action: map[string]interface{}{
				"choose": []interface{}{
					map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"condition": "state",
								"entity_id": "light.test",
								"state":     "on",
							},
						},
						"sequence": []interface{}{
							map[string]interface{}{
								"service": "light.turn_off",
							},
						},
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "choose with multiple options",
			action: map[string]interface{}{
				"choose": []interface{}{
					map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{"condition": "state", "entity_id": "a", "state": "on"},
						},
						"sequence": []interface{}{
							map[string]interface{}{"service": "light.turn_on"},
						},
					},
					map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{"condition": "state", "entity_id": "b", "state": "off"},
						},
						"sequence": []interface{}{
							map[string]interface{}{"service": "light.turn_off"},
						},
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "choose with default",
			action: map[string]interface{}{
				"choose": []interface{}{
					map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{"condition": "state", "entity_id": "a", "state": "on"},
						},
						"sequence": []interface{}{
							map[string]interface{}{"service": "light.turn_on"},
						},
					},
				},
				"default": []interface{}{
					map[string]interface{}{"service": "light.turn_off"},
				},
			},
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
			action: map[string]interface{}{
				"if": []interface{}{
					map[string]interface{}{
						"condition": "state",
						"entity_id": "light.test",
						"state":     "on",
					},
				},
				"then": []interface{}{
					map[string]interface{}{
						"service": "light.turn_off",
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "valid if/then/else",
			action: map[string]interface{}{
				"if": []interface{}{
					map[string]interface{}{
						"condition": "state",
						"entity_id": "light.test",
						"state":     "on",
					},
				},
				"then": []interface{}{
					map[string]interface{}{
						"service": "light.turn_off",
					},
				},
				"else": []interface{}{
					map[string]interface{}{
						"service": "light.turn_on",
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "if without then (error)",
			action: map[string]interface{}{
				"if": []interface{}{
					map[string]interface{}{
						"condition": "state",
						"entity_id": "light.test",
						"state":     "on",
					},
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
			action: map[string]interface{}{
				"repeat": map[string]interface{}{
					"count": 5,
					"sequence": []interface{}{
						map[string]interface{}{
							"service": "light.toggle",
						},
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "valid repeat while",
			action: map[string]interface{}{
				"repeat": map[string]interface{}{
					"while": []interface{}{
						map[string]interface{}{
							"condition": "state",
							"entity_id": "light.test",
							"state":     "on",
						},
					},
					"sequence": []interface{}{
						map[string]interface{}{
							"delay": "00:00:01",
						},
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "valid repeat until",
			action: map[string]interface{}{
				"repeat": map[string]interface{}{
					"until": []interface{}{
						map[string]interface{}{
							"condition": "state",
							"entity_id": "light.test",
							"state":     "off",
						},
					},
					"sequence": []interface{}{
						map[string]interface{}{
							"service": "light.turn_off",
						},
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "repeat without sequence (error)",
			action: map[string]interface{}{
				"repeat": map[string]interface{}{
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
	action := map[string]interface{}{
		"service": "!input my_service",
		"target": map[string]interface{}{
			"entity_id": "!input target_entity",
		},
		"data": map[string]interface{}{
			"brightness": "!input brightness_value",
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
	v.Data = map[string]interface{}{
		"action": []interface{}{
			map[string]interface{}{
				"choose": []interface{}{
					map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"condition": "state",
								"entity_id": "light.test",
								"state":     "on",
							},
						},
						"sequence": []interface{}{
							map[string]interface{}{
								"if": []interface{}{
									map[string]interface{}{
										"condition": "time",
										"after":     "22:00:00",
									},
								},
								"then": []interface{}{
									map[string]interface{}{
										"service": "light.turn_off",
									},
								},
								"else": []interface{}{
									map[string]interface{}{
										"repeat": map[string]interface{}{
											"count": 3,
											"sequence": []interface{}{
												map[string]interface{}{
													"service": "light.toggle",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	v.ValidateActions()

	// All nested structures should be validated without errors
	assert.Empty(t, v.Errors)
}
