package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{"Unknown", ErrorTypeUnknown, "unknown"},
		{"Syntax", ErrorTypeSyntax, "syntax"},
		{"Schema", ErrorTypeSchema, "schema"},
		{"Validation", ErrorTypeValidation, "validation"},
		{"Parsing", ErrorTypeParsing, "parsing"},
		{"Reference", ErrorTypeReference, "reference"},
		{"Template", ErrorTypeTemplate, "template"},
		{"Input", ErrorTypeInput, "input"},
		{"Trigger", ErrorTypeTrigger, "trigger"},
		{"Condition", ErrorTypeCondition, "condition"},
		{"Action", ErrorTypeAction, "action"},
		{"Documentation", ErrorTypeDocumentation, "documentation"},
		{"Internal", ErrorTypeInternal, "internal"},
		{"Invalid", ErrorType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.errType.String())
		})
	}
}

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "Simple message",
			err:      &Error{Type: ErrorTypeSyntax, Message: "syntax error"},
			expected: "syntax error",
		},
		{
			name:     "With code",
			err:      &Error{Type: ErrorTypeSyntax, Code: "yaml_syntax", Message: "invalid YAML"},
			expected: "[yaml_syntax] invalid YAML",
		},
		{
			name:     "With path",
			err:      &Error{Type: ErrorTypeSchema, Path: "blueprint.input", Message: "missing name"},
			expected: "blueprint.input: missing name",
		},
		{
			name:     "With code and path",
			err:      &Error{Type: ErrorTypeSchema, Code: "missing_name", Path: "blueprint.input", Message: "missing name"},
			expected: "[missing_name] blueprint.input: missing name",
		},
		{
			name:     "With cause",
			err:      &Error{Type: ErrorTypeParsing, Message: "failed to read file", Cause: errors.New("permission denied")},
			expected: "failed to read file: permission denied",
		},
		{
			name:     "With path and cause",
			err:      &Error{Type: ErrorTypeParsing, Path: "blueprint.yaml", Message: "parse error", Cause: errors.New("unexpected token")},
			expected: "blueprint.yaml: parse error: unexpected token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &Error{
		Type:    ErrorTypeParsing,
		Message: "wrapped error",
		Cause:   cause,
	}

	assert.Equal(t, cause, err.Unwrap())
	assert.True(t, errors.Is(err, cause))
}

func TestError_Is(t *testing.T) {
	t.Run("Same type and code", func(t *testing.T) {
		err1 := &Error{Type: ErrorTypeSchema, Code: "missing_blueprint"}
		err2 := &Error{Type: ErrorTypeSchema, Code: "missing_blueprint"}
		assert.True(t, err1.Is(err2))
	})

	t.Run("Same type, different code", func(t *testing.T) {
		err1 := &Error{Type: ErrorTypeSchema, Code: "missing_blueprint"}
		err2 := &Error{Type: ErrorTypeSchema, Code: "missing_name"}
		assert.False(t, err1.Is(err2))
	})

	t.Run("Same type, no code", func(t *testing.T) {
		err1 := &Error{Type: ErrorTypeSchema}
		err2 := &Error{Type: ErrorTypeSchema}
		assert.True(t, err1.Is(err2))
	})

	t.Run("Different type", func(t *testing.T) {
		err1 := &Error{Type: ErrorTypeSchema}
		err2 := &Error{Type: ErrorTypeSyntax}
		assert.False(t, err1.Is(err2))
	})
}

func TestError_WithPath(t *testing.T) {
	original := &Error{
		Type:    ErrorTypeSchema,
		Message: "missing field",
	}

	modified := original.WithPath("blueprint.input.my_input")

	// Original should be unchanged
	assert.Equal(t, "", original.Path)

	// Modified should have path
	assert.Equal(t, "blueprint.input.my_input", modified.Path)
}

func TestError_WithDetails(t *testing.T) {
	original := &Error{
		Type:    ErrorTypeReference,
		Message: "undefined input",
		Details: map[string]any{"key1": "value1"},
	}

	modified := original.WithDetails(map[string]any{"key2": "value2"})

	// Original should be unchanged
	assert.Equal(t, map[string]any{"key1": "value1"}, original.Details)

	// Modified should have both
	assert.Equal(t, map[string]any{"key1": "value1", "key2": "value2"}, modified.Details)
}

func TestNewWithPath(t *testing.T) {
	err := NewWithPath(ErrorTypeInput, "blueprint.input.test", "invalid selector")

	assert.Equal(t, ErrorTypeInput, err.Type)
	assert.Equal(t, "blueprint.input.test", err.Path)
	assert.Equal(t, "invalid selector", err.Message)
}

func TestTypeCheckers(t *testing.T) {
	tests := []struct {
		name    string
		errType ErrorType
		checker func(error) bool
	}{
		{"IsSyntax", ErrorTypeSyntax, IsSyntax},
		{"IsSchema", ErrorTypeSchema, IsSchema},
		{"IsValidation", ErrorTypeValidation, IsValidation},
		{"IsReference", ErrorTypeReference, IsReference},
		{"IsTemplate", ErrorTypeTemplate, IsTemplate},
		{"IsInput", ErrorTypeInput, IsInput},
		{"IsTrigger", ErrorTypeTrigger, IsTrigger},
		{"IsCondition", ErrorTypeCondition, IsCondition},
		{"IsAction", ErrorTypeAction, IsAction},
		{"IsDocumentation", ErrorTypeDocumentation, IsDocumentation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{Type: tt.errType}
			assert.True(t, tt.checker(err))

			otherErr := &Error{Type: ErrorTypeUnknown}
			assert.False(t, tt.checker(otherErr))
		})
	}
}

func TestRegistry(t *testing.T) {
	t.Run("Create from registry", func(t *testing.T) {
		err := Create(CodeMissingBlueprint)
		require.NotNil(t, err)
		assert.Equal(t, CodeMissingBlueprint, err.Code)
		assert.Equal(t, ErrorTypeSchema, err.Type)
	})

	t.Run("CreateWithPath from registry", func(t *testing.T) {
		err := CreateWithPath(CodeUndefinedInput, "trigger[0]")
		require.NotNil(t, err)
		assert.Equal(t, CodeUndefinedInput, err.Code)
		assert.Equal(t, "trigger[0]", err.Path)
	})

	t.Run("Unknown code returns nil", func(t *testing.T) {
		err := Create("nonexistent_code")
		assert.Nil(t, err)
	})
}

func TestConvenienceFactories(t *testing.T) {
	t.Run("ErrYAMLSyntax", func(t *testing.T) {
		cause := errors.New("unexpected EOF")
		err := ErrYAMLSyntax(cause)
		require.NotNil(t, err)
		assert.Equal(t, CodeYAMLSyntax, err.Code)
		assert.Equal(t, cause, err.Cause)
	})

	t.Run("ErrFileNotFound", func(t *testing.T) {
		err := ErrFileNotFound("/path/to/file.yaml")
		require.NotNil(t, err)
		assert.Equal(t, CodeFileNotFound, err.Code)
		assert.Equal(t, "/path/to/file.yaml", err.Path)
	})

	t.Run("ErrMissingField", func(t *testing.T) {
		err := ErrMissingField("blueprint", "name")
		require.NotNil(t, err)
		assert.Equal(t, CodeMissingRequiredField, err.Code)
		assert.Equal(t, "blueprint", err.Path)
		assert.Contains(t, err.Message, "name")
	})

	t.Run("ErrUndefinedInput", func(t *testing.T) {
		err := ErrUndefinedInput("trigger[0]", "my_entity")
		require.NotNil(t, err)
		assert.Equal(t, CodeUndefinedInput, err.Code)
		assert.Equal(t, "trigger[0]", err.Path)
		assert.Equal(t, "my_entity", err.Details["input_name"])
	})

	t.Run("ErrUnusedInput", func(t *testing.T) {
		err := ErrUnusedInput("unused_input")
		require.NotNil(t, err)
		assert.Equal(t, CodeUnusedInput, err.Code)
		assert.Contains(t, err.Message, "unused_input")
	})

	t.Run("ErrVersionMismatch", func(t *testing.T) {
		err := ErrVersionMismatch("1.0.0", "1.0.1")
		require.NotNil(t, err)
		assert.Equal(t, CodeVersionMismatch, err.Code)
		assert.Equal(t, "1.0.0", err.Details["expected"])
		assert.Equal(t, "1.0.1", err.Details["actual"])
	})
}

func TestGetPath(t *testing.T) {
	t.Run("Error with path", func(t *testing.T) {
		err := &Error{Type: ErrorTypeInput, Path: "blueprint.input.test"}
		assert.Equal(t, "blueprint.input.test", GetPath(err))
	})

	t.Run("Error without path", func(t *testing.T) {
		err := &Error{Type: ErrorTypeInput}
		assert.Equal(t, "", GetPath(err))
	})

	t.Run("Standard error", func(t *testing.T) {
		err := errors.New("standard error")
		assert.Equal(t, "", GetPath(err))
	})
}
