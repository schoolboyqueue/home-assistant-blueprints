package validator

import (
	"testing"

	"github.com/home-assistant-blueprints/testfixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidationContext(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()

	require.NotNil(t, ctx)
	assert.Empty(t, ctx.Errors)
	assert.Empty(t, ctx.Warnings)
	assert.Empty(t, ctx.CategorizedErrors)
	assert.Empty(t, ctx.CategorizedWarnings)
	assert.NotNil(t, ctx.Data)
	assert.NotNil(t, ctx.DefinedInputs)
	assert.NotNil(t, ctx.UsedInputs)
	assert.NotNil(t, ctx.InputDefaults)
	assert.NotNil(t, ctx.InputSelectors)
	assert.NotNil(t, ctx.EntityInputs)
	assert.NotNil(t, ctx.InputDatetimeInputs)
	assert.NotNil(t, ctx.DefinedVariables)
	assert.NotNil(t, ctx.JoinVariables)
	assert.NotNil(t, ctx.NonzeroDefaultVars)
}

func TestNewValidationContextWithData(t *testing.T) {
	t.Parallel()

	data := testfixtures.MinimalBlueprint()
	ctx := NewValidationContextWithData(data)

	require.NotNil(t, ctx)
	assert.Equal(t, data, ctx.Data)
	assert.NotNil(t, ctx.DefinedInputs)
}

func TestValidationContextReset(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()
	ctx.Errors = append(ctx.Errors, "error1")
	ctx.Warnings = append(ctx.Warnings, "warning1")
	ctx.DefinedInputs["input1"] = true
	ctx.UsedInputs["input2"] = true
	ctx.Data = testfixtures.MinimalBlueprint()

	ctx.Reset()

	assert.Empty(t, ctx.Errors)
	assert.Empty(t, ctx.Warnings)
	assert.Empty(t, ctx.DefinedInputs)
	assert.Empty(t, ctx.UsedInputs)
	// Data should be preserved
	assert.NotNil(t, ctx.Data)
}

func TestValidationContextHasErrors(t *testing.T) {
	t.Parallel()

	t.Run("no errors", func(t *testing.T) {
		t.Parallel()
		ctx := NewValidationContext()
		assert.False(t, ctx.HasErrors())
	})

	t.Run("has legacy errors", func(t *testing.T) {
		t.Parallel()
		ctx := NewValidationContext()
		ctx.Errors = append(ctx.Errors, "error")
		assert.True(t, ctx.HasErrors())
	})

	t.Run("has categorized errors", func(t *testing.T) {
		t.Parallel()
		ctx := NewValidationContext()
		ctx.CategorizedErrors = append(ctx.CategorizedErrors, CategorizedError{
			Category: CategorySchema,
			Message:  "error",
		})
		assert.True(t, ctx.HasErrors())
	})
}

func TestValidationContextHasWarnings(t *testing.T) {
	t.Parallel()

	t.Run("no warnings", func(t *testing.T) {
		t.Parallel()
		ctx := NewValidationContext()
		assert.False(t, ctx.HasWarnings())
	})

	t.Run("has legacy warnings", func(t *testing.T) {
		t.Parallel()
		ctx := NewValidationContext()
		ctx.Warnings = append(ctx.Warnings, "warning")
		assert.True(t, ctx.HasWarnings())
	})

	t.Run("has categorized warnings", func(t *testing.T) {
		t.Parallel()
		ctx := NewValidationContext()
		ctx.CategorizedWarnings = append(ctx.CategorizedWarnings, CategorizedWarning{
			Category: CategorySchema,
			Message:  "warning",
		})
		assert.True(t, ctx.HasWarnings())
	})
}

func TestValidationContextInputTracking(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()

	// Test MarkInputDefined and IsInputDefined
	assert.False(t, ctx.IsInputDefined("my_input"))
	ctx.MarkInputDefined("my_input")
	assert.True(t, ctx.IsInputDefined("my_input"))

	// Test MarkInputUsed and IsInputUsed
	assert.False(t, ctx.IsInputUsed("my_input"))
	ctx.MarkInputUsed("my_input")
	assert.True(t, ctx.IsInputUsed("my_input"))

	// Test counts
	assert.Equal(t, 1, ctx.DefinedInputCount())
	assert.Equal(t, 1, ctx.UsedInputCount())
}

func TestValidationContextInputDefaults(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()

	// Test SetInputDefault and GetInputDefault
	_, exists := ctx.GetInputDefault("my_input")
	assert.False(t, exists)

	ctx.SetInputDefault("my_input", "default_value")
	val, exists := ctx.GetInputDefault("my_input")
	assert.True(t, exists)
	assert.Equal(t, "default_value", val)
}

func TestValidationContextInputSelectors(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()

	// Test SetInputSelector and GetInputSelector
	_, exists := ctx.GetInputSelector("my_input")
	assert.False(t, exists)

	selector := map[string]interface{}{"entity": map[string]interface{}{}}
	ctx.SetInputSelector("my_input", selector)
	sel, exists := ctx.GetInputSelector("my_input")
	assert.True(t, exists)
	assert.Equal(t, selector, sel)
}

func TestValidationContextEntityInputTracking(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()

	// Test MarkEntityInput and IsEntityInput
	assert.False(t, ctx.IsEntityInput("my_entity"))
	ctx.MarkEntityInput("my_entity")
	assert.True(t, ctx.IsEntityInput("my_entity"))

	// Test MarkInputDatetimeInput and IsInputDatetimeInput
	assert.False(t, ctx.IsInputDatetimeInput("my_datetime"))
	ctx.MarkInputDatetimeInput("my_datetime")
	assert.True(t, ctx.IsInputDatetimeInput("my_datetime"))
}

func TestValidationContextVariableTracking(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()

	// Test MarkVariableDefined and IsVariableDefined
	assert.False(t, ctx.IsVariableDefined("my_var"))
	ctx.MarkVariableDefined("my_var")
	assert.True(t, ctx.IsVariableDefined("my_var"))

	// Test MarkJoinVariable and IsJoinVariable
	assert.False(t, ctx.IsJoinVariable("join_var"))
	ctx.MarkJoinVariable("join_var")
	assert.True(t, ctx.IsJoinVariable("join_var"))

	// Test MarkNonzeroDefaultVar and IsNonzeroDefaultVar
	assert.False(t, ctx.IsNonzeroDefaultVar("nonzero_var"))
	ctx.MarkNonzeroDefaultVar("nonzero_var")
	assert.True(t, ctx.IsNonzeroDefaultVar("nonzero_var"))
}

func TestValidationContextGetUndefinedInputRefs(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()
	ctx.DefinedInputs["defined_input"] = true
	ctx.UsedInputs["defined_input"] = true
	ctx.UsedInputs["undefined_input"] = true

	undefined := ctx.GetUndefinedInputRefs()

	require.Len(t, undefined, 1)
	assert.Contains(t, undefined, "undefined_input")
}

func TestValidationContextGetUnusedInputs(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()
	ctx.DefinedInputs["used_input"] = true
	ctx.DefinedInputs["unused_input"] = true
	ctx.UsedInputs["used_input"] = true

	unused := ctx.GetUnusedInputs()

	require.Len(t, unused, 1)
	assert.Contains(t, unused, "unused_input")
}

func TestValidationContextClone(t *testing.T) {
	t.Parallel()

	original := NewValidationContext()
	original.Errors = append(original.Errors, "error1")
	original.Warnings = append(original.Warnings, "warning1")
	original.CategorizedErrors = append(original.CategorizedErrors, CategorizedError{
		Category: CategorySchema,
		Path:     "path",
		Message:  "error",
	})
	original.DefinedInputs["input1"] = true
	original.UsedInputs["input2"] = true
	original.Data = testfixtures.MinimalBlueprint()

	clone := original.Clone()

	// Verify clone has same values
	assert.Equal(t, original.Errors, clone.Errors)
	assert.Equal(t, original.Warnings, clone.Warnings)
	assert.Equal(t, len(original.CategorizedErrors), len(clone.CategorizedErrors))
	assert.True(t, clone.DefinedInputs["input1"])
	assert.True(t, clone.UsedInputs["input2"])

	// Verify modifications to clone don't affect original
	clone.Errors = append(clone.Errors, "error2")
	clone.DefinedInputs["input3"] = true

	assert.Len(t, original.Errors, 1)
	assert.False(t, original.DefinedInputs["input3"])
}

func TestValidationContextErrorAndWarningCounts(t *testing.T) {
	t.Parallel()

	ctx := NewValidationContext()
	ctx.CategorizedErrors = append(ctx.CategorizedErrors,
		CategorizedError{Category: CategorySchema, Message: "err1"},
		CategorizedError{Category: CategorySchema, Message: "err2"},
	)
	ctx.CategorizedWarnings = append(ctx.CategorizedWarnings,
		CategorizedWarning{Category: CategoryInputs, Message: "warn1"},
	)

	assert.Equal(t, 2, ctx.ErrorCount())
	assert.Equal(t, 1, ctx.WarningCount())
}

func TestNewWithContext(t *testing.T) {
	t.Parallel()

	t.Run("with provided context", func(t *testing.T) {
		t.Parallel()
		ctx := NewValidationContext()
		ctx.DefinedInputs["existing_input"] = true

		v := NewWithContext("/path/to/file.yaml", ctx)

		require.NotNil(t, v)
		assert.Equal(t, "/path/to/file.yaml", v.FilePath)
		assert.True(t, v.DefinedInputs["existing_input"])
		assert.Same(t, ctx, v.Context())
	})

	t.Run("with nil context", func(t *testing.T) {
		t.Parallel()
		v := NewWithContext("/path/to/file.yaml", nil)

		require.NotNil(t, v)
		assert.Equal(t, "/path/to/file.yaml", v.FilePath)
		assert.NotNil(t, v.ValidationContext)
	})
}

func TestValidatorContextMethod(t *testing.T) {
	t.Parallel()

	v := New("/path/to/file.yaml")

	ctx := v.Context()

	require.NotNil(t, ctx)
	assert.Same(t, v.ValidationContext, ctx)
}

func TestValidatorEmbeddedContextAccess(t *testing.T) {
	t.Parallel()

	v := New("/path/to/file.yaml")

	// Test that embedded context fields are accessible directly on validator
	v.MarkInputDefined("my_input")
	assert.True(t, v.IsInputDefined("my_input"))

	v.MarkInputUsed("my_input")
	assert.True(t, v.IsInputUsed("my_input"))

	v.SetInputDefault("my_input", "default")
	val, ok := v.GetInputDefault("my_input")
	assert.True(t, ok)
	assert.Equal(t, "default", val)
}
