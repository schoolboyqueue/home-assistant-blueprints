package validator

import (
	"testing"

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
			data:           map[string]interface{}{},
			expectedErrors: 0,
		},
		{
			name: "valid templates",
			data: map[string]interface{}{
				"template1": "{{ states('light.test') }}",
				"template2": "{% if is_state('light.test', 'on') %}on{% endif %}",
			},
			expectedErrors: 0,
		},
		{
			name: "unbalanced double braces",
			data: map[string]interface{}{
				"template": "{{ states('light.test')",
			},
			expectedErrors: 1,
		},
		{
			name: "unbalanced block braces",
			data: map[string]interface{}{
				"template": "{% if true",
			},
			expectedErrors: 1,
		},
		{
			name: "input inside template block",
			data: map[string]interface{}{
				"template": "{{ !input my_input }}",
			},
			expectedErrors: 1,
		},
		{
			name: "input outside template block (valid)",
			data: map[string]interface{}{
				"value": "!input my_input",
			},
			expectedErrors: 0,
		},
		{
			name: "nested templates all valid",
			data: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "{{ value }}",
					"list": []interface{}{
						"{{ item1 }}",
						"{{ item2 }}",
					},
				},
			},
			expectedErrors: 0,
		},
		{
			name: "nested template with error",
			data: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "{{ unbalanced",
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
			template:       "{{ states('sensor.temp') }}",
			expectedErrors: 0,
		},
		{
			name:           "valid block template",
			template:       "{% if x > 5 %}big{% else %}small{% endif %}",
			expectedErrors: 0,
		},
		{
			name:           "mixed valid template",
			template:       "Value: {{ value }} {% if show %}(visible){% endif %}",
			expectedErrors: 0,
		},
		{
			name:           "unbalanced opening double brace",
			template:       "{{ value",
			expectedErrors: 1,
		},
		{
			name:           "unbalanced closing double brace",
			template:       "value }}",
			expectedErrors: 1,
		},
		{
			name:           "unbalanced opening block",
			template:       "{% if true",
			expectedErrors: 1,
		},
		{
			name:           "unbalanced closing block",
			template:       "endif %}",
			expectedErrors: 1,
		},
		{
			name:           "input inside double braces",
			template:       "{{ !input sensor }}",
			expectedErrors: 1,
		},
		{
			name:           "input reference outside template (valid)",
			template:       "!input sensor_entity",
			expectedErrors: 0,
		},
		{
			name:           "multiple errors",
			template:       "{{ !input x }} {% incomplete",
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
		value := map[string]interface{}{
			"valid":   "{{ states('light.test') }}",
			"invalid": "{{ unclosed",
		}

		v.validateTemplatesInValue(value, "root")

		assert.Len(t, v.Errors, 1)
		assert.Contains(t, v.Errors[0], "unbalanced")
	})

	t.Run("validates strings in list", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		value := []interface{}{
			"{{ valid }}",
			"{% if true %}{% endif %}",
			"{{ unclosed",
		}

		v.validateTemplatesInValue(value, "items")

		assert.Len(t, v.Errors, 1)
	})

	t.Run("validates deeply nested structures", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		value := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": []interface{}{
					map[string]interface{}{
						"template": "{{ !input bad }}",
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
		value := map[string]interface{}{
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

	// Test a realistic blueprint structure
	v := New("test.yaml")
	v.Data = map[string]interface{}{
		"blueprint": map[string]interface{}{
			"name":        "Test Blueprint",
			"description": "A test blueprint",
		},
		"variables": map[string]interface{}{
			"temp": "{{ states('sensor.temperature') | float(0) }}",
			"on":   "{{ true if temp > 20 else false }}",
		},
		"trigger": []interface{}{
			map[string]interface{}{
				"platform":       "template",
				"value_template": "{{ is_state('light.test', 'on') }}",
			},
		},
		"action": []interface{}{
			map[string]interface{}{
				"service": "notify.mobile",
				"data": map[string]interface{}{
					"message": "Temperature is {{ states('sensor.temp') }}",
				},
			},
		},
	}

	v.ValidateTemplates()

	assert.Empty(t, v.Errors, "No errors expected for valid templates")
}
