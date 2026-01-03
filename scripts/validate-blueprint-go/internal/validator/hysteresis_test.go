package validator

import (
	"testing"

	"github.com/home-assistant-blueprints/testfixtures"
	"github.com/stretchr/testify/assert"
)

func TestValidateHysteresisBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		definedInputs    map[string]bool
		inputDefaults    map[string]interface{}
		expectedErrors   int
		expectedWarnings int
	}{
		{
			name:          "no hysteresis pairs",
			definedInputs: map[string]bool{"some_input": true},
			inputDefaults: map[string]interface{}{"some_input": 50},
		},
		{
			name: "valid hysteresis pair (on > off)",
			definedInputs: map[string]bool{
				"threshold_on":  true,
				"threshold_off": true,
			},
			inputDefaults: map[string]interface{}{
				"threshold_on":  75.0,
				"threshold_off": 25.0,
			},
		},
		{
			name: "inverted hysteresis pair (on < off) - error",
			definedInputs: map[string]bool{
				"threshold_on":  true,
				"threshold_off": true,
			},
			inputDefaults: map[string]interface{}{
				"threshold_on":  25.0,
				"threshold_off": 75.0,
			},
			expectedErrors: 1,
		},
		{
			name: "equal hysteresis values - warning",
			definedInputs: map[string]bool{
				"threshold_on":  true,
				"threshold_off": true,
			},
			inputDefaults: map[string]interface{}{
				"threshold_on":  50.0,
				"threshold_off": 50.0,
			},
			expectedWarnings: 1,
		},
		{
			name: "small hysteresis gap - warning",
			definedInputs: map[string]bool{
				"threshold_on":  true,
				"threshold_off": true,
			},
			inputDefaults: map[string]interface{}{
				"threshold_on":  100.0,
				"threshold_off": 99.0, // Gap is 1% of on value
			},
			expectedWarnings: 1,
		},
		{
			name: "high/low pair valid",
			definedInputs: map[string]bool{
				"temp_high": true,
				"temp_low":  true,
			},
			inputDefaults: map[string]interface{}{
				"temp_high": 30.0,
				"temp_low":  20.0,
			},
		},
		{
			name: "upper/lower pair inverted",
			definedInputs: map[string]bool{
				"limit_upper": true,
				"limit_lower": true,
			},
			inputDefaults: map[string]interface{}{
				"limit_upper": 10.0,
				"limit_lower": 20.0,
			},
			expectedErrors: 1,
		},
		{
			name: "start/stop pair",
			definedInputs: map[string]bool{
				"fan_start": true,
				"fan_stop":  true,
			},
			inputDefaults: map[string]interface{}{
				"fan_start": 75.0,
				"fan_stop":  25.0,
			},
		},
		{
			name: "enable/disable pair",
			definedInputs: map[string]bool{
				"mode_enable":  true,
				"mode_disable": true,
			},
			inputDefaults: map[string]interface{}{
				"mode_enable":  80.0,
				"mode_disable": 20.0,
			},
		},
		{
			name: "only on defined (no pair)",
			definedInputs: map[string]bool{
				"threshold_on": true,
			},
			inputDefaults: map[string]interface{}{
				"threshold_on": 50.0,
			},
		},
		{
			name: "missing default values",
			definedInputs: map[string]bool{
				"threshold_on":  true,
				"threshold_off": true,
			},
			inputDefaults: map[string]interface{}{
				"threshold_on": 50.0,
				// threshold_off has no default
			},
		},
		{
			name: "non-numeric defaults",
			definedInputs: map[string]bool{
				"threshold_on":  true,
				"threshold_off": true,
			},
			inputDefaults: map[string]interface{}{
				"threshold_on":  "high",
				"threshold_off": "low",
			},
		},
		{
			name: "integer defaults",
			definedInputs: map[string]bool{
				"threshold_on":  true,
				"threshold_off": true,
			},
			inputDefaults: map[string]interface{}{
				"threshold_on":  75,
				"threshold_off": 25,
			},
		},
		{
			name: "delta_on/delta_off pair",
			definedInputs: map[string]bool{
				"delta_on":  true,
				"delta_off": true,
			},
			inputDefaults: map[string]interface{}{
				"delta_on":  10.0,
				"delta_off": 5.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.DefinedInputs = tt.definedInputs
			v.InputDefaults = tt.inputDefaults
			v.ValidateHysteresisBoundaries()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)
		})
	}
}

func TestValidateVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		data             map[string]interface{}
		expectedErrors   int
		expectedWarnings int
		expectedVars     []string
	}{
		{
			name:             "no variables section",
			data:             testfixtures.Map{},
			expectedWarnings: 1, // Warning about missing variables section
		},
		{
			name: "variables is not a map",
			data: testfixtures.Map{
				"variables": "not a map",
			},
			expectedErrors: 1,
		},
		{
			name: "valid variables section",
			data: testfixtures.Map{
				"variables": testfixtures.VariablesWithVersion("1.0", testfixtures.Map{
					"my_var": testfixtures.ValidTemplates.States,
				}),
			},
			expectedVars: []string{"my_var", "blueprint_version"},
		},
		{
			name: "missing blueprint_version",
			data: testfixtures.Map{
				"variables": testfixtures.Map{
					"my_var": "value",
				},
			},
			expectedWarnings: 1, // Warning about missing blueprint_version
		},
		{
			name: "variable with join filter",
			data: testfixtures.Map{
				"variables": testfixtures.VariablesWithVersion("1.0", testfixtures.Map{
					"list_var": testfixtures.ValidTemplates.WithFilter,
				}),
			},
			expectedVars: []string{"list_var"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.Data = tt.data
			v.ValidateVariables()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)

			for _, varName := range tt.expectedVars {
				assert.True(t, v.DefinedVariables[varName],
					"Expected variable '%s' to be defined", varName)
			}
		})
	}
}

func TestCheckBareBooleanLiterals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		varName          string
		value            string
		expectedWarnings int
	}{
		{
			name:             "no bare boolean",
			varName:          "test",
			value:            "{{ true }}",
			expectedWarnings: 0,
		},
		{
			name:             "bare true",
			varName:          "my_bool",
			value:            "true",
			expectedWarnings: 1,
		},
		{
			name:             "bare false",
			varName:          "my_bool",
			value:            "false",
			expectedWarnings: 1,
		},
		{
			name:             "true inside template",
			varName:          "test",
			value:            "{{ true if x else false }}",
			expectedWarnings: 0,
		},
		{
			name:    "bare true in multiline with template context",
			varName: "test",
			value: `{% if condition %}
{{ true }}
{% else %}
{{ false }}
{% endif %}`,
			expectedWarnings: 0, // Properly wrapped in template blocks
		},
		{
			name:    "bare booleans on own lines triggers warnings",
			varName: "test",
			value: `{% if condition %}
true
{% else %}
false
{% endif %}`,
			expectedWarnings: 2, // Bare true and false on separate lines
		},
		{
			name:             "normal string",
			varName:          "test",
			value:            "some value",
			expectedWarnings: 0,
		},
		{
			name:             "true as part of string",
			varName:          "test",
			value:            "this is true story",
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.CheckBareBooleanLiterals(tt.varName, tt.value)

			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)
		})
	}
}

func TestCheckUnsafeMathOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		varName          string
		value            string
		expectedWarnings int
	}{
		{
			name:             "no math operations",
			varName:          "test",
			value:            testfixtures.ValidTemplates.States,
			expectedWarnings: 0,
		},
		{
			name:             "log without guard",
			varName:          "log_val",
			value:            "{{ log(value) }}",
			expectedWarnings: 1,
		},
		{
			name:             "log with guard",
			varName:          "log_val",
			value:            "{{ log(value) if value > 0 else 0 }}",
			expectedWarnings: 0,
		},
		{
			name:             "sqrt without guard",
			varName:          "sqrt_val",
			value:            "{{ sqrt(value) }}",
			expectedWarnings: 1,
		},
		{
			name:             "sqrt with max(0, x)",
			varName:          "sqrt_val",
			value:            "{{ sqrt(max(0, value)) }}",
			expectedWarnings: 0,
		},
		{
			name:             "sqrt with abs",
			varName:          "sqrt_val",
			value:            "{{ sqrt(abs(value)) }}",
			expectedWarnings: 0,
		},
		{
			name:             "sqrt with literal positive number",
			varName:          "sqrt_val",
			value:            "{{ sqrt(25) }}",
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.CheckUnsafeMathOperations(tt.varName, tt.value)

			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)
		})
	}
}

func TestCheckPythonStyleMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		varName        string
		value          string
		expectedErrors int
	}{
		{
			name:           "no python methods",
			varName:        "test",
			value:          "{{ [a, b] | min }}",
			expectedErrors: 0,
		},
		{
			name:           "python style min",
			varName:        "test",
			value:          "{{ [a, b].min() }}",
			expectedErrors: 1,
		},
		{
			name:           "python style max",
			varName:        "test",
			value:          "{{ [1, 2, 3].max() }}",
			expectedErrors: 1,
		},
		{
			name:           "python style sum",
			varName:        "test",
			value:          "{{ [1, 2, 3].sum() }}",
			expectedErrors: 1,
		},
		{
			name:           "python style sort",
			varName:        "test",
			value:          "{{ items.sort() }}",
			expectedErrors: 0, // Not a list literal
		},
		{
			name:           "proper jinja filter",
			varName:        "test",
			value:          "{{ items | sort }}",
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.CheckPythonStyleMethods(tt.varName, tt.value)

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestVariableTracking(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.Data = testfixtures.Map{
		"variables": testfixtures.VariablesWithVersion("1.0", testfixtures.Map{
			"var1": testfixtures.ValidTemplates.Float,
			"var2": testfixtures.ValidTemplates.WithFilter,
			"var3": "{{ value | float(10) }}",
		}),
	}

	v.ValidateVariables()

	// Check defined variables tracking
	assert.True(t, v.DefinedVariables["var1"])
	assert.True(t, v.DefinedVariables["var2"])
	assert.True(t, v.DefinedVariables["var3"])
	assert.True(t, v.DefinedVariables["blueprint_version"])

	// Check join variables tracking
	assert.True(t, v.JoinVariables["var2"])
	assert.False(t, v.JoinVariables["var1"])

	// Check nonzero default vars tracking
	assert.True(t, v.NonzeroDefaultVars["var3"])  // Has float(10)
	assert.False(t, v.NonzeroDefaultVars["var1"]) // Has float(0)
}

func TestInputRefCollectionInVariables(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.Data = testfixtures.Map{
		"variables": testfixtures.VariablesWithVersion("1.0", testfixtures.Map{
			"entity":     testfixtures.InputRef("target_entity"),
			"brightness": testfixtures.InputRef("brightness_level"),
		}),
	}

	v.ValidateVariables()

	// Check that input references were collected
	assert.True(t, v.UsedInputs["target_entity"])
	assert.True(t, v.UsedInputs["brightness_level"])
}
