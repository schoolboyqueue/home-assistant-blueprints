package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/common"
)

func TestNew(t *testing.T) {
	t.Parallel()

	v := New("/path/to/blueprint.yaml")

	require.NotNil(t, v)
	assert.Equal(t, "/path/to/blueprint.yaml", v.FilePath)
	assert.Empty(t, v.Errors)
	assert.Empty(t, v.Warnings)
	assert.NotNil(t, v.Data)
	assert.NotNil(t, v.DefinedInputs)
	assert.NotNil(t, v.UsedInputs)
	assert.NotNil(t, v.InputDefaults)
	assert.NotNil(t, v.InputSelectors)
	assert.NotNil(t, v.EntityInputs)
	assert.NotNil(t, v.InputDatetimeInputs)
	assert.NotNil(t, v.DefinedVariables)
	assert.NotNil(t, v.JoinVariables)
	assert.NotNil(t, v.NonzeroDefaultVars)
}

func TestAddError(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.AddError("test error message")

	require.Len(t, v.Errors, 1)
	assert.Equal(t, "test error message", v.Errors[0])
}

func TestAddErrorf(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.AddErrorf("error: %s at %d", "invalid value", 42)

	require.Len(t, v.Errors, 1)
	assert.Equal(t, "error: invalid value at 42", v.Errors[0])
}

func TestAddWarning(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.AddWarning("test warning message")

	require.Len(t, v.Warnings, 1)
	assert.Equal(t, "test warning message", v.Warnings[0])
}

func TestAddWarningf(t *testing.T) {
	t.Parallel()

	v := New("test.yaml")
	v.AddWarningf("warning: %s is deprecated", "old_feature")

	require.Len(t, v.Warnings, 1)
	assert.Contains(t, v.Warnings[0], "deprecated")
}

func TestValidateStructure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		data           map[string]interface{}
		expectedErrors int
	}{
		{
			name: "valid structure",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{"name": "test"},
				"trigger":   []interface{}{},
				"action":    []interface{}{},
			},
			expectedErrors: 0,
		},
		{
			name:           "missing all required keys",
			data:           map[string]interface{}{},
			expectedErrors: 3, // blueprint, trigger, action
		},
		{
			name: "missing trigger",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{"name": "test"},
				"action":    []interface{}{},
			},
			expectedErrors: 1,
		},
		{
			name: "variables nested under blueprint (error)",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{
					"name":      "test",
					"variables": map[string]interface{}{},
				},
				"trigger": []interface{}{},
				"action":  []interface{}{},
			},
			expectedErrors: 1,
		},
		{
			name: "variables at root (valid)",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{"name": "test"},
				"trigger":   []interface{}{},
				"action":    []interface{}{},
				"variables": map[string]interface{}{"test": "value"},
			},
			expectedErrors: 0,
		},
		{
			name: "variables not a dictionary",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{"name": "test"},
				"trigger":   []interface{}{},
				"action":    []interface{}{},
				"variables": "not a map",
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.Data = tt.data
			v.ValidateStructure()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateBlueprintSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		data           map[string]interface{}
		expectedErrors int
	}{
		{
			name: "valid blueprint section",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{
					"name":        "Test Blueprint",
					"description": "A test blueprint",
					"domain":      "automation",
					"input":       map[string]interface{}{},
				},
			},
			expectedErrors: 0,
		},
		{
			name:           "no blueprint section",
			data:           map[string]interface{}{},
			expectedErrors: 0, // No errors, just returns
		},
		{
			name: "blueprint is not a map",
			data: map[string]interface{}{
				"blueprint": "not a map",
			},
			expectedErrors: 1,
		},
		{
			name: "missing required blueprint keys",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{
					"name": "Only name",
				},
			},
			expectedErrors: 3, // description, domain, input
		},
		{
			name: "invalid domain",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{
					"name":        "Test",
					"description": "Test",
					"domain":      "invalid_domain",
					"input":       map[string]interface{}{},
				},
			},
			expectedErrors: 1,
		},
		{
			name: "valid script domain",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{
					"name":        "Test",
					"description": "Test",
					"domain":      "script",
					"input":       map[string]interface{}{},
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
			v.ValidateBlueprintSection()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		data           map[string]interface{}
		expectedErrors int
	}{
		{
			name:           "no mode (default single)",
			data:           map[string]interface{}{},
			expectedErrors: 0,
		},
		{
			name:           "valid single mode",
			data:           map[string]interface{}{"mode": "single"},
			expectedErrors: 0,
		},
		{
			name:           "valid restart mode",
			data:           map[string]interface{}{"mode": "restart"},
			expectedErrors: 0,
		},
		{
			name: "valid queued mode with max",
			data: map[string]interface{}{
				"mode": "queued",
				"max":  10,
			},
			expectedErrors: 0,
		},
		{
			name: "valid parallel mode with max",
			data: map[string]interface{}{
				"mode": "parallel",
				"max":  5,
			},
			expectedErrors: 0,
		},
		{
			name:           "invalid mode",
			data:           map[string]interface{}{"mode": "invalid"},
			expectedErrors: 1,
		},
		{
			name:           "mode not a string",
			data:           map[string]interface{}{"mode": 123},
			expectedErrors: 1,
		},
		{
			name: "queued mode with invalid max",
			data: map[string]interface{}{
				"mode": "queued",
				"max":  0,
			},
			expectedErrors: 1,
		},
		{
			name: "parallel mode with negative max",
			data: map[string]interface{}{
				"mode": "parallel",
				"max":  -1,
			},
			expectedErrors: 1,
		},
		{
			name: "queued mode with non-integer max",
			data: map[string]interface{}{
				"mode": "queued",
				"max":  "ten",
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.Data = tt.data
			v.ValidateMode()

			assert.Len(t, v.Errors, tt.expectedErrors, "Errors: %v", v.Errors)
		})
	}
}

func TestValidateVersionSync(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		data             map[string]interface{}
		expectedWarnings int
	}{
		{
			name:             "no blueprint section",
			data:             map[string]interface{}{},
			expectedWarnings: 0,
		},
		{
			name: "no variables section",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{"name": "Test v1.0"},
			},
			expectedWarnings: 0,
		},
		{
			name: "no blueprint_version variable",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{"name": "Test v1.0"},
				"variables": map[string]interface{}{"other": "value"},
			},
			expectedWarnings: 0,
		},
		{
			name: "versions match",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{"name": "Test v1.0"},
				"variables": map[string]interface{}{"blueprint_version": "1.0"},
			},
			expectedWarnings: 0,
		},
		{
			name: "versions mismatch",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{"name": "Test v1.0"},
				"variables": map[string]interface{}{"blueprint_version": "2.0"},
			},
			expectedWarnings: 1,
		},
		{
			name: "no version in name",
			data: map[string]interface{}{
				"blueprint": map[string]interface{}{"name": "Test Blueprint"},
				"variables": map[string]interface{}{"blueprint_version": "1.0"},
			},
			expectedWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := New("test.yaml")
			v.Data = tt.data
			v.ValidateVersionSync()

			assert.Len(t, v.Warnings, tt.expectedWarnings, "Warnings: %v", v.Warnings)
		})
	}
}

func TestMergeValidationResult(t *testing.T) {
	t.Parallel()

	t.Run("merge nil result", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		v.MergeValidationResult(nil)

		assert.Empty(t, v.Errors)
		assert.Empty(t, v.Warnings)
	})

	t.Run("merge result with errors and warnings", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")

		result := common.NewValidationResult()
		result.AddError("path1", "error message")
		result.AddWarning("path2", "warning message")

		v.MergeValidationResult(result)

		require.Len(t, v.Errors, 1)
		require.Len(t, v.Warnings, 1)
		assert.Contains(t, v.Errors[0], "error message")
		assert.Contains(t, v.Warnings[0], "warning message")
	})
}

func TestAddValidationIssue(t *testing.T) {
	t.Parallel()

	t.Run("add error issue", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		issue := common.ValidationIssue{
			Severity: common.SeverityError,
			Path:     "test.path",
			Message:  "error message",
		}

		v.AddValidationIssue(issue)

		require.Len(t, v.Errors, 1)
		assert.Contains(t, v.Errors[0], "error message")
	})

	t.Run("add warning issue", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		issue := common.ValidationIssue{
			Severity: common.SeverityWarning,
			Path:     "test.path",
			Message:  "warning message",
		}

		v.AddValidationIssue(issue)

		require.Len(t, v.Warnings, 1)
		assert.Contains(t, v.Warnings[0], "warning message")
	})
}

func TestAddIssueFromError(t *testing.T) {
	t.Parallel()

	t.Run("add non-empty error", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		v.AddIssueFromError("error message")

		require.Len(t, v.Errors, 1)
		assert.Equal(t, "error message", v.Errors[0])
	})

	t.Run("skip empty error", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		v.AddIssueFromError("")

		assert.Empty(t, v.Errors)
	})
}

func TestAddIssueFromWarning(t *testing.T) {
	t.Parallel()

	t.Run("add non-empty warning", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		v.AddIssueFromWarning("warning message")

		require.Len(t, v.Warnings, 1)
		assert.Equal(t, "warning message", v.Warnings[0])
	})

	t.Run("skip empty warning", func(t *testing.T) {
		t.Parallel()
		v := New("test.yaml")
		v.AddIssueFromWarning("")

		assert.Empty(t, v.Warnings)
	})
}
