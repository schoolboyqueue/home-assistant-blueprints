package validator

import (
	"testing"

	"github.com/home-assistant-blueprints/testfixtures"
	"github.com/stretchr/testify/assert"
)

func TestValidateTemplates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		data           map[string]interface{}
		expectedErrors int
	}{
		{
			name:           "empty data",
			data:           testfixtures.Map{},
			expectedErrors: 0,
		},
		{
			name: "valid templates",
			data: testfixtures.Map{
				"template1": testfixtures.ValidTemplates.States,
				"template2": testfixtures.ValidTemplates.MultiLine,
			},
			expectedErrors: 0,
		},
		{
			name: "unbalanced double braces",
			data: testfixtures.Map{
				"template": testfixtures.InvalidTemplates.UnbalancedOpen,
			},
			expectedErrors: 1,
		},
		{
			name: "unbalanced block braces",
			data: testfixtures.Map{
				"template": testfixtures.InvalidTemplates.UnbalancedBlock,
			},
			expectedErrors: 1,
		},
		{
			name: "input inside template block",
			data: testfixtures.Map{
				"template": testfixtures.InvalidTemplates.InputInTemplate,
			},
			expectedErrors: 1,
		},
		{
			name: "input outside template block (valid)",
			data: testfixtures.Map{
				"value": testfixtures.ValidTemplates.InputRef,
			},
			expectedErrors: 0,
		},
		{
			name: "nested templates all valid",
			data: testfixtures.Map{
				"outer": testfixtures.Map{
					"inner": "{{ value }}",
					"list": testfixtures.List{
						"{{ item1 }}",
						"{{ item2 }}",
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "nested template with error",
			data: testfixtures.Map{
				"outer": testfixtures.Map{
					"inner": testfixtures.InvalidTemplates.UnbalancedOpen,
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
			v.ValidateTemplates()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateTemplateString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		template       string
		expectedErrors int
	}{
		{
			name:           "plain text",
			template:       "plain text without templates",
			expectedErrors: 0,
		},
		{
			name:           "valid double brace template",
			template:       testfixtures.ValidTemplates.States,
			expectedErrors: 0,
		},
		{
			name:           "valid block template",
			template:       testfixtures.ValidTemplates.MultiLine,
			expectedErrors: 0,
		},
		{
			name:           "mixed valid template",
			template:       "Value: {{ value }} {% if show %}(visible){% endif %}",
			expectedErrors: 0,
		},
		{
			name:           "unbalanced opening double brace",
			template:       testfixtures.InvalidTemplates.UnbalancedOpen,
			expectedErrors: 1,
		},
		{
			name:           "unbalanced closing double brace",
			template:       testfixtures.InvalidTemplates.UnbalancedClose,
			expectedErrors: 1,
		},
		{
			name:           "unbalanced opening block",
			template:       testfixtures.InvalidTemplates.UnbalancedBlock,
			expectedErrors: 1,
		},
		{
			name:           "unbalanced closing block",
			template:       "endif %}",
			expectedErrors: 1,
		},
		{
			name:           "input inside double braces",
			template:       testfixtures.InvalidTemplates.InputInTemplate,
			expectedErrors: 1,
		},
		{
			name:           "input reference outside template (valid)",
			template:       testfixtures.ValidTemplates.InputRef,
			expectedErrors: 0,
		},
		{
			name:           "multiple errors",
			template:       testfixtures.InvalidTemplates.MultipleErrors,
			expectedErrors: 2, // input in template + unbalanced block
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.validateTemplateString(tt.template, "test.path")

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateTemplatesInValue(t *testing.T) {
	t.Parallel()

	t.Run("validates strings in map", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		value := testfixtures.Map{
			"valid":   testfixtures.ValidTemplates.States,
			"invalid": testfixtures.InvalidTemplates.UnbalancedOpen,
		}

		v.validateTemplatesInValue(value, "root")

		assert.Len(t, v.Errors, 1)
		assert.Contains(t, v.Errors[0], "unbalanced")
	})

	t.Run("validates strings in list", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		value := testfixtures.List{
			"{{ valid }}",
			testfixtures.ValidTemplates.MultiLine,
			testfixtures.InvalidTemplates.UnbalancedOpen,
		}

		v.validateTemplatesInValue(value, "items")

		assert.Len(t, v.Errors, 1)
	})

	t.Run("validates deeply nested structures", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		value := testfixtures.Map{
			"level1": testfixtures.Map{
				"level2": testfixtures.List{
					testfixtures.Map{
						"template": testfixtures.InvalidTemplates.InputInTemplate,
					},
				},
			},
		}

		v.validateTemplatesInValue(value, "root")

		assert.Len(t, v.Errors, 1)
		assert.Contains(t, v.Errors[0], "cannot use !input")
	})

	t.Run("skips non-string values", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		value := testfixtures.Map{
			"number": 42,
			"bool":   true,
			"nil":    nil,
		}

		v.validateTemplatesInValue(value, "root")

		assert.Empty(t, v.Errors)
	})
}

func TestComplexTemplateValidation(t *testing.T) {
	t.Parallel()

	// Test a realistic blueprint structure using fixtures
	v := New("test.yaml")
	v.Data = testfixtures.AutomationBlueprintWithVariables(
		"Test Blueprint",
		testfixtures.Map{},
		testfixtures.Map{
			"temp": testfixtures.ValidTemplates.Float,
			"on":   "{{ true if temp > 20 else false }}",
		},
		[]testfixtures.Map{
			testfixtures.TemplateTrigger(testfixtures.ValidTemplates.IsState),
		},
		[]testfixtures.Map{
			testfixtures.ServiceCallWithData(
				testfixtures.CommonServices.NotifyMobile,
				testfixtures.Map{
					"message": "Temperature is {{ states('sensor.temp') }}",
				},
			),
		},
	)

	v.ValidateTemplates()

	assert.Empty(t, v.Errors, "No errors expected for valid templates")
}
