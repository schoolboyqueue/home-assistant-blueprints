package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/testfixtures"
)

func TestValidateInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		data                map[string]interface{}
		expectedErrors      int
		expectedDefinedKeys []string
	}{
		{
			name:           "no blueprint section",
			data:           testfixtures.Map{},
			expectedErrors: 0,
		},
		{
			name: "no input section",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("Test"),
			},
			expectedErrors: 0,
		},
		{
			name: "input is not a map",
			data: testfixtures.Map{
				"blueprint": testfixtures.Map{
					"name":  "Test",
					"input": "not a map",
				},
			},
			expectedErrors: 1,
		},
		{
			name: "valid simple input",
			data: testfixtures.Map{
				"blueprint": testfixtures.Map{
					"name": "Test",
					"input": testfixtures.Map{
						"my_input": testfixtures.EntityInput("My Input"),
					},
				},
			},
			expectedErrors:      0,
			expectedDefinedKeys: []string{"my_input"},
		},
		{
			name: "input without selector (warning)",
			data: testfixtures.Map{
				"blueprint": testfixtures.Map{
					"name": "Test",
					"input": testfixtures.Map{
						"my_input": testfixtures.InputWithoutSelector("My Input"),
					},
				},
			},
			expectedErrors:      0, // Warnings, not errors
			expectedDefinedKeys: []string{"my_input"},
		},
		{
			name: "nested input group",
			data: testfixtures.Map{
				"blueprint": testfixtures.Map{
					"name": "Test",
					"input": testfixtures.Map{
						"group": testfixtures.InputGroup("Group", testfixtures.Map{
							"nested_input": testfixtures.TextInput("Nested"),
						}),
					},
				},
			},
			expectedErrors:      0,
			expectedDefinedKeys: []string{"nested_input"},
		},
		{
			name: "input definition is not a map",
			data: testfixtures.Map{
				"blueprint": testfixtures.Map{
					"name": "Test",
					"input": testfixtures.Map{
						"my_input": "not a map",
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
			v.Data = tt.data
			v.ValidateInputs()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
			for _, key := range tt.expectedDefinedKeys {
				assert.True(t, v.DefinedInputs[key], "Expected input %s to be defined", key)
			}
		})
	}
}

func TestValidateSingleInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		inputDef         map[string]interface{}
		inputName        string
		expectedErrors   int
		expectedWarnings int
		trackEntityInput bool
		trackDatetime    bool
		trackDefault     bool
		trackSelector    bool
	}{
		{
			name:             "valid entity selector",
			inputDef:         testfixtures.EntityInput("Entity"),
			inputName:        "my_entity",
			expectedErrors:   0,
			expectedWarnings: 0,
			trackEntityInput: true,
			trackSelector:    true,
		},
		{
			name:             "input_datetime entity",
			inputDef:         testfixtures.EntityInputWithDomain("Datetime Entity", "input_datetime"),
			inputName:        "datetime_entity",
			expectedErrors:   0,
			expectedWarnings: 0,
			trackEntityInput: true,
			trackDatetime:    true,
			trackSelector:    true,
		},
		{
			name:             "no selector (warning)",
			inputDef:         testfixtures.InputWithoutSelector("No Selector"),
			inputName:        "no_selector",
			expectedErrors:   0,
			expectedWarnings: 1,
		},
		{
			name: "selector is not a map",
			inputDef: testfixtures.Map{
				"name":     "Invalid Selector",
				"selector": "not a map",
			},
			inputName:        "invalid_selector",
			expectedErrors:   1,
			expectedWarnings: 0,
		},
		{
			name: "unknown selector type (warning)",
			inputDef: testfixtures.Map{
				"name": "Unknown Selector",
				"selector": testfixtures.Map{
					"unknown_type": testfixtures.Map{},
				},
			},
			inputName:        "unknown",
			expectedErrors:   0,
			expectedWarnings: 1,
			trackSelector:    true,
		},
		{
			name:           "with default value",
			inputDef:       testfixtures.TextInputWithDefault("With Default", "default_value"),
			inputName:      "with_default",
			expectedErrors: 0,
			trackDefault:   true,
			trackSelector:  true,
		},
		{
			name: "valid select selector",
			inputDef: testfixtures.SelectInputWithLabelValue("Select Input", []testfixtures.Map{
				testfixtures.SelectOption("Option 1", "option1"),
				testfixtures.SelectOption("Option 2", "option2"),
			}),
			inputName:      "select_input",
			expectedErrors: 0,
			trackSelector:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.validateSingleInput(tt.inputDef, "blueprint.input."+tt.inputName, tt.inputName)

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)

			if tt.trackEntityInput {
				assert.True(t, v.EntityInputs[tt.inputName], "Expected entity input to be tracked")
			}
			if tt.trackDatetime {
				assert.True(t, v.InputDatetimeInputs[tt.inputName], "Expected datetime input to be tracked")
			}
			if tt.trackDefault {
				assert.NotNil(t, v.InputDefaults[tt.inputName], "Expected default to be tracked")
			}
			if tt.trackSelector {
				assert.NotNil(t, v.InputSelectors[tt.inputName], "Expected selector to be tracked")
			}
		})
	}
}

func TestValidateSelectOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		selectConfig   map[string]interface{}
		expectedErrors int
	}{
		{
			name:           "no options",
			selectConfig:   testfixtures.Map{},
			expectedErrors: 0,
		},
		{
			name: "options not a list",
			selectConfig: testfixtures.Map{
				"options": "not a list",
			},
			expectedErrors: 1,
		},
		{
			name: "valid string options",
			selectConfig: testfixtures.Map{
				"options": testfixtures.List{"option1", "option2", "option3"},
			},
			expectedErrors: 0,
		},
		{
			name: "valid label/value options",
			selectConfig: testfixtures.Map{
				"options": testfixtures.List{
					testfixtures.SelectOption("Label 1", "value1"),
					testfixtures.SelectOption("Label 2", "value2"),
				},
			},
			expectedErrors: 0,
		},
		{
			name: "nil option (error)",
			selectConfig: testfixtures.Map{
				"options": testfixtures.List{nil},
			},
			expectedErrors: 1,
		},
		{
			name: "option value is nil",
			selectConfig: testfixtures.Map{
				"options": testfixtures.List{
					testfixtures.Map{
						"label": "Label",
						"value": nil,
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name: "option value is empty string",
			selectConfig: testfixtures.Map{
				"options": testfixtures.List{
					testfixtures.Map{
						"label": "Label",
						"value": "",
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name: "option value is not a string",
			selectConfig: testfixtures.Map{
				"options": testfixtures.List{
					testfixtures.Map{
						"label": "Label",
						"value": 123,
					},
				},
			},
			expectedErrors: 1,
		},
		{
			name: "invalid option type",
			selectConfig: testfixtures.Map{
				"options": testfixtures.List{
					123, // Not string or map
				},
			},
			expectedErrors: 1,
		},
		{
			name: "mixed valid and invalid options",
			selectConfig: testfixtures.Map{
				"options": testfixtures.List{
					"valid_option",
					nil,
					testfixtures.SelectOption("Valid", "valid"),
					testfixtures.Map{
						"label": "Invalid",
						"value": nil,
					},
				},
			},
			expectedErrors: 2, // nil and nil value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.validateSelectOptions(tt.selectConfig, "test.path")

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateSelectOptionsWarnings(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.validateSelectOptions(testfixtures.Map{
		"options": testfixtures.List{""},
	}, "test.path")

	// Empty string option should produce a warning, not an error
	assert.Len(t, v.Errors, 0)
	assert.Len(t, v.Warnings, 1)
	assert.Contains(t, v.Warnings[0], "Empty string option")
}

func TestValidateInputReferences(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		definedInputs  map[string]bool
		usedInputs     map[string]bool
		expectedErrors int
	}{
		{
			name: "all used inputs are defined",
			definedInputs: map[string]bool{
				"input1": true,
				"input2": true,
			},
			usedInputs: map[string]bool{
				"input1": true,
				"input2": true,
			},
			expectedErrors: 0,
		},
		{
			name: "undefined input reference",
			definedInputs: map[string]bool{
				"input1": true,
			},
			usedInputs: map[string]bool{
				"input1":    true,
				"undefined": true,
			},
			expectedErrors: 1,
		},
		{
			name: "multiple undefined inputs",
			definedInputs: map[string]bool{
				"input1": true,
			},
			usedInputs: map[string]bool{
				"undefined1": true,
				"undefined2": true,
			},
			expectedErrors: 2,
		},
		{
			name:           "no used inputs",
			definedInputs:  map[string]bool{"input1": true},
			usedInputs:     map[string]bool{},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.DefinedInputs = tt.definedInputs
			v.UsedInputs = tt.usedInputs
			v.ValidateInputReferences()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateInputDict(t *testing.T) {
	t.Parallel()

	t.Run("nested input groups", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		inputs := testfixtures.Map{
			"group1": testfixtures.InputGroup("Group 1", testfixtures.Map{
				"inner1": testfixtures.TextInput("Inner 1"),
			}),
			"standalone": testfixtures.NumberInput("Standalone", 0, 100),
		}

		v.validateInputDict(inputs, "blueprint.input")

		assert.True(t, v.DefinedInputs["inner1"], "inner1 should be defined")
		assert.True(t, v.DefinedInputs["standalone"], "standalone should be defined")
	})

	t.Run("invalid nested input type", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		inputs := testfixtures.Map{
			"group": testfixtures.Map{
				"name":  "Group",
				"input": "not a map",
			},
		}

		v.validateInputDict(inputs, "blueprint.input")

		require.Len(t, v.Errors, 1)
		assert.Contains(t, v.Errors[0], "must be a dictionary")
	})
}

func TestInputTrackingDuringValidation(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.Data = testfixtures.Map{
		"blueprint": testfixtures.Map{
			"name": "Test",
			"input": testfixtures.Map{
				"entity_input": testfixtures.EntityInputWithDefault("Entity", "light.default"),
				"number_input": testfixtures.NumberInputWithDefault("Number", 0, 100, 50),
			},
		},
	}

	v.ValidateInputs()

	// Check inputs are tracked
	assert.True(t, v.DefinedInputs["entity_input"])
	assert.True(t, v.DefinedInputs["number_input"])
	assert.True(t, v.EntityInputs["entity_input"])
	assert.False(t, v.EntityInputs["number_input"])

	// Check defaults are tracked
	assert.Equal(t, "light.default", v.InputDefaults["entity_input"])
	assert.Equal(t, float64(50), v.InputDefaults["number_input"])

	// Check selectors are tracked
	assert.NotNil(t, v.InputSelectors["entity_input"])
	assert.NotNil(t, v.InputSelectors["number_input"])
}
