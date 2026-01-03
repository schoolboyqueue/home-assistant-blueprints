package validator

import (
	"testing"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/testfixtures"
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
			data:           testfixtures.Map{},
			expectedErrors: 0,
		},
		{
			name: "single condition as map",
			data: testfixtures.Map{
				"condition": testfixtures.StateCondition("light.test", "on"),
			},
			expectedErrors: 0,
		},
		{
			name: "condition list",
			data: testfixtures.Map{
				"condition": testfixtures.List{
					testfixtures.StateCondition("light.test", "on"),
					testfixtures.TimeCondition("07:00:00", ""),
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
			name:             "valid state condition",
			condition:        testfixtures.StateCondition("light.test", "on"),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "valid numeric_state condition",
			condition:        testfixtures.NumericStateCondition(testfixtures.CommonEntityIDs.Sensor, 20, nil),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "valid template condition",
			condition:        testfixtures.TemplateCondition(testfixtures.ValidTemplates.IsState),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "valid time condition",
			condition:        testfixtures.TimeCondition("07:00:00", "23:00:00"),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "shorthand condition (entity_id only)",
			condition:        testfixtures.ShorthandCondition("light.test", "on"),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "missing condition key (warning)",
			condition: testfixtures.Map{
				"value_template": "{{ true }}",
			},
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name: "unknown condition type (warning)",
			condition: testfixtures.Map{
				"condition": "unknown_type",
			},
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name: "valid and condition",
			condition: testfixtures.AndCondition(
				testfixtures.StateCondition("light.one", "on"),
				testfixtures.StateCondition("light.two", "on"),
			),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "valid or condition",
			condition: testfixtures.OrCondition(
				testfixtures.StateCondition("light.one", "on"),
			),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "valid not condition",
			condition: testfixtures.NotCondition(
				testfixtures.StateCondition("light.test", "on"),
			),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "zone condition",
			condition:        testfixtures.ZoneCondition(testfixtures.CommonEntityIDs.Person, testfixtures.CommonEntityIDs.Zone),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "trigger condition",
			condition:        testfixtures.TriggerCondition("motion_detected"),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "sun condition",
			condition:        testfixtures.SunCondition("sunset"),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "device condition",
			condition:        testfixtures.DeviceCondition("abc123", "light", "is_on"),
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
	v.Data = testfixtures.Map{
		"condition": testfixtures.AndCondition(
			testfixtures.OrCondition(
				testfixtures.StateCondition("light.one", "on"),
				testfixtures.Map{
					// Missing condition key - warning
					"value_template": "{{ true }}",
				},
			),
			testfixtures.Map{
				"condition": "unknown_type", // warning
			},
		),
	}

	v.ValidateConditions()

	// Should have 2 warnings: missing condition key and unknown type
	assert.Len(t, v.Errors, 0)
	assert.Len(t, v.Warnings, 2)
}

func TestConditionListProcessing(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	conditions := testfixtures.List{
		testfixtures.StateCondition("light.one", "on"),
		testfixtures.StateCondition("light.two", "off"),
	}

	v.validateConditionList(conditions, "condition")

	assert.Empty(t, v.Errors)
	assert.Empty(t, v.Warnings)
}

func TestConditionWithInvalidNestedConditions(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	condition := testfixtures.AndCondition(
		testfixtures.Map{
			"condition": "invalid_type",
		},
	)

	v.validateSingleCondition(condition, "condition")

	// Should have 1 warning for unknown condition type
	assert.Empty(t, v.Errors)
	assert.Len(t, v.Warnings, 1)
	assert.Contains(t, v.Warnings[0], "Unknown condition type")
}
