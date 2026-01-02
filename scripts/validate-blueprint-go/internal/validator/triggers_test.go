package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTriggers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		data             map[string]interface{}
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name:           "no triggers",
			data:           map[string]interface{}{},
			expectedErrors: 0,
		},
		{
			name: "single trigger as map",
			data: map[string]interface{}{
				"trigger": map[string]interface{}{
					"platform": "state",
				},
			},
			expectedErrors: 0,
		},
		{
			name: "trigger list with valid triggers",
			data: map[string]interface{}{
				"trigger": []interface{}{
					map[string]interface{}{"platform": "state"},
					map[string]interface{}{"trigger": "time"},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "trigger missing platform or trigger key",
			data: map[string]interface{}{
				"trigger": []interface{}{
					map[string]interface{}{"entity_id": "light.test"},
				},
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.Data = tt.data
			v.ValidateTriggers()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)
		})
	}
}

func TestValidateSingleTrigger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		trigger          map[string]interface{}
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name: "valid state trigger",
			trigger: map[string]interface{}{
				"platform":  "state",
				"entity_id": "light.living_room",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "valid trigger using 'trigger' key",
			trigger: map[string]interface{}{
				"trigger": "time",
				"at":      "07:00:00",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "missing platform and trigger key",
			trigger:          map[string]interface{}{"entity_id": "light.test"},
			expectedErrors:   1,
			expectedWarnings: 0,
		},
		{
			name: "template trigger with variable reference (warning)",
			trigger: map[string]interface{}{
				"platform":       "template",
				"value_template": "{{ my_variable > 10 }}",
			},
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name: "template trigger with input reference (no warning)",
			trigger: map[string]interface{}{
				"platform":       "template",
				"value_template": "{{ states(!input my_entity) }}",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "entity_id with template (error)",
			trigger: map[string]interface{}{
				"platform":  "state",
				"entity_id": "{{ entity }}",
			},
			expectedErrors:   1,
			expectedWarnings: 0,
		},
		{
			name: "entity_id with input reference (valid)",
			trigger: map[string]interface{}{
				"platform":  "state",
				"entity_id": "!input target_entity",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "for with template containing variable (warning)",
			trigger: map[string]interface{}{
				"platform":  "state",
				"entity_id": "light.test",
				"for":       "{{ my_duration }}",
			},
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name: "for with input reference (no warning)",
			trigger: map[string]interface{}{
				"platform":  "state",
				"entity_id": "light.test",
				"for":       "!input duration",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "for with dict value (no warning)",
			trigger: map[string]interface{}{
				"platform":  "state",
				"entity_id": "light.test",
				"for": map[string]interface{}{
					"minutes": 5,
				},
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "for with static string",
			trigger: map[string]interface{}{
				"platform":  "state",
				"entity_id": "light.test",
				"for":       "00:05:00",
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.validateSingleTrigger(tt.trigger, "trigger[0]")

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)
		})
	}
}

func TestTriggerVariableWarnings(t *testing.T) {
	t.Parallel()

	t.Run("template trigger references variable but not input", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		trigger := map[string]interface{}{
			"platform":       "template",
			"value_template": "{{ threshold_value > 100 }}",
		}

		v.validateSingleTrigger(trigger, "trigger")

		assert.Len(t, v.Warnings, 1)
		assert.Contains(t, v.Warnings[0], "references variables")
	})

	t.Run("for clause references variable but not input", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		trigger := map[string]interface{}{
			"platform":  "state",
			"entity_id": "light.test",
			"for":       "{{ delay_time }}",
		}

		v.validateSingleTrigger(trigger, "trigger")

		assert.Len(t, v.Warnings, 1)
		assert.Contains(t, v.Warnings[0], "Variables may not be available")
	})
}

func TestMultipleTriggers(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.Data = map[string]interface{}{
		"trigger": []interface{}{
			map[string]interface{}{
				"platform":  "state",
				"entity_id": "light.one",
			},
			map[string]interface{}{
				"trigger": "time",
				"at":      "sunset",
			},
			map[string]interface{}{
				// Missing platform/trigger - error
				"entity_id": "light.error",
			},
		},
	}

	v.ValidateTriggers()

	assert.Len(t, v.Errors, 1, "Should have 1 error for missing platform")
}
