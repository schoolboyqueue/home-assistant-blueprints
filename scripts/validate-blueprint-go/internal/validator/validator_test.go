package validator

import (
	"testing"

	"github.com/home-assistant-blueprints/validate-blueprint-go/internal/testfixtures"
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
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("test"),
				"trigger":   testfixtures.List{},
				"action":    testfixtures.List{},
			},
			expectedErrors: 0,
		},
		{
			name:           "missing all required keys",
			data:           testfixtures.Map{},
			expectedErrors: 3, // blueprint, trigger, action
		},
		{
			name: "missing trigger",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("test"),
				"action":    testfixtures.List{},
			},
			expectedErrors: 1,
		},
		{
			name: "variables nested under blueprint (error)",
			data: testfixtures.Map{
				"blueprint": testfixtures.Map{
					"name":      "test",
					"variables": testfixtures.Map{},
				},
				"trigger": testfixtures.List{},
				"action":  testfixtures.List{},
			},
			expectedErrors: 1,
		},
		{
			name: "variables at root (valid)",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("test"),
				"trigger":   testfixtures.List{},
				"action":    testfixtures.List{},
				"variables": testfixtures.Map{"test": "value"},
			},
			expectedErrors: 0,
		},
		{
			name: "variables not a dictionary",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("test"),
				"trigger":   testfixtures.List{},
				"action":    testfixtures.List{},
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
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintSection("Test Blueprint", "A test blueprint", "automation"),
			},
			expectedErrors: 0,
		},
		{
			name:           "no blueprint section",
			data:           testfixtures.Map{},
			expectedErrors: 0, // No errors, just returns
		},
		{
			name: "blueprint is not a map",
			data: testfixtures.Map{
				"blueprint": "not a map",
			},
			expectedErrors: 1,
		},
		{
			name: "missing required blueprint keys",
			data: testfixtures.Map{
				"blueprint": testfixtures.Map{
					"name": "Only name",
				},
			},
			expectedErrors: 3, // description, domain, input
		},
		{
			name: "invalid domain",
			data: testfixtures.Map{
				"blueprint": testfixtures.Map{
					"name":        "Test",
					"description": "Test",
					"domain":      "invalid_domain",
					"input":       testfixtures.Map{},
				},
			},
			expectedErrors: 1,
		},
		{
			name: "valid script domain",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintSection("Test", "Test", "script"),
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
			data:           testfixtures.Map{},
			expectedErrors: 0,
		},
		{
			name:           "valid single mode",
			data:           testfixtures.ModeSection("single"),
			expectedErrors: 0,
		},
		{
			name:           "valid restart mode",
			data:           testfixtures.ModeSection("restart"),
			expectedErrors: 0,
		},
		{
			name:           "valid queued mode with max",
			data:           testfixtures.ModeSectionWithMax("queued", 10),
			expectedErrors: 0,
		},
		{
			name:           "valid parallel mode with max",
			data:           testfixtures.ModeSectionWithMax("parallel", 5),
			expectedErrors: 0,
		},
		{
			name:           "invalid mode",
			data:           testfixtures.ModeSection("invalid"),
			expectedErrors: 1,
		},
		{
			name:           "mode not a string",
			data:           testfixtures.Map{"mode": 123},
			expectedErrors: 1,
		},
		{
			name:           "queued mode with invalid max",
			data:           testfixtures.ModeSectionWithMax("queued", 0),
			expectedErrors: 1,
		},
		{
			name:           "parallel mode with negative max",
			data:           testfixtures.ModeSectionWithMax("parallel", -1),
			expectedErrors: 1,
		},
		{
			name: "queued mode with non-integer max",
			data: testfixtures.Map{
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
			data:             testfixtures.Map{},
			expectedWarnings: 0,
		},
		{
			name: "no variables section",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("Test v1.0"),
			},
			expectedWarnings: 0,
		},
		{
			name: "no blueprint_version variable",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("Test v1.0"),
				"variables": testfixtures.Map{"other": "value"},
			},
			expectedWarnings: 0,
		},
		{
			name: "versions match",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("Test v1.0"),
				"variables": testfixtures.VariablesWithVersion("1.0", testfixtures.Map{}),
			},
			expectedWarnings: 0,
		},
		{
			name: "versions mismatch",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("Test v1.0"),
				"variables": testfixtures.VariablesWithVersion("2.0", testfixtures.Map{}),
			},
			expectedWarnings: 1,
		},
		{
			name: "no version in name",
			data: testfixtures.Map{
				"blueprint": testfixtures.BlueprintWithName("Test Blueprint"),
				"variables": testfixtures.VariablesWithVersion("1.0", testfixtures.Map{}),
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
