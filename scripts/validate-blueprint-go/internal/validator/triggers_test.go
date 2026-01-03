package validator

import (
	"testing"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/testfixtures"
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
			data:           testfixtures.Map{},
			expectedErrors: 0,
		},
		{
			name: "single trigger as map",
			data: testfixtures.Map{
				"trigger": testfixtures.TriggerWithPlatform("state"),
			},
			expectedErrors: 0,
		},
		{
			name: "trigger list with valid triggers",
			data: testfixtures.Map{
				"trigger": testfixtures.List{
					testfixtures.TriggerWithPlatform("state"),
					testfixtures.TimeTrigger("07:00:00"),
				},
			},
			expectedErrors: 0,
		},
		{
			name: "trigger missing platform or trigger key",
			data: testfixtures.Map{
				"trigger": testfixtures.List{
					testfixtures.InvalidTrigger(),
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
			name:             "valid state trigger",
			trigger:          testfixtures.StateTrigger(testfixtures.CommonEntityIDs.Light),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "valid trigger using 'trigger' key",
			trigger:          testfixtures.TimeTrigger("07:00:00"),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "missing platform and trigger key",
			trigger:          testfixtures.InvalidTrigger(),
			expectedErrors:   1,
			expectedWarnings: 0,
		},
		{
			name:             "template trigger with variable reference (warning)",
			trigger:          testfixtures.TemplateTrigger("{{ my_variable > 10 }}"),
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name:             "template trigger with input reference (no warning)",
			trigger:          testfixtures.TemplateTrigger("{{ states(!input my_entity) }}"),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "entity_id with template (error)",
			trigger: testfixtures.Map{
				"platform":  "state",
				"entity_id": "{{ entity }}",
			},
			expectedErrors:   1,
			expectedWarnings: 0,
		},
		{
			name: "entity_id with input reference (valid)",
			trigger: testfixtures.Map{
				"platform":  "state",
				"entity_id": testfixtures.InputRef("target_entity"),
			},
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "for with template containing variable (warning)",
			trigger:          testfixtures.StateTriggerWithFor("light.test", "{{ my_duration }}"),
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name:             "for with input reference (no warning)",
			trigger:          testfixtures.StateTriggerWithFor("light.test", testfixtures.InputRef("duration")),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name: "for with dict value (no warning)",
			trigger: testfixtures.StateTriggerWithFor("light.test", testfixtures.Map{
				"minutes": 5,
			}),
			expectedErrors:   0,
			expectedWarnings: 0,
		},
		{
			name:             "for with static string",
			trigger:          testfixtures.StateTriggerWithFor("light.test", "00:05:00"),
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
		trigger := testfixtures.TemplateTrigger("{{ threshold_value > 100 }}")

		v.validateSingleTrigger(trigger, "trigger")

		assert.Len(t, v.Warnings, 1)
		assert.Contains(t, v.Warnings[0], "references variables")
	})

	t.Run("for clause references variable but not input", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		trigger := testfixtures.StateTriggerWithFor("light.test", "{{ delay_time }}")

		v.validateSingleTrigger(trigger, "trigger")

		assert.Len(t, v.Warnings, 1)
		assert.Contains(t, v.Warnings[0], "Variables may not be available")
	})
}

func TestMultipleTriggers(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.Data = testfixtures.Map{
		"trigger": testfixtures.List{
			testfixtures.StateTrigger("light.one"),
			testfixtures.TimeTrigger("sunset"),
			testfixtures.InvalidTrigger(), // Missing platform/trigger - error
		},
	}

	v.ValidateTriggers()

	assert.Len(t, v.Errors, 1, "Should have 1 error for missing platform")
}
