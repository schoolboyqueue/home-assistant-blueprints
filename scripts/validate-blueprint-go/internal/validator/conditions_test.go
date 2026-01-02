package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConditions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		data             map[string]interface{}
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name:           "no conditions",
			data:           map[string]interface{}{},
			expectedErrors: 0,
		},
		{
			name: "single condition as map",
			data: map[string]interface{}{
				"condition": map[string]interface{}{
					"condition": "state",
					"entity_id": "light.test",
					"state":     "on",
				},
			},
			expectedErrors: 0,
		},
		{
			name: "condition list",
			data: map[string]interface{}{
				"condition": []interface{}{
					map[string]interface{}{
						"condition": "state",
						"entity_id": "light.test",
						"state":     "on",
					},
					map[string]interface{}{
						"condition": "time",
						"after":     "07:00:00",
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
			v.ValidateConditions()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)
		})
	}
}

func TestValidateSingleCondition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		condition        map[string]interface{}
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "valid state condition",
			condition: map[string]interface{}{
				"condition": "state",
				"entity_id": "light.test",
				"state":     "on",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "valid numeric_state condition",
			condition: map[string]interface{}{
				"condition": "numeric_state",
				"entity_id": "sensor.temperature",
				"above":     20,
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "valid template condition",
			condition: map[string]interface{}{
				"condition":      "template",
				"value_template": "{{ is_state('light.test', 'on') }}",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "valid time condition",
			condition: map[string]interface{}{
				"condition": "time",
				"after":     "07:00:00",
				"before":    "23:00:00",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "shorthand condition (entity_id only)",
			condition: map[string]interface{}{
				"entity_id": "light.test",
				"state":     "on",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "missing condition key (warning)",
			condition: map[string]interface{}{
				"value_template": "{{ true }}",
			},
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name: "unknown condition type (warning)",
			condition: map[string]interface{}{
				"condition": "unknown_type",
			},
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name: "valid and condition",
			condition: map[string]interface{}{
				"condition": "and",
				"conditions": []interface{}{
					map[string]interface{}{
						"condition": "state",
						"entity_id": "light.one",
						"state":     "on",
					},
					map[string]interface{}{
						"condition": "state",
						"entity_id": "light.two",
						"state":     "on",
					},
				},
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "valid or condition",
			condition: map[string]interface{}{
				"condition": "or",
				"conditions": []interface{}{
					map[string]interface{}{
						"condition": "state",
						"entity_id": "light.one",
						"state":     "on",
					},
				},
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "valid not condition",
			condition: map[string]interface{}{
				"condition": "not",
				"conditions": []interface{}{
					map[string]interface{}{
						"condition": "state",
						"entity_id": "light.test",
						"state":     "on",
					},
				},
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "zone condition",
			condition: map[string]interface{}{
				"condition": "zone",
				"entity_id": "person.me",
				"zone":      "zone.home",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "trigger condition",
			condition: map[string]interface{}{
				"condition": "trigger",
				"id":        "motion_detected",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "sun condition",
			condition: map[string]interface{}{
				"condition": "sun",
				"after":     "sunset",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "device condition",
			condition: map[string]interface{}{
				"condition": "device",
				"device_id": "abc123",
				"domain":    "light",
				"type":      "is_on",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.validateSingleCondition(tt.condition, "condition[0]")

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)
		})
	}
}

func TestNestedConditions(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.Data = map[string]interface{}{
		"condition": map[string]interface{}{
			"condition": "and",
			"conditions": []interface{}{
				map[string]interface{}{
					"condition": "or",
					"conditions": []interface{}{
						map[string]interface{}{
							"condition": "state",
							"entity_id": "light.one",
							"state":     "on",
						},
						map[string]interface{}{
							// Missing condition key - warning
							"value_template": "{{ true }}",
						},
					},
				},
				map[string]interface{}{
					"condition": "unknown_type", // warning
				},
			},
		},
	}

	v.ValidateConditions()

	// Should have 2 warnings: missing condition key and unknown type
	assert.Len(t, v.Errors, 0)
	assert.Len(t, v.Warnings, 2)
}

func TestConditionListProcessing(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	conditions := []interface{}{
		map[string]interface{}{
			"condition": "state",
			"entity_id": "light.one",
			"state":     "on",
		},
		map[string]interface{}{
			"condition": "state",
			"entity_id": "light.two",
			"state":     "off",
		},
	}

	v.validateConditionList(conditions, "condition")

	assert.Empty(t, v.Errors)
	assert.Empty(t, v.Warnings)
}

func TestConditionWithInvalidNestedConditions(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	condition := map[string]interface{}{
		"condition": "and",
		"conditions": []interface{}{
			map[string]interface{}{
				"condition": "invalid_type",
			},
		},
	}

	v.validateSingleCondition(condition, "condition")

	// Should have 1 warning for unknown condition type
	assert.Empty(t, v.Errors)
	assert.Len(t, v.Warnings, 1)
	assert.Contains(t, v.Warnings[0], "Unknown condition type")
}
