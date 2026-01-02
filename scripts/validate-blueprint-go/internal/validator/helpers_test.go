package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToFloat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     interface{}
		expected  float64
		expectErr bool
	}{
		{
			name:     "float64",
			value:    3.14,
			expected: 3.14,
		},
		{
			name:     "float32",
			value:    float32(2.5),
			expected: 2.5,
		},
		{
			name:     "int",
			value:    42,
			expected: 42.0,
		},
		{
			name:     "int64",
			value:    int64(100),
			expected: 100.0,
		},
		{
			name:     "string number",
			value:    "3.14159",
			expected: 3.14159,
		},
		{
			name:     "string integer",
			value:    "42",
			expected: 42.0,
		},
		{
			name:      "invalid string",
			value:     "not a number",
			expectErr: true,
		},
		{
			name:      "unsupported type",
			value:     []int{1, 2, 3},
			expectErr: true,
		},
		{
			name:     "negative float",
			value:    -5.5,
			expected: -5.5,
		},
		{
			name:     "zero",
			value:    0,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := ToFloat(tt.value)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.InDelta(t, tt.expected, result, 0.0001)
			}
		})
	}
}

func TestAbs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    float64
		expected float64
	}{
		{name: "positive", value: 5.5, expected: 5.5},
		{name: "negative", value: -5.5, expected: 5.5},
		{name: "zero", value: 0, expected: 0},
		{name: "large negative", value: -1000.5, expected: 1000.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := Abs(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsVariableRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		expected bool
	}{
		{name: "has variable", template: "{{ my_var }}", expected: true},
		{name: "no template", template: "plain text", expected: false},
		{name: "empty template", template: "{{ }}", expected: false},
		{name: "underscore variable", template: "{{ my_long_variable_name }}", expected: true},
		{name: "with function call", template: "{{ states(entity_id) }}", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ContainsVariableRef(tt.template)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsInputRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		expected bool
	}{
		{name: "has input ref", template: "!input my_input", expected: true},
		{name: "no input ref", template: "plain text", expected: false},
		{name: "input in template", template: "{{ states(!input entity) }}", expected: true},
		{name: "partial match", template: "input something", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ContainsInputRef(tt.template)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCollectInputRefs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		value         string
		expectedInput string
	}{
		{
			name:          "valid input reference",
			value:         "!input my_entity",
			expectedInput: "my_entity",
		},
		{
			name:          "not an input reference",
			value:         "regular string",
			expectedInput: "",
		},
		{
			name:          "input reference with spaces",
			value:         "!input   spaced_input  ",
			expectedInput: "spaced_input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.CollectInputRefs(tt.value)

			if tt.expectedInput != "" {
				assert.True(t, v.UsedInputs[tt.expectedInput],
					"Expected input '%s' to be tracked", tt.expectedInput)
			} else {
				assert.Empty(t, v.UsedInputs)
			}
		})
	}
}

func TestCollectInputRefsFromMap(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	m := map[string]interface{}{
		"service": "!input service_name",
		"target": map[string]interface{}{
			"entity_id": "!input target_entity",
		},
		"data": map[string]interface{}{
			"brightness": 255,
			"transition": "!input transition_time",
		},
		"nested_list": []interface{}{
			"!input list_item",
			"regular string",
		},
	}

	v.CollectInputRefsFromMap(m)

	assert.True(t, v.UsedInputs["service_name"])
	assert.True(t, v.UsedInputs["target_entity"])
	assert.True(t, v.UsedInputs["transition_time"])
	assert.True(t, v.UsedInputs["list_item"])
	assert.Len(t, v.UsedInputs, 4)
}

func TestCollectInputRefsFromMapEmpty(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	m := map[string]interface{}{
		"service": "light.turn_on",
		"data": map[string]interface{}{
			"brightness": 255,
		},
	}

	v.CollectInputRefsFromMap(m)

	assert.Empty(t, v.UsedInputs)
}
