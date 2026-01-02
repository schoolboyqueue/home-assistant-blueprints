package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Type Extraction Tests ---

func TestGetMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     interface{}
		path      string
		wantOk    bool
		wantErr   string
		wantValue map[string]interface{}
	}{
		{
			name:      "valid map",
			value:     map[string]interface{}{"key": "value"},
			path:      "test",
			wantOk:    true,
			wantValue: map[string]interface{}{"key": "value"},
		},
		{
			name:    "nil value",
			value:   nil,
			path:    "test.path",
			wantOk:  false,
			wantErr: "test.path: value is nil",
		},
		{
			name:    "wrong type - string",
			value:   "not a map",
			path:    "field",
			wantOk:  false,
			wantErr: "field: must be a dictionary, got string",
		},
		{
			name:    "wrong type - int",
			value:   42,
			path:    "field",
			wantOk:  false,
			wantErr: "field: must be a dictionary, got int",
		},
		{
			name:    "wrong type - slice",
			value:   []interface{}{1, 2, 3},
			path:    "field",
			wantOk:  false,
			wantErr: "field: must be a dictionary, got []interface {}",
		},
		{
			name:      "empty map",
			value:     map[string]interface{}{},
			path:      "empty",
			wantOk:    true,
			wantValue: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok, errMsg := GetMap(tt.value, tt.path)

			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantValue, result)
				assert.Empty(t, errMsg)
			} else {
				assert.Equal(t, tt.wantErr, errMsg)
				assert.Nil(t, result)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     interface{}
		path      string
		wantOk    bool
		wantErr   string
		wantValue string
	}{
		{
			name:      "valid string",
			value:     "hello",
			path:      "test",
			wantOk:    true,
			wantValue: "hello",
		},
		{
			name:      "empty string",
			value:     "",
			path:      "test",
			wantOk:    true,
			wantValue: "",
		},
		{
			name:    "nil value",
			value:   nil,
			path:    "test.path",
			wantOk:  false,
			wantErr: "test.path: value is nil",
		},
		{
			name:    "wrong type - int",
			value:   42,
			path:    "field",
			wantOk:  false,
			wantErr: "field: must be a string, got int",
		},
		{
			name:    "wrong type - map",
			value:   map[string]interface{}{},
			path:    "field",
			wantOk:  false,
			wantErr: "field: must be a string, got map[string]interface {}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok, errMsg := GetString(tt.value, tt.path)

			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantValue, result)
				assert.Empty(t, errMsg)
			} else {
				assert.Equal(t, tt.wantErr, errMsg)
				assert.Empty(t, result)
			}
		})
	}
}

func TestGetList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     interface{}
		path      string
		wantOk    bool
		wantErr   string
		wantValue []interface{}
	}{
		{
			name:      "valid list",
			value:     []interface{}{1, "two", 3.0},
			path:      "test",
			wantOk:    true,
			wantValue: []interface{}{1, "two", 3.0},
		},
		{
			name:      "empty list",
			value:     []interface{}{},
			path:      "test",
			wantOk:    true,
			wantValue: []interface{}{},
		},
		{
			name:    "nil value",
			value:   nil,
			path:    "test.path",
			wantOk:  false,
			wantErr: "test.path: value is nil",
		},
		{
			name:    "wrong type - string",
			value:   "not a list",
			path:    "field",
			wantOk:  false,
			wantErr: "field: must be a list, got string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok, errMsg := GetList(tt.value, tt.path)

			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantValue, result)
				assert.Empty(t, errMsg)
			} else {
				assert.Equal(t, tt.wantErr, errMsg)
				assert.Nil(t, result)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     interface{}
		path      string
		wantOk    bool
		wantErr   string
		wantValue bool
	}{
		{
			name:      "true value",
			value:     true,
			path:      "test",
			wantOk:    true,
			wantValue: true,
		},
		{
			name:      "false value",
			value:     false,
			path:      "test",
			wantOk:    true,
			wantValue: false,
		},
		{
			name:    "nil value",
			value:   nil,
			path:    "test.path",
			wantOk:  false,
			wantErr: "test.path: value is nil",
		},
		{
			name:    "wrong type - string",
			value:   "true",
			path:    "field",
			wantOk:  false,
			wantErr: "field: must be a boolean, got string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok, errMsg := GetBool(tt.value, tt.path)

			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantValue, result)
				assert.Empty(t, errMsg)
			} else {
				assert.Equal(t, tt.wantErr, errMsg)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     interface{}
		path      string
		wantOk    bool
		wantErr   string
		wantValue int
	}{
		{
			name:      "positive int",
			value:     42,
			path:      "test",
			wantOk:    true,
			wantValue: 42,
		},
		{
			name:      "zero",
			value:     0,
			path:      "test",
			wantOk:    true,
			wantValue: 0,
		},
		{
			name:      "negative int",
			value:     -10,
			path:      "test",
			wantOk:    true,
			wantValue: -10,
		},
		{
			name:    "nil value",
			value:   nil,
			path:    "test.path",
			wantOk:  false,
			wantErr: "test.path: value is nil",
		},
		{
			name:    "wrong type - float",
			value:   3.14,
			path:    "field",
			wantOk:  false,
			wantErr: "field: must be an integer, got float64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok, errMsg := GetInt(tt.value, tt.path)

			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantValue, result)
				assert.Empty(t, errMsg)
			} else {
				assert.Equal(t, tt.wantErr, errMsg)
			}
		})
	}
}

// --- Optional Type Extraction Tests ---

func TestTryGetMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		parent    map[string]interface{}
		key       string
		wantOk    bool
		wantValue map[string]interface{}
	}{
		{
			name:      "key exists and is map",
			parent:    map[string]interface{}{"nested": map[string]interface{}{"a": 1}},
			key:       "nested",
			wantOk:    true,
			wantValue: map[string]interface{}{"a": 1},
		},
		{
			name:   "key does not exist",
			parent: map[string]interface{}{"other": "value"},
			key:    "nested",
			wantOk: false,
		},
		{
			name:   "key exists but wrong type",
			parent: map[string]interface{}{"nested": "not a map"},
			key:    "nested",
			wantOk: false,
		},
		{
			name:   "empty parent",
			parent: map[string]interface{}{},
			key:    "any",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok := TryGetMap(tt.parent, tt.key)

			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantValue, result)
			} else {
				assert.Nil(t, result)
			}
		})
	}
}

func TestTryGetString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		parent    map[string]interface{}
		key       string
		wantOk    bool
		wantValue string
	}{
		{
			name:      "key exists and is string",
			parent:    map[string]interface{}{"name": "test"},
			key:       "name",
			wantOk:    true,
			wantValue: "test",
		},
		{
			name:   "key does not exist",
			parent: map[string]interface{}{"other": "value"},
			key:    "name",
			wantOk: false,
		},
		{
			name:   "key exists but wrong type",
			parent: map[string]interface{}{"name": 123},
			key:    "name",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok := TryGetString(tt.parent, tt.key)

			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, result)
		})
	}
}

func TestTryGetList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		parent    map[string]interface{}
		key       string
		wantOk    bool
		wantValue []interface{}
	}{
		{
			name:      "key exists and is list",
			parent:    map[string]interface{}{"items": []interface{}{1, 2, 3}},
			key:       "items",
			wantOk:    true,
			wantValue: []interface{}{1, 2, 3},
		},
		{
			name:   "key does not exist",
			parent: map[string]interface{}{"other": "value"},
			key:    "items",
			wantOk: false,
		},
		{
			name:   "key exists but wrong type",
			parent: map[string]interface{}{"items": "not a list"},
			key:    "items",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok := TryGetList(tt.parent, tt.key)

			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.wantValue, result)
			}
		})
	}
}

func TestTryGetBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		parent    map[string]interface{}
		key       string
		wantOk    bool
		wantValue bool
	}{
		{
			name:      "key exists and is true",
			parent:    map[string]interface{}{"enabled": true},
			key:       "enabled",
			wantOk:    true,
			wantValue: true,
		},
		{
			name:      "key exists and is false",
			parent:    map[string]interface{}{"enabled": false},
			key:       "enabled",
			wantOk:    true,
			wantValue: false,
		},
		{
			name:   "key does not exist",
			parent: map[string]interface{}{"other": "value"},
			key:    "enabled",
			wantOk: false,
		},
		{
			name:   "key exists but wrong type",
			parent: map[string]interface{}{"enabled": "true"},
			key:    "enabled",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok := TryGetBool(tt.parent, tt.key)

			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, result)
		})
	}
}

func TestTryGetInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		parent    map[string]interface{}
		key       string
		wantOk    bool
		wantValue int
	}{
		{
			name:      "key exists and is int",
			parent:    map[string]interface{}{"count": 42},
			key:       "count",
			wantOk:    true,
			wantValue: 42,
		},
		{
			name:   "key does not exist",
			parent: map[string]interface{}{"other": "value"},
			key:    "count",
			wantOk: false,
		},
		{
			name:   "key exists but wrong type",
			parent: map[string]interface{}{"count": "42"},
			key:    "count",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, ok := TryGetInt(tt.parent, tt.key)

			assert.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, result)
		})
	}
}

// --- Path Building Tests ---

func TestJoinPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		parent   string
		child    string
		expected string
	}{
		{
			name:     "non-empty parent",
			parent:   "parent",
			child:    "child",
			expected: "parent.child",
		},
		{
			name:     "empty parent",
			parent:   "",
			child:    "child",
			expected: "child",
		},
		{
			name:     "nested path",
			parent:   "a.b.c",
			child:    "d",
			expected: "a.b.c.d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := JoinPath(tt.parent, tt.child)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIndexPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		parent   string
		index    int
		expected string
	}{
		{
			name:     "simple index",
			parent:   "items",
			index:    0,
			expected: "items[0]",
		},
		{
			name:     "large index",
			parent:   "items",
			index:    42,
			expected: "items[42]",
		},
		{
			name:     "nested path",
			parent:   "parent.items",
			index:    5,
			expected: "parent.items[5]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := IndexPath(tt.parent, tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKeyPath(t *testing.T) {
	t.Parallel()

	result := KeyPath("parent", "child")
	assert.Equal(t, "parent.child", result)
}

// --- Validation Issue Tests ---

func TestValidationIssue_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		issue    ValidationIssue
		expected string
	}{
		{
			name: "issue with path",
			issue: ValidationIssue{
				Severity: SeverityError,
				Path:     "blueprint.input",
				Message:  "missing required key",
			},
			expected: "blueprint.input: missing required key",
		},
		{
			name: "issue without path",
			issue: ValidationIssue{
				Severity: SeverityWarning,
				Path:     "",
				Message:  "general warning",
			},
			expected: "general warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.issue.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidationIssue_IsError(t *testing.T) {
	t.Parallel()

	errorIssue := ValidationIssue{Severity: SeverityError}
	warningIssue := ValidationIssue{Severity: SeverityWarning}

	assert.True(t, errorIssue.IsError())
	assert.False(t, warningIssue.IsError())
}

func TestValidationIssue_IsWarning(t *testing.T) {
	t.Parallel()

	errorIssue := ValidationIssue{Severity: SeverityError}
	warningIssue := ValidationIssue{Severity: SeverityWarning}

	assert.False(t, errorIssue.IsWarning())
	assert.True(t, warningIssue.IsWarning())
}

// --- ValidationResult Tests ---

func TestNewValidationResult(t *testing.T) {
	t.Parallel()

	result := NewValidationResult()
	require.NotNil(t, result)
	assert.Empty(t, result.Issues)
}

func TestValidationResult_AddError(t *testing.T) {
	t.Parallel()

	result := NewValidationResult()
	result.AddError("path.to.field", "error message")

	require.Len(t, result.Issues, 1)
	assert.Equal(t, SeverityError, result.Issues[0].Severity)
	assert.Equal(t, "path.to.field", result.Issues[0].Path)
	assert.Equal(t, "error message", result.Issues[0].Message)
}

func TestValidationResult_AddErrorf(t *testing.T) {
	t.Parallel()

	result := NewValidationResult()
	result.AddErrorf("field", "value %d is invalid", 42)

	require.Len(t, result.Issues, 1)
	assert.Equal(t, "value 42 is invalid", result.Issues[0].Message)
}

func TestValidationResult_AddWarning(t *testing.T) {
	t.Parallel()

	result := NewValidationResult()
	result.AddWarning("path", "warning message")

	require.Len(t, result.Issues, 1)
	assert.Equal(t, SeverityWarning, result.Issues[0].Severity)
}

func TestValidationResult_AddWarningf(t *testing.T) {
	t.Parallel()

	result := NewValidationResult()
	result.AddWarningf("field", "consider using %s", "better value")

	require.Len(t, result.Issues, 1)
	assert.Contains(t, result.Issues[0].Message, "consider using better value")
}

func TestValidationResult_Merge(t *testing.T) {
	t.Parallel()

	t.Run("merge non-nil result", func(t *testing.T) {
		t.Parallel()
		result1 := NewValidationResult()
		result1.AddError("path1", "error1")

		result2 := NewValidationResult()
		result2.AddWarning("path2", "warning1")
		result2.AddError("path3", "error2")

		result1.Merge(result2)

		assert.Len(t, result1.Issues, 3)
	})

	t.Run("merge nil result", func(t *testing.T) {
		t.Parallel()
		result := NewValidationResult()
		result.AddError("path", "error")

		result.Merge(nil)

		assert.Len(t, result.Issues, 1)
	})
}

func TestValidationResult_HasErrors(t *testing.T) {
	t.Parallel()

	t.Run("no issues", func(t *testing.T) {
		t.Parallel()
		result := NewValidationResult()
		assert.False(t, result.HasErrors())
	})

	t.Run("only warnings", func(t *testing.T) {
		t.Parallel()
		result := NewValidationResult()
		result.AddWarning("path", "warning")
		assert.False(t, result.HasErrors())
	})

	t.Run("has errors", func(t *testing.T) {
		t.Parallel()
		result := NewValidationResult()
		result.AddError("path", "error")
		assert.True(t, result.HasErrors())
	})
}

func TestValidationResult_HasWarnings(t *testing.T) {
	t.Parallel()

	t.Run("no issues", func(t *testing.T) {
		t.Parallel()
		result := NewValidationResult()
		assert.False(t, result.HasWarnings())
	})

	t.Run("only errors", func(t *testing.T) {
		t.Parallel()
		result := NewValidationResult()
		result.AddError("path", "error")
		assert.False(t, result.HasWarnings())
	})

	t.Run("has warnings", func(t *testing.T) {
		t.Parallel()
		result := NewValidationResult()
		result.AddWarning("path", "warning")
		assert.True(t, result.HasWarnings())
	})
}

func TestValidationResult_Errors(t *testing.T) {
	t.Parallel()

	result := NewValidationResult()
	result.AddError("path1", "error1")
	result.AddWarning("path2", "warning1")
	result.AddError("path3", "error2")

	errors := result.Errors()
	require.Len(t, errors, 2)
	assert.Equal(t, "error1", errors[0].Message)
	assert.Equal(t, "error2", errors[1].Message)
}

func TestValidationResult_Warnings(t *testing.T) {
	t.Parallel()

	result := NewValidationResult()
	result.AddError("path1", "error1")
	result.AddWarning("path2", "warning1")
	result.AddWarning("path3", "warning2")

	warnings := result.Warnings()
	require.Len(t, warnings, 2)
	assert.Equal(t, "warning1", warnings[0].Message)
	assert.Equal(t, "warning2", warnings[1].Message)
}

func TestValidationResult_ErrorStrings(t *testing.T) {
	t.Parallel()

	result := NewValidationResult()
	result.AddError("path1", "error1")
	result.AddWarning("path2", "warning1")
	result.AddError("path3", "error2")

	errorStrings := result.ErrorStrings()
	require.Len(t, errorStrings, 2)
	assert.Contains(t, errorStrings[0], "path1")
	assert.Contains(t, errorStrings[1], "path3")
}

func TestValidationResult_WarningStrings(t *testing.T) {
	t.Parallel()

	result := NewValidationResult()
	result.AddWarning("path1", "warning1")
	result.AddError("path2", "error1")

	warningStrings := result.WarningStrings()
	require.Len(t, warningStrings, 1)
	assert.Contains(t, warningStrings[0], "warning1")
}

// --- Common Validation Functions Tests ---

func TestValidateRequired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		path     string
		wantErr  bool
		expected string
	}{
		{
			name:    "key exists",
			m:       map[string]interface{}{"required_key": "value"},
			key:     "required_key",
			path:    "test",
			wantErr: false,
		},
		{
			name:     "key missing",
			m:        map[string]interface{}{"other": "value"},
			key:      "required_key",
			path:     "test",
			wantErr:  true,
			expected: "test: missing required key 'required_key'",
		},
		{
			name:    "nil value is present",
			m:       map[string]interface{}{"key": nil},
			key:     "key",
			path:    "test",
			wantErr: false, // Key exists, even if value is nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateRequired(tt.m, tt.key, tt.path)

			if tt.wantErr {
				assert.Equal(t, tt.expected, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidateRequiredKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		m            map[string]interface{}
		keys         []string
		path         string
		expectedLen  int
		expectedMsgs []string
	}{
		{
			name:        "all keys present",
			m:           map[string]interface{}{"a": 1, "b": 2, "c": 3},
			keys:        []string{"a", "b"},
			path:        "test",
			expectedLen: 0,
		},
		{
			name:        "one key missing",
			m:           map[string]interface{}{"a": 1},
			keys:        []string{"a", "b"},
			path:        "test",
			expectedLen: 1,
		},
		{
			name:        "multiple keys missing",
			m:           map[string]interface{}{},
			keys:        []string{"a", "b", "c"},
			path:        "test",
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateRequiredKeys(tt.m, tt.keys, tt.path)
			assert.Len(t, result, tt.expectedLen)
		})
	}
}

func TestValidateEnumValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     string
		allowed   []string
		path      string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid value",
			value:     "single",
			allowed:   []string{"single", "restart", "queued"},
			path:      "mode",
			fieldName: "mode",
			wantErr:   false,
		},
		{
			name:      "invalid value",
			value:     "invalid",
			allowed:   []string{"single", "restart", "queued"},
			path:      "mode",
			fieldName: "mode",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateEnumValue(tt.value, tt.allowed, tt.path, tt.fieldName)

			if tt.wantErr {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "invalid")
				assert.Contains(t, result, tt.value)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidateEnumMap(t *testing.T) {
	t.Parallel()

	allowed := map[string]bool{"entity": true, "device": true, "area": true}

	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "valid value",
			value:     "entity",
			fieldName: "selector type",
			wantErr:   false,
		},
		{
			name:      "invalid value",
			value:     "unknown",
			fieldName: "selector type",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateEnumMap(tt.value, allowed, "path", tt.fieldName)

			if tt.wantErr {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "unknown")
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidatePositiveInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{name: "positive", value: 5, wantErr: false},
		{name: "one", value: 1, wantErr: false},
		{name: "zero", value: 0, wantErr: true},
		{name: "negative", value: -1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidatePositiveInt(tt.value, "path", "count")

			if tt.wantErr {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidateNotNil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{name: "non-nil string", value: "test", wantErr: false},
		{name: "non-nil int", value: 0, wantErr: false},
		{name: "nil", value: nil, wantErr: true},
		{name: "empty string", value: "", wantErr: false}, // Empty string is not nil
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateNotNil(tt.value, "path", "field")

			if tt.wantErr {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "cannot be None")
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidateNotEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "non-empty", value: "test", wantErr: false},
		{name: "empty", value: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateNotEmpty(tt.value, "path", "field")

			if tt.wantErr {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "cannot be empty")
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

// --- Domain and Selector Validation Tests ---

func TestValidateSelector(t *testing.T) {
	t.Parallel()

	validTypes := map[string]bool{"entity": true, "device": true, "area": true}

	tests := []struct {
		name         string
		selectorType string
		wantWarning  bool
	}{
		{name: "valid entity", selectorType: "entity", wantWarning: false},
		{name: "valid device", selectorType: "device", wantWarning: false},
		{name: "unknown type", selectorType: "unknown", wantWarning: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateSelector(tt.selectorType, validTypes, "path")

			if tt.wantWarning {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "unknown selector type")
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidateEntityDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		entityDomain    string
		expectedDomains []string
		wantErr         bool
	}{
		{
			name:            "valid domain",
			entityDomain:    "light",
			expectedDomains: []string{"light", "switch"},
			wantErr:         false,
		},
		{
			name:            "unexpected domain",
			entityDomain:    "sensor",
			expectedDomains: []string{"light", "switch"},
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateEntityDomain(tt.entityDomain, tt.expectedDomains, "path")

			if tt.wantErr {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "unexpected entity domain")
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidateServiceFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service string
		wantErr bool
	}{
		{name: "valid domain.service", service: "light.turn_on", wantErr: false},
		{name: "valid input reference", service: "!input service_name", wantErr: false},
		{name: "valid template", service: "{{ my_service }}", wantErr: false},
		{name: "valid jinja block", service: "{% if x %}light.on{% endif %}", wantErr: false},
		{name: "invalid format", service: "light_turn_on", wantErr: true},
		{name: "single word", service: "service", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateServiceFormat(tt.service, "path")

			if tt.wantErr {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "domain.service")
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

// --- Template Validation Tests ---

func TestContainsTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "has double braces", input: "{{ variable }}", expected: true},
		{name: "has block", input: "{% if x %}yes{% endif %}", expected: true},
		{name: "no template", input: "plain text", expected: false},
		{name: "partial double brace", input: "{ not template }", expected: false},
		{name: "mixed content", input: "text {{ var }} more", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ContainsTemplate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsInputRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "has input ref", input: "!input my_input", expected: true},
		{name: "no input ref", input: "plain text", expected: false},
		{name: "input in template", input: "{{ !input x }}", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ContainsInputRef(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsVariableRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "has variable", input: "{{ my_var }}", expected: true},
		{name: "no template", input: "plain text", expected: false},
		{name: "only numbers", input: "{{ 123 }}", expected: false},
		{name: "complex var", input: "{{ some_variable_name }}", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ContainsVariableRef(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateBalancedDelimiters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		template   string
		wantErrors int
	}{
		{name: "balanced double braces", template: "{{ var }}", wantErrors: 0},
		{name: "balanced block", template: "{% if x %}{% endif %}", wantErrors: 0},
		{name: "unbalanced double braces", template: "{{ var", wantErrors: 1},
		{name: "unbalanced block opening", template: "{% if x", wantErrors: 1},
		{name: "both unbalanced", template: "{{ x {% y", wantErrors: 2},
		{name: "nested balanced", template: "{{ x }} {{ y }}", wantErrors: 0},
		{name: "single block tag is balanced", template: "{% if x %}", wantErrors: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateBalancedDelimiters(tt.template, "path")
			assert.Len(t, result, tt.wantErrors)
		})
	}
}

func TestValidateNoInputInTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "input outside template", input: "!input name", wantErr: false},
		{name: "input inside template", input: "{{ !input name }}", wantErr: true},
		{name: "no input", input: "{{ variable }}", wantErr: false},
		{name: "plain text", input: "plain text", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateNoInputInTemplate(tt.input, "path")

			if tt.wantErr {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "cannot use !input")
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestValidateNoTemplateInField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "no template", value: "light.living_room", wantErr: false},
		{name: "has template", value: "{{ entity }}", wantErr: true},
		{name: "has block", value: "{% if x %}{% endif %}", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidateNoTemplateInField(tt.value, "path", "entity_id")

			if tt.wantErr {
				assert.NotEmpty(t, result)
				assert.Contains(t, result, "cannot contain templates")
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

// --- Traversal Tests ---

func TestTraverseValue(t *testing.T) {
	t.Parallel()

	t.Run("traverses map", func(t *testing.T) {
		t.Parallel()
		value := map[string]interface{}{
			"key1": "value1",
			"key2": map[string]interface{}{
				"nested": "value2",
			},
		}

		var visited []string
		TraverseValue(value, "root", func(_ interface{}, path string) bool {
			visited = append(visited, path)
			return true
		})

		assert.Contains(t, visited, "root")
		assert.Contains(t, visited, "root.key1")
		assert.Contains(t, visited, "root.key2")
		assert.Contains(t, visited, "root.key2.nested")
	})

	t.Run("traverses list", func(t *testing.T) {
		t.Parallel()
		value := []interface{}{"a", "b", "c"}

		var visited []string
		TraverseValue(value, "items", func(_ interface{}, path string) bool {
			visited = append(visited, path)
			return true
		})

		assert.Contains(t, visited, "items")
		assert.Contains(t, visited, "items[0]")
		assert.Contains(t, visited, "items[1]")
		assert.Contains(t, visited, "items[2]")
	})

	t.Run("stops traversal when visitor returns false", func(t *testing.T) {
		t.Parallel()
		value := map[string]interface{}{
			"key1": map[string]interface{}{
				"nested": "value",
			},
		}

		var visited []string
		TraverseValue(value, "root", func(_ interface{}, path string) bool {
			visited = append(visited, path)
			return path != "root.key1" // Stop at key1
		})

		assert.Contains(t, visited, "root")
		assert.Contains(t, visited, "root.key1")
		assert.NotContains(t, visited, "root.key1.nested")
	})
}

func TestCollectStrings(t *testing.T) {
	t.Parallel()

	value := map[string]interface{}{
		"string1": "value1",
		"number":  42,
		"nested": map[string]interface{}{
			"string2": "value2",
		},
		"list": []interface{}{"item1", "item2"},
	}

	result := CollectStrings(value)

	assert.Contains(t, result, "value1")
	assert.Contains(t, result, "value2")
	assert.Contains(t, result, "item1")
	assert.Contains(t, result, "item2")
	assert.Len(t, result, 4)
}

func TestTraverseMaps(t *testing.T) {
	t.Parallel()

	value := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"key": "value",
			},
		},
		"list": []interface{}{
			map[string]interface{}{"item": "value"},
		},
	}

	var visitedPaths []string
	TraverseMaps(value, "root", func(_ map[string]interface{}, path string) {
		visitedPaths = append(visitedPaths, path)
	})

	assert.Contains(t, visitedPaths, "root")
	assert.Contains(t, visitedPaths, "root.level1")
	assert.Contains(t, visitedPaths, "root.level1.level2")
	assert.Contains(t, visitedPaths, "root.list[0]")
}

// --- Input Reference Tests ---

func TestExtractInputRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "valid input ref", value: "!input my_input", expected: "my_input"},
		{name: "input ref with spaces", value: "!input  my_input  ", expected: "my_input"},
		{name: "not an input ref", value: "plain text", expected: ""},
		{name: "partial match", value: "some !input", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ExtractInputRef(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCollectInputRefsFromValue(t *testing.T) {
	t.Parallel()

	value := map[string]interface{}{
		"service": "!input service_name",
		"entity":  "!input entity_id",
		"nested": map[string]interface{}{
			"target": "!input target_entity",
		},
		"list": []interface{}{
			"!input list_item",
		},
		"plain": "no input here",
	}

	result := CollectInputRefsFromValue(value)

	assert.True(t, result["service_name"])
	assert.True(t, result["entity_id"])
	assert.True(t, result["target_entity"])
	assert.True(t, result["list_item"])
	assert.Len(t, result, 4)
}

// --- List/Map Validation Tests ---

func TestValidateListItems(t *testing.T) {
	t.Parallel()

	list := []interface{}{"valid", "", "also valid"}

	result := ValidateListItems(list, "items", func(item interface{}, _ int, itemPath string) *ValidationResult {
		r := NewValidationResult()
		if s, ok := item.(string); ok && s == "" {
			r.AddError(itemPath, "item cannot be empty")
		}
		return r
	})

	assert.True(t, result.HasErrors())
	require.Len(t, result.Errors(), 1)
	assert.Contains(t, result.Errors()[0].Path, "items[1]")
}

func TestValidateMapEntries(t *testing.T) {
	t.Parallel()

	m := map[string]interface{}{
		"valid":   "value",
		"invalid": nil,
	}

	result := ValidateMapEntries(m, "config", func(_ string, value interface{}, entryPath string) *ValidationResult {
		r := NewValidationResult()
		if value == nil {
			r.AddError(entryPath, "value cannot be nil")
		}
		return r
	})

	assert.True(t, result.HasErrors())
	require.Len(t, result.Errors(), 1)
	assert.Contains(t, result.Errors()[0].Path, "config.invalid")
}

// --- Conditional Validation Tests ---

func TestValidateIf(t *testing.T) {
	t.Parallel()

	t.Run("condition true", func(t *testing.T) {
		t.Parallel()
		result := ValidateIf(true, func() string {
			return "error message"
		})
		assert.Equal(t, "error message", result)
	})

	t.Run("condition false", func(t *testing.T) {
		t.Parallel()
		result := ValidateIf(false, func() string {
			return "error message"
		})
		assert.Empty(t, result)
	})
}

func TestValidateIfPresent(t *testing.T) {
	t.Parallel()

	t.Run("key present and valid", func(t *testing.T) {
		t.Parallel()
		m := map[string]interface{}{"key": "value"}
		result := ValidateIfPresent(m, "key", func(_ interface{}) string {
			return ""
		})
		assert.Empty(t, result)
	})

	t.Run("key present and invalid", func(t *testing.T) {
		t.Parallel()
		m := map[string]interface{}{"key": ""}
		result := ValidateIfPresent(m, "key", func(value interface{}) string {
			if s, ok := value.(string); ok && s == "" {
				return "value cannot be empty"
			}
			return ""
		})
		assert.Equal(t, "value cannot be empty", result)
	})

	t.Run("key absent", func(t *testing.T) {
		t.Parallel()
		m := map[string]interface{}{"other": "value"}
		result := ValidateIfPresent(m, "key", func(_ interface{}) string {
			return "this should not be called"
		})
		assert.Empty(t, result)
	})
}
